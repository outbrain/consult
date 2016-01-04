package main

import (
	"github.com/hashicorp/consul/api"
	"github.com/wushilin/stream"
	"gopkg.in/alecthomas/kingpin.v2"
)

type catalogServicesMerger func(a []*api.CatalogService, b []*api.CatalogService) (mergedList []*api.CatalogService)

func queryMulti(
	config *api.Config,
	service string,
	tags []string,
	mergeFunc catalogServicesMerger,
) []*api.CatalogService {
	// handle the case of no tags
	if len(tags) == 0 {
		return query(config, service, "")
	}

	res, ok := (&basePStream{stream.FromArray(tags)}).PMap(func(tag interface{}) interface{} {
		return query(config, service, tag.(string))
	}).Reduce(mergeFunc).Value()

	if ok {
		return res.([]*api.CatalogService)
	} else {
		// perhaps blow up instead?
		return nil
	}
}

func query(config *api.Config, service string, tag string) []*api.CatalogService {
	client, err := api.NewClient(config)

	services, _, err := client.Catalog().Service(service, tag, &api.QueryOptions{AllowStale: true, RequireConsistent: false})
	if err != nil {
		kingpin.Fatalf("Error querying Consul: %s\n", err.Error())
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
