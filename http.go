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

func httpCmdHandler(endpoints []*api.CatalogService,
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
