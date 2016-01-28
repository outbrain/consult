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
	short := make([]string, 0)
	long := make([]string, 0)
	long = append(long, "Datacenter\tService\tTags")
	long = append(long, "")
	if exp, err := regexp.Compile(l.filterExp); err != nil {
		return err
	} else {
		results, err := l.QueryWithClients(func(client *api.Client) interface{} {
			if services, _, err := client.Catalog().Services(&api.QueryOptions{}); err != nil {
				panic(err)
				return nil
			} else {
				filtered_services := make(map[string][]string)
				for service, tags := range services {
					if exp == nil || exp.Match([]byte(service)) {
						filtered_services[service] = tags
					}
				}
				return filtered_services
			}
		})

		if err != nil {
			return err
		}

		// generate short and long text output
		for dc, dc_results := range results {
			for service, tags := range dc_results.(map[string][]string) {
				short = append(short, dc+"\t"+service)
				long = append(long, dc+"\t"+service+"\t"+strings.Join(tags, ","))
			}
		}

		l.Output(results, long, short)
	}

	return nil
}

func (l *listCommand) listDCHandler(context *kingpin.ParseContext) error {
	l.opts.allDCs = true
	if dcs, err := l.GetDCs(); err != nil {
		return err
	} else {
		l.Output(dcs, dcs, dcs)
		return nil
	}
}

func (l *listCommand) listNodeHandler(context *kingpin.ParseContext) error {
	if results, err := l.listNodes(); err != nil {
		return err
	} else {
		short := make([]string, 0)
		long := make([]string, 2)
		long[0] = "Datacenter\tNode\tAddress"
		long[1] = ""
		filtered_results := make(map[string][]*api.Node, 0)

		if exp, err := regexp.Compile(l.filterExp); err != nil {
			return err
		} else {
			for dc, nodes := range results {
				filtered_results[dc] = make([]*api.Node, 0)

				for _, node := range nodes {
					if exp == nil || exp.Match([]byte(node.Node)) {
						long = append(long, dc+"\t"+node.Node+"\t"+node.Address)
						short = append(short, dc+"\t"+node.Node)
						filtered_results[dc] = append(filtered_results[dc], node)
					}
				}
			}

			l.Output(filtered_results, long, short)
		}
		return nil
	}
}

func (l *listCommand) listNodes() (map[string][]*api.Node, error) {
	if results, err := l.QueryWithClients(func(client *api.Client) interface{} {
		if nodes, _, err := client.Catalog().Nodes(&api.QueryOptions{}); err != nil {
			panic(err)
			return nil
		} else {
			return nodes
		}
	}); err != nil {
		return nil, err
	} else {
		typed_results := make(map[string][]*api.Node)
		for dc, dc_res := range results {
			typed_results[dc] = dc_res.([]*api.Node)
		}
		return typed_results, nil
	}
}
