package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
)

var (
	app       = kingpin.New("consul-ssh", "Query Consul catalog for service")
	serverURL = app.Flag("server", "Consul URL").Default("http://127.0.0.1:8500/").Envar("CONSUL_URL").URL()
	tag       = app.Flag("tag", "Consul tag").String()
	service   = app.Flag("service", "Consul service").Required().String()
	dc        = app.Flag("dc", "Consul datacenter").String()
	queryCmd  = app.Command("query", "Query Consul catalog")
	// json      = queryCmd.Flag("json", "JSON query output").Default(false).Bool()
	sshCmd = app.Command("ssh", "ssh into server using Consul query")
	user   = sshCmd.Flag("username", "ssh user name").String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case queryCmd.FullCommand():
		printQueryResults(query(consulConfig(), *service, *tag))
	case sshCmd.FullCommand():
		ssh(selectRandomNode(query(consulConfig(), *service, *tag)), *user)
	}

}

func consulConfig() *api.Config {
	config := &api.Config{Address: (*serverURL).Host, Scheme: (*serverURL).Scheme}
	if *dc != "" {
		config.Datacenter = *dc
	}
	return config
}

func printQueryResults(results []*api.CatalogService) {
	for _, catalogService := range results {
		fmt.Println(catalogService.Node)
	}
}

func selectRandomNode(services []*api.CatalogService) string {
	return services[rand.Intn(len(services))].Node
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
