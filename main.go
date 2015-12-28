package main

import (
	"github.com/hashicorp/consul/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"math/rand"
	"os"
	"os/exec"
	"syscall"
)

var (
	serverURL = kingpin.Flag("server", "Consul URL").Default("http://127.0.0.1:8500/").Envar("CONSUL_URL").URL()
	tag       = kingpin.Flag("tag", "Consul tag").String()
	service   = kingpin.Flag("service", "Consul service").Required().String()
	dc        = kingpin.Flag("dc", "Consul datacenter").String()
	user      = kingpin.Flag("username", "ssh user name").String()
)

func main() {
	kingpin.Parse()
	config := &api.Config{Address: (*serverURL).Host, Scheme: (*serverURL).Scheme}
	if *dc != "" {
		config.Datacenter = *dc
	}

	client, err := api.NewClient(config)

	services, _, err := client.Catalog().Service(*service, *tag, &api.QueryOptions{AllowStale: true, RequireConsistent: false})
	if err != nil {
		kingpin.Fatalf("Error querying Consul: %s\n", err.Error())
	}

	address := services[rand.Intn(len(services))].Node
	bin, err := exec.LookPath("ssh")
	if err != nil {
		kingpin.Fatalf("Failed to find ssh binary: %s\n", err.Error())
	}

	ssh_args := make([]string, 2, 3)
	ssh_args[0] = "ssh"
	ssh_args[1] = address
	if *user != "" {
		ssh_args = append(ssh_args, "-l "+*user)
	}

	syscall.Exec(bin, ssh_args, os.Environ())
}
