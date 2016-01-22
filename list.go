package main

import (
	"github.com/hashicorp/consul/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

type listCommand struct {
	Command
}

func listRegisterCli(app *kingpin.Application, opts *appOpts) {
	l := &listCommand{}
	l.opts = opts
	list := app.Command("list", "List Consul entities")
	list.Command("service", "List Consult services").Action(l.listServiceHandler)
	list.Command("dc", "List Consul DataCenters").Action(l.listDCHandler)
}

func (l *listCommand) listServiceHandler(context *kingpin.ParseContext) error {
	allServices := make(map[string]map[string][]string)
	if consultClients, err := l.GetConsulClients(); err != nil {
		return err
	} else {
		for dc, client := range consultClients {

			if services, _, err := client.Catalog().Services(&api.QueryOptions{}); err != nil {
				return err
			} else {
				allServices[dc] = services
			}
		}
		l.Output(allServices)
	}
	return nil
}

func (l *listCommand) listDCHandler(context *kingpin.ParseContext) error {
	dcs, err := l.listDCs()
	if err != nil {
		return err
	}
	l.Output(dcs)
	return nil
}

func (l *listCommand) listDCs() ([]string, error) {
	client, err := l.GetConsulClient()
	if err != nil {
		return nil, err
	}

	dcs, qErr := client.Catalog().Datacenters()
	if qErr != nil {
		return nil, qErr
	}

	return dcs, nil
}
