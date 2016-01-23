package main

import (
	"github.com/hashicorp/consul/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"regexp"
	"strings"
)

type listCommand struct {
	Command
	filterExp string
}

func listRegisterCli(app *kingpin.Application, opts *appOpts) {
	l := &listCommand{}
	l.opts = opts
	list := app.Command("list", "List Consul entities")
	list.Flag("regex", "Filter by regular expression").Short('r').StringVar(&l.filterExp)
	list.Command("service", "List Consult services").Action(l.listServiceHandler)
	list.Command("dc", "List Consul DataCenters").Action(l.listDCHandler)
	list.Command("node", "List Consul nodes").Action(l.listNodeHandler)
}

func (l *listCommand) listServiceHandler(context *kingpin.ParseContext) error {
	allServices := make(map[string]map[string][]string)
	if consultClients, err := l.GetConsulClients(); err != nil {
		return err
	} else {
		short := make([]string, 0)
		long := make([]string, 0)
		long = append(long, "Datacenter\tService\tTags")
		long = append(long, "")
		if exp, err := regexp.Compile(l.filterExp); err != nil {
			return err
		} else {
			for dc, client := range consultClients {
				if services, _, err := client.Catalog().Services(&api.QueryOptions{}); err != nil {
					return err
				} else {
					for service, tags := range services {
						if exp == nil || exp.Match([]byte(service)) {
							short = append(short, dc+"\t"+service)
							long = append(long, dc+"\t"+service+"\t"+strings.Join(tags, ","))
						}
					}
					allServices[dc] = services
				}
			}
			l.Output(allServices, long, short)
		}
	}
	return nil
}

func (l *listCommand) listDCHandler(context *kingpin.ParseContext) error {
	dcs, err := l.listDCs()
	if err != nil {
		return err
	}
	l.Output(dcs, dcs, dcs)
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

func (l *listCommand) listNodeHandler(context *kingpin.ParseContext) error {
	if results, err := l.listNodes(); err != nil {
		return err
	} else {
		short := make([]string, 0)
		long := make([]string, 2)
		long[0] = "Datacenter\tNode\tAddress"
		long[1] = ""

		if exp, err := regexp.Compile(l.filterExp); err != nil {
			return err
		} else {
			for dc, nodes := range results {
				for _, node := range nodes {
					if exp == nil || exp.Match([]byte(node.Node)) {
						long = append(long, dc+"\t"+node.Node+"\t"+node.Address)
						short = append(short, dc+"\t"+node.Node)
					}
				}
			}

			l.Output(results, long, short)
		}
		return nil
	}
}

func (l *listCommand) listNodes() (map[string][]*api.Node, error) {
	if clients, err := l.GetConsulClients(); err != nil {
		return nil, err
	} else {
		results := make(map[string][]*api.Node)
		for dc, client := range clients {
			if nodes, _, err := client.Catalog().Nodes(&api.QueryOptions{}); err != nil {
				return nil, err
			} else {
				results[dc] = nodes
			}
		}
		return results, nil
	}
}
