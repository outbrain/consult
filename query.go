package main

import (
	"github.com/hashicorp/consul/api"
	"github.com/outbrain/consult/misc"
	"github.com/wushilin/stream"
	"gopkg.in/alecthomas/kingpin.v2"
)

type QueryCommand struct {
	Command
	misc.IQuery
	service   string
	tags      []string
	tagsMerge string
}

type queryCommand struct {
	QueryCommand
}

func queryRegisterCli(app *kingpin.Application, opts *appOpts) {
	q := &queryCommand{}
	q.IQuery = q
	q.opts = opts
	queryCmd := app.Command("query", "Query Consul catalog").Action(q.run)
	queryCmd.Flag("tag", "Consul tag").Short('t').StringsVar(&q.tags)
	queryCmd.Flag("service", "Consul service").Required().Short('s').StringVar(&q.service)
	q.registerCli(queryCmd)
}

func (q *queryCommand) run(c *kingpin.ParseContext) error {
	if results, err := q.queryServicesGeneric(); err != nil {
		return err
	} else {
		short := make([]string, 0)
		long := make([]string, 2)
		long[0] = misc.StructHeaderLine(api.CatalogService{})
		long[1] = ""

		for dc, res := range results {
			for _, service := range res {
				short = append(short, dc+"\t"+service.Node)
				long = append(long, dc+"\t"+misc.StructToString(service))
			}
		}

		q.Output(results, long, short)
		return nil
	}
}

func (q *QueryCommand) registerCli(cmd *kingpin.CmdClause) {
	cmd.Flag("tag", "Consul tag").Short('t').StringsVar(&q.tags)
	cmd.Flag("service", "Consul service").Required().Short('s').StringVar(&q.service)
	cmd.Flag("tags-mode", "Find nodes with *all* or *any* of the tags").Short('m').Default("all").EnumVar(&q.tagsMerge, "all", "any")
}

func (q *QueryCommand) queryServicesGeneric() (services map[string][]*api.CatalogService, err_ error) {
	defer func() {
		if r := recover(); r != nil {
			err_ = r.(error)
		}
	}()

	mergeFunc := intersectionMerge
	if q.tagsMerge == "any" {
		mergeFunc = unionMerge
	}

	results_by_dc, err := q.QueryWithClients(func(client *api.Client) interface{} {
		return q.queryMulti(client, mergeFunc)
	})
	if err != nil {
		return nil, err
	}
	typed_results := make(map[string][]*api.CatalogService)
	for k, v := range results_by_dc {
		typed_results[k] = v.([]*api.CatalogService)
	}
	return typed_results, nil
}

func (q *QueryCommand) queryMulti(
	client *api.Client,
	mergeFunc misc.CatalogServicesMerger,
) []*api.CatalogService {
	// handle the case of no tags
	if len(q.tags) == 0 {
		return q.IQuery.Query(client, "")
	}

	res, ok := (&basePStream{stream.FromArray(q.tags)}).PMap(
		func(tag interface{}) interface{} {
			return q.IQuery.Query(client, tag.(string)) // use q.IQuery.Query so we can override for testing
		}).Reduce(mergeFunc).Value()

	if ok {
		return res.([]*api.CatalogService)
	} else {
		// perhaps blow up instead?
		return []*api.CatalogService{}
	}
}

func (q *QueryCommand) Query(client *api.Client, tag string) []*api.CatalogService {
	services, _, err := client.Catalog().Service(q.service, tag, &api.QueryOptions{AllowStale: true, RequireConsistent: false})
	if err != nil {
		panic(err)
		return nil
	}

	return services
}

type CatalogServiceList []*api.CatalogService

func (c CatalogServiceList) Contains(elmnt *api.CatalogService) bool {
	for _, x := range c {
		if x.Node == elmnt.Node {
			return true
		}
	}
	return false
}

// i really don't care about efficiency. merge every two items with these functions
func unionMerge(a []*api.CatalogService, b []*api.CatalogService) []*api.CatalogService {
	res := make(CatalogServiceList, 0, len(a)+len(b))
	for _, x := range a {
		res = append(res, x)
	}
	for _, x := range b {
		if !res.Contains(x) {
			res = append(res, x)
		}
	}
	return res
}

func intersectionMerge(a []*api.CatalogService, b []*api.CatalogService) []*api.CatalogService {
	res := make([]*api.CatalogService, 0)
	for _, x := range a {
		if CatalogServiceList(b).Contains(x) {
			res = append(res, x)
		}
	}
	return res
}
