/*
 * SPDX-License-Identifier: Apache-2.0
 */

package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

// Proxy represents a proxy to CouchDB.
type Proxy struct {
	transport  http.RoundTripper
	prefix     string
	httpServer *http.Server
}

// NewProxy creates a new proxy to CouchDB.
func (c *CouchDB) NewProxy(prefix string, port int) (*Proxy, error) {
	result := &Proxy{transport: http.DefaultTransport, prefix: prefix}
	director := func(req *http.Request) {
		req.URL.Scheme = c.url.Scheme
		req.URL.Host = c.url.Host
		if req.URL.Path == "/_all_dbs" {
			// Do nothing.
		} else if req.URL.Path == "/" || strings.HasPrefix(req.URL.Path, "/_") {
			// Do nothing.
		} else {
			// Rewrite to include the prefix.
			req.URL.Path = "/" + prefix + "_" + req.URL.Path[1:]
			if len(req.URL.RawPath) != 0 {
				req.URL.RawPath = "/" + prefix + "_" + req.URL.RawPath[1:]
			}
		}
	}
	reverseProxy := &httputil.ReverseProxy{
		Director:      director,
		Transport:     result,
		FlushInterval: -1,
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: reverseProxy,
	}
	result.httpServer = httpServer
	return result, nil
}

// Start starts the proxy to CouchDB.
func (p *Proxy) Start() error {
	return p.httpServer.ListenAndServe()
}

// Stop stops the proxy to CouchDB.
func (p *Proxy) Stop() error {
	return p.httpServer.Close()
}

// RoundTrip handles a request to CouchDB.
func (p *Proxy) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.URL.Path == "/_all_dbs" {
		return p.handleAllDbs(req)
	}
	return p.transport.RoundTrip(req)
}

func (p *Proxy) handleAllDbs(req *http.Request) (*http.Response, error) {
	resp, err := p.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, nil
	}
	allDBs := []string{}
	err = json.NewDecoder(resp.Body).Decode(&allDBs)
	if err != nil {
		return nil, err
	}
	filteredDBs := []string{}
	for _, db := range allDBs {
		if strings.HasPrefix(db, "_") {
			filteredDBs = append(filteredDBs, db)
		} else if strings.HasPrefix(db, p.prefix+"_") {
			filteredDBs = append(filteredDBs, db[len(p.prefix)+1:])
		}
	}
	data, err := json.Marshal(filteredDBs)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	resp.ContentLength = int64(len(data))
	resp.Header.Set("Content-Length", strconv.Itoa(len(data)))
	return resp, nil
}
