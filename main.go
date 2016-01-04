package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
)

var (
	app           = kingpin.New("consult", "Query Consul catalog for service")
	serverURL     = app.Flag("server", "Consul URL; can also be provided using the CONSUL_URL environment variable").Default("http://127.0.0.1:8500/").Envar("CONSUL_URL").URL()
	tags          = app.Flag("tag", "Consul tag").Short('t').Strings()
	service       = app.Flag("service", "Consul service").Required().Short('s').String()
	dc            = app.Flag("dc", "Consul datacenter").String()
	queryCmd      = app.Command("query", "Query Consul catalog")
	jsonFmt       = queryCmd.Flag("json", "JSON query output").Short('j').Bool()
	sshCmd        = app.Command("ssh", "ssh into server using Consul query")
	user          = sshCmd.Flag("username", "ssh user name").Short('u').String()
	tagsMerge     = app.Flag("tags-mode", "Find nodes with *all* or *any* of the tags").Short('m').Default("all").Enum("all", "any")
	httpCmd       = app.Command("http", "HTTP Query a Consul service endpoint")
	httpMethod    = httpCmd.Flag("method", "HTTP method to use").Default("GET").Enum("GET", "POST", "DELETE", "PUT", "HEAD")
	httpBody      = httpCmd.Flag("body", "Request body").String()
	httpHeaders   = httpCmd.Flag("header", "Request headers").Short('H').StringMap()
	httpScheme    = httpCmd.Flag("scheme", "Request scheme").Default("http").String()
	httpUri       = httpCmd.Flag("uri", "Request URI path").Default("/").String()
	httpEndpoints = httpCmd.Flag("all-endpoints", "HTTP Query all endpoint").Bool()
	_             = app.HelpFlag.Short('h')
)

func main() {
	app.Version("0.0.2")
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	mergeFunc := intersectionMerge
	if *tagsMerge == "any" {
		mergeFunc = unionMerge
	}
	results := queryMulti(consulConfig(), *service, *tags, mergeFunc)
	if len(results) == 0 {
		kingpin.Fatalf("No results from Consul query")
	}
	switch cmd {
	case queryCmd.FullCommand():
		if *jsonFmt {
			printJsonResults(results)
		} else {
			printQueryResults(results)
		}
	case sshCmd.FullCommand():
		ssh(selectRandomSvc(results).Node, *user)
	case httpCmd.FullCommand():
		if *httpEndpoints {
			httpCmdHandler(results, *httpMethod, *httpScheme, *httpUri, *httpBody, *httpHeaders)
		} else {
			httpCmdHandler([]*api.CatalogService{selectRandomSvc(results)}, *httpMethod, *httpScheme, *httpUri, *httpBody, *httpHeaders)
		}
	}
}

func consulConfig() *api.Config {
	config := &api.Config{Address: (*serverURL).Host, Scheme: (*serverURL).Scheme}
	if *dc != "" {
		config.Datacenter = *dc
	}
	return config
}

func printJsonResults(results []*api.CatalogService) {
	if b, err := json.MarshalIndent(results, "", "    "); err != nil {
		kingpin.Fatalf("Failed to convert results to json, %s\n", err.Error())
	} else {
		fmt.Println(string(b))
	}
}

func printQueryResults(results []*api.CatalogService) {
	for _, catalogService := range results {
		fmt.Println(catalogService.Node)
	}
}

func selectRandomSvc(services []*api.CatalogService) *api.CatalogService {
	return services[rand.Intn(len(services))]
}

func ssh(address string, user string) {
	bin, err := exec.LookPath("ssh")
	if err != nil {
		kingpin.Fatalf("Failed to find ssh binary: %s\n", err.Error())
	}

	ssh_args := make([]string, 2, 3)
	ssh_args[0] = "ssh"
	ssh_args[1] = address
	if user != "" {
		ssh_args = append(ssh_args, "-l "+user)
	}

	syscall.Exec(bin, ssh_args, os.Environ())
}
