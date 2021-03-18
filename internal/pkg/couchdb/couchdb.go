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
	internalURL *nurl.URL
	externalURL *nurl.URL
}

// New creates a new CouchDB instance.
func New(internalURL, externalURL string) (*CouchDB, error) {
	parsedInternalURL, err := nurl.Parse(internalURL)
	if err != nil {
		return nil, err
	}
	parsedExternalURL, err := nurl.Parse(externalURL)
	if err != nil {
		return nil, err
	}
	return &CouchDB{internalURL: parsedInternalURL, externalURL: parsedExternalURL}, nil
}

// WaitFor waits for the CouchDB instance to start.
func (c *CouchDB) WaitFor(timeout time.Duration) error {
	timeoutCh := time.After(timeout)
	tick := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-timeoutCh:
			return errors.New("timeout whilst waiting for CouchDB to start")
		case <-tick:
			if c.hasStarted() {
				return nil
			}
		}
	}
}

// URL returns the URL of the CouchDB instance.
func (c *CouchDB) URL(internal bool) *url.URL {
	if internal {
		return c.internalURL
	}
	return c.externalURL
}

func (c *CouchDB) hasStarted() bool {
	upURL := c.internalURL.ResolveReference(&url.URL{Path: "/_up"}).String()
	resp, err := http.Get(upURL)
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
