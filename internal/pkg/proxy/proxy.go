/*
 * SPDX-License-Identifier: Apache-2.0
 */

package proxy

import (
	"crypto/tls"
	gotls "crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"

	"github.com/IBM-Blockchain/microfab/internal/pkg/ca"
	"github.com/IBM-Blockchain/microfab/internal/pkg/console"
	"github.com/IBM-Blockchain/microfab/internal/pkg/couchdb"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var logger = log.New(os.Stdout, fmt.Sprintf("[%16s] ", "console"), log.LstdFlags)

type route struct {
	SourceHost string
	TargetHost string
	UseHTTP2   bool
	UseTLS     bool
}

type routeMap map[string]*route

// Proxy represents an instance of a proxy.
type Proxy struct {
	httpServer *http.Server
	routes     []*route
	routeMap   routeMap
	tls        *identity.Identity
	caCertPool *x509.CertPool
	peerCert   gotls.Certificate
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
func New(port int) (*Proxy, error) {
	p := &Proxy{routeMap: routeMap{}}
	director := func(req *http.Request) {
		host := req.Host
		if !portRegex.MatchString(host) {
			host += fmt.Sprintf(":%d", port)
		}
		route, ok := p.routeMap[host]
		if !ok && len(p.routes) > 0 {
			route = p.routes[0]
			logger.Printf("No route found for '%s' assuming ['%s','%s']", host, route.SourceHost, route.TargetHost)
		}
		logger.Printf("Using route mapping for '%s' ['%s','%s','%t']", host, route.SourceHost, route.TargetHost, route.UseTLS)
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
			DialTLS: func(network, addr string, cfg *gotls.Config) (net.Conn, error) {
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
	p.httpServer = httpServer
	return p, nil
}

// NewWithTLS creates a new instance of a proxy that is TLS enabled.
func NewWithTLS(tls *identity.Identity, port int) (*Proxy, error) {
	p := &Proxy{routeMap: routeMap{}, tls: tls}
	p.caCertPool = x509.NewCertPool()
	p.caCertPool.AddCert(tls.CA().Certificate())

	director := func(req *http.Request) {
		host := req.Host
		logger.Printf("RemoteAddr=%s RequestURI=%s  Host=%s", req.RemoteAddr, req.RequestURI, host)
		if !portRegex.MatchString(host) {
			host += fmt.Sprintf(":%d", port)
		}
		route, ok := p.routeMap[host]
		if !ok && len(p.routes) > 0 {
			route = p.routes[0]
			logger.Printf("No route found for '%s' assuming ['%s','%s']", host, route.SourceHost, route.TargetHost)
		}
		logger.Printf("Using route mapping for '%s' ['%s','%s','%t','%t']", host, route.SourceHost, route.TargetHost, route.UseTLS, route.UseHTTP2)
		if route.UseTLS {
			req.URL.Scheme = "https"
		} else {
			req.URL.Scheme = "http"
		}
		req.URL.Host = route.TargetHost
	}

	// have attempted to ensure the TLS Cert is passed through. Used a customer 'dialTLS' implementation
	// but didn't seem to really work
	// DialTLS:           p.dialTLS,

	httpTransport := &http.Transport{
		TLSClientConfig: &gotls.Config{
			InsecureSkipVerify: true,
		},
		// DialTLS:           p.dialTLS,
		ForceAttemptHTTP2: true,
	}
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
		Handler: reverseProxy,
	}
	err = http2.ConfigureServer(httpServer, nil)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	certificate, err := gotls.X509KeyPair(tls.Certificate().Bytes(), tls.PrivateKey().Bytes())
	p.peerCert = certificate
	if err != nil {
		return nil, err
	}
	httpServer.TLSConfig = &gotls.Config{
		NextProtos: []string{
			"h2",
			"http/1.1",
		},
		Certificates: []gotls.Certificate{certificate},
	}
	p.httpServer = httpServer
	return p, nil
}

func (p *Proxy) dialTLS(network, addr string) (net.Conn, error) {
	logger.Printf("dialTLS %s %s", network, addr)
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	// host, _, err := net.SplitHostPort(addr)
	// if err != nil {
	// 	return nil, err
	// }
	cfg := &tls.Config{
		RootCAs:            p.caCertPool,
		Certificates:       []tls.Certificate{p.peerCert},
		InsecureSkipVerify: true,
	}

	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	cs := tlsConn.ConnectionState()
	cert := cs.PeerCertificates[0]

	// Verify here
	log.Println(cert.Subject)

	return tlsConn, nil
}

// RegisterConsole registers the specified console with the proxy.
func (p *Proxy) RegisterConsole(console *console.Console) {
	route := &route{
		SourceHost: console.URL().Host,
		TargetHost: fmt.Sprintf("localhost:%d", console.Port()),
		UseHTTP2:   false,
		UseTLS:     true,
	}
	p.routes = append(p.routes, route)
	p.buildRouteMap()
}

// RegisterPeer registers the specified peer with the proxy.
func (p *Proxy) RegisterPeer(peer *peer.Peer) {
	routes := []*route{
		{
			SourceHost: peer.APIHost(false),
			TargetHost: peer.APIHost(true),
			UseHTTP2:   true,
			UseTLS:     true,
		},
		{
			SourceHost: peer.ChaincodeHost(false),
			TargetHost: peer.ChaincodeHost(true),
			UseHTTP2:   true,
			UseTLS:     true,
		},
		{
			SourceHost: peer.OperationsHost(false),
			TargetHost: peer.OperationsHost(true),
			UseHTTP2:   false,
			UseTLS:     true,
		},
	}
	p.routes = append(p.routes, routes...)
	p.buildRouteMap()
}

// RegisterOrderer registers the specified orderer with the proxy.
func (p *Proxy) RegisterOrderer(orderer *orderer.Orderer) {
	routes := []*route{
		{
			SourceHost: orderer.APIHost(false),
			TargetHost: orderer.APIHost(true),
			UseHTTP2:   true,
			UseTLS:     true,
		},
		{
			SourceHost: orderer.OperationsHost(false),
			TargetHost: orderer.OperationsHost(true),
			UseHTTP2:   false,
			UseTLS:     true,
		},
	}
	p.routes = append(p.routes, routes...)
	p.buildRouteMap()
}

// RegisterCA registers the specified CA with the proxy.
func (p *Proxy) RegisterCA(ca *ca.CA) {
	routes := []*route{
		{
			SourceHost: ca.APIHost(false),
			TargetHost: ca.APIHost(true),
			UseHTTP2:   false,
			UseTLS:     true,
		},
		{
			SourceHost: ca.OperationsHost(false),
			TargetHost: ca.OperationsHost(true),
			UseHTTP2:   false,
			UseTLS:     true,
		},
	}
	p.routes = append(p.routes, routes...)
	p.buildRouteMap()
}

// RegisterCouchDB registers the specified CouchDB with the proxy.
func (p *Proxy) RegisterCouchDB(couchDB couchdb.CouchDB) {
	route := &route{
		SourceHost: couchDB.URL(false).Host,
		TargetHost: couchDB.URL(true).Host,
		UseHTTP2:   false,
		UseTLS:     false,
	}
	p.routes = append(p.routes, route)
	p.buildRouteMap()
}

// Start starts the proxy.
func (p *Proxy) Start() error {
	if p.tls != nil {
		return p.httpServer.ListenAndServeTLS("", "")
	}
	return p.httpServer.ListenAndServe()
}

// Stop stops the proxy.
func (p *Proxy) Stop() error {
	return p.httpServer.Close()
}

func (p *Proxy) buildRouteMap() {
	for _, route := range p.routes {
		p.routeMap[route.SourceHost] = route
	}
}

// DumpRouteMap logs the route mappings that have been registered
func (p *Proxy) DumpRouteMap() {
	for key, route := range p.routeMap {
		logger.Printf("%s ==> %v \n", key, route)
	}
}
