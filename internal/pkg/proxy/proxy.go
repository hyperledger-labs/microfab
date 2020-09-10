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

	"github.com/IBM-Blockchain/microfab/internal/pkg/ca"
	"github.com/IBM-Blockchain/microfab/internal/pkg/console"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type route struct {
	SourceHost string
	TargetHost string
	UseHTTP2   bool
}

type routeMap map[string]*route

// Proxy represents an instance of a proxy.
type Proxy struct {
	httpServer *http.Server
}

type h2cTransportWrapper struct {
	*http2.Transport
}

func (tw *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return tw.Transport.RoundTrip(req)
}

var portRegex = regexp.MustCompile(":\\d+$")

// New creates a new instance of a proxy.
func New(console *console.Console, orderer *orderer.Orderer, peers []*peer.Peer, cas []*ca.CA, port int) (*Proxy, error) {
	routes := []*route{
		{
			SourceHost: console.URL().Host,
			TargetHost: fmt.Sprintf("localhost:%d", console.Port()),
			UseHTTP2:   false,
		},
		{
			SourceHost: orderer.APIHost(false),
			TargetHost: orderer.APIHost(true),
			UseHTTP2:   true,
		},
		{
			SourceHost: orderer.OperationsHost(false),
			TargetHost: orderer.OperationsHost(true),
			UseHTTP2:   false,
		},
	}
	for _, peer := range peers {
		orgRoutes := []*route{
			{
				SourceHost: peer.APIHost(false),
				TargetHost: peer.APIHost(true),
				UseHTTP2:   true,
			},
			{
				SourceHost: peer.ChaincodeHost(false),
				TargetHost: peer.ChaincodeHost(true),
				UseHTTP2:   true,
			},
			{
				SourceHost: peer.OperationsHost(false),
				TargetHost: peer.OperationsHost(true),
				UseHTTP2:   false,
			},
		}
		routes = append(routes, orgRoutes...)
	}
	for _, ca := range cas {
		orgRoutes := []*route{
			{
				SourceHost: ca.APIHost(false),
				TargetHost: ca.APIHost(true),
				UseHTTP2:   false,
			},
			{
				SourceHost: ca.OperationsHost(false),
				TargetHost: ca.OperationsHost(true),
				UseHTTP2:   false,
			},
		}
		routes = append(routes, orgRoutes...)
	}
	rm := routeMap{}
	for _, route := range routes {
		rm[route.SourceHost] = route
	}
	director := func(req *http.Request) {
		host := req.Host
		if !portRegex.MatchString(host) {
			host += fmt.Sprintf(":%d", port)
		}
		route, ok := rm[host]
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
	httpTransport.RegisterProtocol("h2c", &h2cTransportWrapper{
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
	return &Proxy{httpServer: httpServer}, nil
}

// Start starts the proxy.
func (p *Proxy) Start() error {
	return p.httpServer.ListenAndServe()
}

// Stop stops the proxy.
func (p *Proxy) Stop() error {
	return p.httpServer.Close()
}
