package main

import (
	"github.com/hashicorp/consul/api"
	"github.com/outbrain/consult/misc"
	"gopkg.in/alecthomas/kingpin.v2"
)

type healthCommand struct {
	Command
	arg string
}

func HealthRegisterCli(app *kingpin.Application, opts *appOpts) {
	h := &healthCommand{}
	h.opts = opts
	cmd := app.Command("health", "List health status of service endpoinds")
	n := cmd.Command("node", "Show node health status").Action(h.nodeHealth)
	svc := cmd.Command("service", "Show service health status").Action(h.serviceHealth)
	chk := cmd.Command("check", "Show checks for service").Action(h.checkHealth)
	st := cmd.Command("state", "Show checks in state").Action(h.stateHealth)
	n.Arg("node", "Node name").Required().StringVar(&h.arg)
	st.Arg("state", "State name").Default("any").EnumVar(&h.arg, "any", "passing", "critical", "warning", "unknown")
	svc.Arg("service", "Service name").Required().StringVar(&h.arg)
	chk.Arg("service", "Service name").Required().StringVar(&h.arg)
}

func (h *healthCommand) nodeHealth(c *kingpin.ParseContext) error {
	if res, err := h.QueryWithClients(func(client *api.Client) interface{} {
		if res, _, err := client.Health().Node(h.arg, &api.QueryOptions{}); err != nil {
			return err
		} else {
			return res
		}
	}); err != nil {
		return err
	} else {
		long, short := healthChecksToShortLongString(res)
		h.Output(res, long, short)
		return nil
	}
}

func (h *healthCommand) serviceHealth(c *kingpin.ParseContext) error {
	if res, err := h.QueryWithClients(func(client *api.Client) interface{} {
		if res, _, err := client.Health().Service(h.arg, "", false, &api.QueryOptions{}); err != nil {
			return err
		} else {
			return res
		}
	}); err != nil {
		return err
	} else {
		long, short := make([]string, 2), make([]string, 0)
		long[0] = "Datacenter" + misc.SEPARATOR + misc.StructHeaderLine(api.ServiceEntry{})
		long[1] = ""
		for dc, results := range res {
			for _, item := range results.([]*api.ServiceEntry) {
				long = append(long, dc+misc.SEPARATOR+misc.StructToString(item))
				for _, check := range item.Checks {
					short = append(short, misc.JoinWithSep(dc, item.Node.Node, item.Service.Service, check.CheckID, check.Status))
				}
			}
		}
		h.Output(res, long, short)
		return nil
	}
}

func (h *healthCommand) checkHealth(c *kingpin.ParseContext) error {
	if res, err := h.QueryWithClients(func(client *api.Client) interface{} {
		if res, _, err := client.Health().Checks(h.arg, &api.QueryOptions{}); err != nil {
			return err
		} else {
			return res
		}
	}); err != nil {
		return err
	} else {
		long, short := healthChecksToShortLongString(res)
		h.Output(res, long, short)
		return nil
	}
}

func (h *healthCommand) stateHealth(c *kingpin.ParseContext) error {
	if res, err := h.QueryWithClients(func(client *api.Client) interface{} {
		if res, _, err := client.Health().State(h.arg, &api.QueryOptions{}); err != nil {
			return err
		} else {
			return res
		}
	}); err != nil {
		return err
	} else {
		long, short := healthChecksToShortLongString(res)
		h.Output(res, long, short)
		return nil
	}
}

func healthChecksToShortLongString(res map[string]interface{}) ([]string, []string) {
	long, short := make([]string, 2), make([]string, 0)
	long[0] = "Datacenter" + misc.SEPARATOR + misc.StructHeaderLine(api.HealthCheck{})
	long[1] = ""
	for dc, results := range res {
		for _, item := range results.([]*api.HealthCheck) {
			long = append(long, dc+misc.SEPARATOR+misc.StructToString(item))
			short = append(short, misc.JoinWithSep(dc, item.Node, item.ServiceName, item.CheckID, item.Status))
		}
	}
	return long, short
}
