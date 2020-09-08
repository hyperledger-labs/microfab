/*
 * SPDX-License-Identifier: Apache-2.0
 */

package couchdb

import (
	"net/http"
	"net/url"
	nurl "net/url"
	"time"

	"github.com/pkg/errors"
)

// CouchDB represents a CouchDB instance.
type CouchDB struct {
	url *nurl.URL
}

// New creates a new CouchDB instance.
func New(url string) (*CouchDB, error) {
	parsedURL, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}
	return &CouchDB{url: parsedURL}, nil
}

// WaitFor waits for the CouchDB instance to start.
func (c *CouchDB) WaitFor() error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return errors.New("timeout whilst waiting for CouchDB to start")
		case <-tick:
			if c.hasStarted() {
				return nil
			}
		}
	}
}

func (c *CouchDB) hasStarted() bool {
	upURL := c.url.ResolveReference(&url.URL{Path: "/_up"}).String()
	resp, err := http.Get(upURL)
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
