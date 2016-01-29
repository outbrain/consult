package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/exec"
	"syscall"
)

type sshCommand struct {
	QueryCommand
	user string
}

func sshRegisterCli(app *kingpin.Application, opts *appOpts) {
	s := &sshCommand{}
	s.IQuery = s
	s.opts = opts
	sshCmd := app.Command("ssh", "ssh into server using Consul query").Action(s.run)
	sshCmd.Flag("username", "ssh user name").Short('u').StringVar(&s.user)
	s.registerCli(sshCmd)
}

func (s *sshCommand) run(c *kingpin.ParseContext) error {
	results_by_dc, err := s.queryServicesGeneric()
	if err != nil {
		return err
	}
	results := flattenSvcMap(results_by_dc)
	if len(results) == 0 {
		kingpin.Errorf("No results from query\n")
		return nil
	}
	ssh(selectRandomSvc(results).Node, s.user)
	return nil
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
