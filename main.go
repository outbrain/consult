package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/wushilin/stream"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"time"
)

var (
	app = kingpin.New("consult", "Query Consul catalog for service")
)

type appOpts struct {
	dcs            []string
	allDCs         bool
	JsonFormat     bool
	DetailedOutput bool
	serverURL      *url.URL
	ConsulConfigs  []*api.Config
}

type Command struct {
	opts *appOpts
}

func main() {
	app.Version("0.0.7")
	opts := &appOpts{}

	app.Flag("dc", "Consul datacenter").StringsVar(&opts.dcs)
	app.Flag("all-dcs", "Query all datacenters").BoolVar(&opts.allDCs)
	app.Flag("server", "Consul URL; can also be provided using the CONSUL_URL environment variable").Default("http://127.0.0.1:8500/").Envar("CONSUL_URL").URLVar(&opts.serverURL)
	app.Flag("json", "JSON query output").Short('j').BoolVar(&opts.JsonFormat)
	app.Flag("detailed", "Detailed output (ignored if --json given)").Short('d').BoolVar(&opts.DetailedOutput)
	app.HelpFlag.Short('h')

	listRegisterCli(app, opts)
	httpRegisterCli(app, opts)
	sshRegisterCli(app, opts)
	queryRegisterCli(app, opts)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func flattenSvcMap(services_by_dc map[string][]*api.CatalogService) []*api.CatalogService {
	results := make([]*api.CatalogService, 0)
	for _, dc_results := range services_by_dc {
		results = append(results, dc_results...)
	}
	return results
}

func selectRandomSvc(services []*api.CatalogService) *api.CatalogService {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return services[r.Intn(len(services))]
}

func getCurrentDC(c *api.Client) (string, error) {
	if config, err := c.Agent().Self(); err != nil {
		return "", err
	} else {
		return config["Config"]["Datacenter"].(string), nil
	}
}

func (o *Command) GetConsulClient() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = o.opts.serverURL.Host
	config.Scheme = o.opts.serverURL.Scheme
	return api.NewClient(config)
}

func (o *Command) GetDCs() ([]string, error) {
	var client *api.Client
	var err error
	var dcs []string

	client, err = o.GetConsulClient()
	if err != nil {
		return nil, err
	}

	if o.opts.allDCs {
		dcs, err = client.Catalog().Datacenters()
		if err != nil {
			return nil, err
		} else {
			return dcs, nil
		}
	} else if len(o.opts.dcs) == 0 {
		// we want to have the name of the current dc
		var dc string
		if dc, err = getCurrentDC(client); err != nil {
			return nil, err
		} else {
			return []string{dc}, nil
		}
	} else {
		return o.opts.dcs, nil
	}
}

func (o *Command) GetConsulClients() (map[string]*api.Client, error) {
	clients := make(map[string]*api.Client, len(o.opts.dcs))
	if dcs, err := o.GetDCs(); err != nil {
		return nil, err
	} else {
		for _, dc := range dcs {
			config := api.DefaultConfig()
			config.Address = o.opts.serverURL.Host
			config.Scheme = o.opts.serverURL.Scheme
			config.Datacenter = dc

			if client, err := api.NewClient(config); err != nil {
				return nil, err
			} else {
				clients[dc] = client
			}
		}
		return clients, nil
	}
}

// Query in parallel. This function only deals with the parallelization logic
// returns - a map of dc -> query func results
func (o *Command) QueryWithClients(f func(*api.Client) interface{}) (map[string]interface{}, error) {
	if clients, err := o.GetConsulClients(); err != nil {
		return nil, err
	} else {
		results := make(map[string]interface{})
		(&basePStream{stream.FromMapEntries(clients)}).PMap(func(me interface{}) interface{} {
			dc := me.(stream.MapEntry).Key.(reflect.Value).String()
			client := me.(stream.MapEntry).Value.(*api.Client)

			res := f(client)
			return stream.MapEntry{dc, res}
		}).Each(func(me interface{}) {
			results[me.(stream.MapEntry).Key.(string)] = me.(stream.MapEntry).Value
		})

		return results, nil
	}
}

func (o *Command) Output(data interface{}, simpleLong []string, simpleShort []string) {
	if o.opts.JsonFormat {
		if b, err := json.MarshalIndent(data, "", "    "); err != nil {
			kingpin.Fatalf("Failed to convert results to json, %s\n", err.Error())
		} else {
			fmt.Println(string(b))
		}
	} else if o.opts.DetailedOutput {
		for _, line := range simpleLong {
			fmt.Println(line)
		}
	} else {
		for _, line := range simpleShort {
			fmt.Println(line)
		}
	}
}
