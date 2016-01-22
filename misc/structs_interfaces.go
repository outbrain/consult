package misc

import (
	"github.com/hashicorp/consul/api"
)

type CatalogServicesMerger func(a []*api.CatalogService, b []*api.CatalogService) (mergedList []*api.CatalogService)

type IQuery interface {
	Query(client *api.Client, tag string) []*api.CatalogService
}
