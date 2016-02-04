package main

import (
	"errors"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

type MockCommand struct {
	Command
}

func TestMultiDCs(t *testing.T) {
	t.Parallel()
	c := new(MockCommand)
	c.opts = new(appOpts)
	c.opts.dcs = []string{"dc1", "dc2"}
	c.opts.serverURL = &url.URL{Host: "localhost", Scheme: "http"}
	clients, err := c.GetConsulClients()

	assert.NoError(t, err, "GetConsulClients should not return error")
	assert.Len(t, clients, 2)
	assert.Contains(t, clients, "dc1")
	assert.Contains(t, clients, "dc2")

	var res map[string]interface{}

	// QueryWithClients should return error if returned from query function
	res, err = c.QueryWithClients(func(client *api.Client) interface{} {
		return errors.New("TEST")
	})
	assert.Error(t, err)
	assert.EqualError(t, err, "TEST")
	assert.Nil(t, res)
}
