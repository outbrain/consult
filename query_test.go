package main

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

// copy-paste from https://github.com/hashicorp/consul/blob/master/api/api_test.go
type configCallback func(c *api.Config)

func makeClient(t *testing.T) (*api.Client, *testutil.TestServer) {
	return makeClientWithConfig(t, nil, nil)
}

func makeClientWithConfig(
	t *testing.T,
	cb1 configCallback,
	cb2 testutil.ServerConfigCallback) (*api.Client, *testutil.TestServer) {

	// Make client config
	conf := api.DefaultConfig()
	if cb1 != nil {
		cb1(conf)
	}

	// Create server
	server := testutil.NewTestServerConfig(t, cb2)
	conf.Address = server.HTTPAddr

	// Create client
	client, err := api.NewClient(conf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	return client, server
}

// end copy-paste

func TestQuery(t *testing.T) {
	t.Parallel()
	client, server := makeClient(t)
	server.AddService("testService", "passing", []string{"tag1", "tag2"})
	defer server.Stop()

	q := QueryCommand{service: "testService"}
	assert.Len(t, q.Query(client, ""), 1)
}

func TestQueryErr(t *testing.T) {
	t.Parallel()
	client, server := makeClient(t)
	server.Stop()
	q := QueryCommand{service: "testService"}
	assert.Panics(t, func() { q.Query(client, "") })
}

type QueryCommandMock struct {
	QueryCommand
	queriedTags []string
}

func (q *QueryCommandMock) Query(client *api.Client, tag string) []*api.CatalogService {
	q.queriedTags = append(q.queriedTags, tag)
	return []*api.CatalogService{}
}

func TestQueryMulti(t *testing.T) {
	t.Parallel()
	q := &QueryCommandMock{}
	q.IQuery = q
	q.service = "testService"
	q.tags = []string{"tag1", "tag2"}

	client, _ := api.NewClient(api.DefaultConfig())

	q.queryMulti(client, unionMerge)
	assert.Contains(t, q.queriedTags, "tag1")
	assert.Contains(t, q.queriedTags, "tag2")
	assert.Len(t, q.queriedTags, 2)
}

func TestQueryMultiNoTags(t *testing.T) {
	t.Parallel()
	q := &QueryCommandMock{}
	q.IQuery = q
	q.service = "testService"
	q.tags = []string{}

	client, _ := api.NewClient(api.DefaultConfig())

	q.queryMulti(client, unionMerge)
	assert.Contains(t, q.queriedTags, "")
	assert.Len(t, q.queriedTags, 1)
}
