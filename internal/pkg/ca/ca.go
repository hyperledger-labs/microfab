/*
 * SPDX-License-Identifier: Apache-2.0
 */

package ca

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/hyperledger-labs/microfab/internal/pkg/identity"
	"github.com/hyperledger-labs/microfab/internal/pkg/organization"
)

// CA represents a loaded CA definition.
type CA struct {
	organization   *organization.Organization
	identity       *identity.Identity
	directory      string
	apiPort        int32
	apiURL         *url.URL
	operationsPort int32
	operationsURL  *url.URL
	command        *exec.Cmd
	tls            *identity.Identity
}

// New creates a new CA.
func New(organization *organization.Organization, directory string, apiPort int32, apiURL string, operationsPort int32, operationsURL string) (*CA, error) {
	identity := organization.CA()
	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	parsedOperationsURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, err
	}
	return &CA{organization, identity, directory, apiPort, parsedAPIURL, operationsPort, parsedOperationsURL, nil, nil}, nil
}

// TLS gets the TLS identity for this CA.
func (c *CA) TLS() *identity.Identity {
	return c.tls
}

// EnableTLS enables TLS for this CA.
func (c *CA) EnableTLS(tls *identity.Identity) {
	c.tls = tls
}

// Organization returns the organization of the CA.
func (c *CA) Organization() *organization.Organization {
	return c.organization
}

// APIHostname returns the hostname of the CA.
func (c *CA) APIHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return c.apiURL.Hostname()
}

// APIHost returns the host (hostname:port) of the CA.
func (c *CA) APIHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", c.apiPort)
	}
	return c.apiURL.Host
}

// APIPort returns the API port of the CA.
func (c *CA) APIPort(internal bool) int32 {
	if internal {
		return c.apiPort
	}
	port, _ := strconv.Atoi(c.apiURL.Port())
	return int32(port)
}

// APIURL returns the API URL of the CA.
func (c *CA) APIURL(internal bool) *url.URL {
	scheme := "http"
	if c.tls != nil {
		scheme = "https"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, c.apiPort))
		return url
	}
	return c.apiURL
}

// OperationsHostname returns the hostname of the CA.
func (c *CA) OperationsHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return c.operationsURL.Hostname()
}

// OperationsHost returns the host (hostname:port) of the CA.
func (c *CA) OperationsHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", c.operationsPort)
	}
	return c.operationsURL.Host
}

// OperationsPort returns the operations port of the CA.
func (c *CA) OperationsPort(internal bool) int32 {
	if internal {
		return c.operationsPort
	}
	port, _ := strconv.Atoi(c.operationsURL.Port())
	return int32(port)
}

// OperationsURL returns the operations URL of the CA.
func (c *CA) OperationsURL(internal bool) *url.URL {
	scheme := "http"
	if c.tls != nil {
		scheme = "https"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, c.operationsPort))
		return url
	}
	return c.operationsURL
}
