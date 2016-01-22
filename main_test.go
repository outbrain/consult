package main

import (
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
}
