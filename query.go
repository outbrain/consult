package main

import (
	"errors"
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
		short := make([]string, len(results))
		long := make([]string, len(results)+2)
		long[0] = misc.StructHeaderLine(api.CatalogService{})
		long[1] = ""

		for i, res := range results {
			short[i] = res.Node
			long[i+2] = misc.StructToString(res)
		}

		q.Output(results, long, short)
		return nil
	}
}

func (q *QueryCommand) queryServicesGeneric() (services []*api.CatalogService, err_ error) {
	defer func() {
		if r := recover(); r != nil {
			err_ = r.(error)
		}
	}()

	mergeFunc := intersectionMerge
	if q.tagsMerge == "any" {
		mergeFunc = unionMerge
	}

	if clients, err := q.GetConsulClients(); err != nil {
		return nil, err
	} else {
		results := make([]*api.CatalogService, 0)

		for _, client := range clients {
			results = unionMerge(results, q.queryMulti(client, mergeFunc))
		}

		if len(results) == 0 {
			return nil, errors.New("No results from Consul query")
		}
		return results, nil
	}
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
