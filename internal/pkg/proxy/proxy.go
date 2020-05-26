/*
 * SPDX-License-Identifier: Apache-2.0
 */

package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"

	"github.com/IBM-Blockchain/microfab/internal/pkg/console"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Route struct {
	SourceHost string
	TargetHost string
	UseHTTP2   bool
}

type RouteMap map[string]*Route

type Proxy struct {
	HTTPServer *http.Server
}

type H2CTransportWrapper struct {
	*http2.Transport
}

func (tw *H2CTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return tw.Transport.RoundTrip(req)
}

var portRegex = regexp.MustCompile(":\\d+$")

func New(console *console.Console, orderer *orderer.Orderer, peers []*peer.Peer, port int) (*Proxy, error) {
	routes := []*Route{
		{
			SourceHost: console.URL().Host,
			TargetHost: fmt.Sprintf("localhost:%d", console.Port()),
			UseHTTP2:   false,
		},
		{
			SourceHost: orderer.APIURL().Host,
			TargetHost: fmt.Sprintf("localhost:%d", orderer.APIPort()),
			UseHTTP2:   true,
		},
		{
			SourceHost: orderer.OperationsURL().Host,
			TargetHost: fmt.Sprintf("localhost:%d", orderer.OperationsPort()),
			UseHTTP2:   false,
		},
	}
	for _, peer := range peers {
		orgRoutes := []*Route{
			{
				SourceHost: peer.APIURL().Host,
				TargetHost: fmt.Sprintf("localhost:%d", peer.APIPort()),
				UseHTTP2:   true,
			},
			{
				SourceHost: peer.ChaincodeURL().Host,
				TargetHost: fmt.Sprintf("localhost:%d", peer.ChaincodePort()),
				UseHTTP2:   true,
			},
			{
				SourceHost: peer.OperationsURL().Host,
				TargetHost: fmt.Sprintf("localhost:%d", peer.OperationsPort()),
				UseHTTP2:   false,
			},
		}
		routes = append(routes, orgRoutes...)
	}
	routeMap := RouteMap{}
	for _, route := range routes {
		routeMap[route.SourceHost] = route
	}
	director := func(req *http.Request) {
		host := req.Host
		if !portRegex.MatchString(host) {
			host += fmt.Sprintf(":%d", port)
		}
		route, ok := routeMap[host]
		if !ok {
			route = routes[0]
		}
		if route.UseHTTP2 {
			req.URL.Scheme = "h2c"
		} else {
			req.URL.Scheme = "http"
		}
		req.URL.Host = route.TargetHost
	}
	httpTransport := &http.Transport{}
	httpTransport.RegisterProtocol("h2c", &H2CTransportWrapper{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	})
	err := http2.ConfigureTransport(httpTransport)
	if err != nil {
		return nil, err
	}
	reverseProxy := &httputil.ReverseProxy{
		Director:      director,
		Transport:     httpTransport,
		FlushInterval: -1,
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h2c.NewHandler(reverseProxy, &http2.Server{}),
	}
	err = http2.ConfigureServer(httpServer, nil)
	if err != nil {
		return nil, err
	}
	return &Proxy{HTTPServer: httpServer}, nil
}

func (p *Proxy) Start() error {
	return p.HTTPServer.ListenAndServe()
}

func (p *Proxy) Stop() error {
	return p.HTTPServer.Close()
}
