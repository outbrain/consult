package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/wushilin/stream"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type httpReqResp struct {
	req  *http.Request
	resp *http.Response
	host string
	port int
}

type httpCommand struct {
	QueryCommand
	Method    string
	Body      string
	Headers   map[string]string
	Scheme    string
	Uri       string
	Endpoints bool
}

func httpCall(method string, scheme string, host string, port int, uri string, body string, headers map[string]string) (*httpReqResp, error) {
	body_reader := strings.NewReader(body)
	url_struct := &url.URL{Scheme: scheme, Host: host + ":" + strconv.Itoa(port), Path: uri}
	if req, err := http.NewRequest(method, url_struct.String(), body_reader); err == nil {
		for header, value := range headers {
			req.Header.Add(header, value)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		return &httpReqResp{req, resp, host, port}, err
	} else {
		return nil, err
	}
}

func printHttp(x *httpReqResp) {
	fmt.Printf("* Connect: %s:%d\n", x.host, x.port)
	if b, err := httputil.DumpRequestOut(x.req, false); err == nil {
		stream.FromArray(strings.Split(string(b), "\n")).Each(func(l string) {
			fmt.Println("> " + l)
		})
	}
	if b, err := httputil.DumpResponse(x.resp, true); err == nil {
		stream.FromArray(strings.Split(string(b), "\n")).Each(func(l string) {
			fmt.Println("< " + l)
		})
	}
}

func httpCallEndpoint(endpoint *api.CatalogService,
	method string,
	scheme string,
	uri string,
	body string,
	headers map[string]string,
) *httpReqResp {
	var host string
	if endpoint.ServiceAddress == "" {
		host = endpoint.Address
	} else {
		host = endpoint.Address
	}

	if reqResp, err := httpCall(method, scheme,
		host,
		endpoint.ServicePort, uri, body,
		headers); err != nil {
		kingpin.Fatalf("HTTP Request failed: %s\n", err.Error())
		return nil
	} else {
		return reqResp
	}
}

func httpRegisterCli(app *kingpin.Application, opts *appOpts) {
	h := &httpCommand{}
	h.opts = opts
	h.IQuery = h
	httpCmd := app.Command("http", "HTTP Query a Consul service endpoint").Action(h.run)
	httpCmd.Flag("method", "HTTP method to use").Default("GET").EnumVar(&h.Method, "GET", "POST", "DELETE", "PUT", "HEAD")
	httpCmd.Flag("body", "Request body").StringVar(&h.Body)
	httpCmd.Flag("header", "Request headers").Short('H').StringMapVar(&h.Headers)
	httpCmd.Flag("scheme", "Request scheme").Default("http").StringVar(&h.Scheme)
	httpCmd.Flag("uri", "Request URI path").Default("/").StringVar(&h.Uri)
	httpCmd.Flag("all-endpoints", "HTTP Query all endpoint").BoolVar(&h.Endpoints)
	h.registerCli(httpCmd)
}

func (h *httpCommand) run(c *kingpin.ParseContext) error {
	if results_by_dc, err := h.queryServicesGeneric(); err != nil {
		return err
	} else {
		results := flattenSvcMap(results_by_dc)
		if len(results) == 0 {
			kingpin.Errorf("No results from query\n")
		}
		if h.Endpoints {
			httpExecute(results, h.Method, h.Scheme, h.Uri, h.Body, h.Headers)
		} else {
			httpExecute([]*api.CatalogService{selectRandomSvc(results)}, h.Method, h.Scheme, h.Uri, h.Body, h.Headers)
		}
		return nil
	}
}

func httpExecute(endpoints []*api.CatalogService,
	method string,
	scheme string,
	uri string,
	body string,
	headers map[string]string,
) {
	(&basePStream{stream.FromArray(endpoints)}).PMap(
		func(endpoint interface{}) interface{} {
			return httpCallEndpoint(endpoint.(*api.CatalogService), method, scheme, uri, body, headers)
		},
	).Each(printHttp)
}
