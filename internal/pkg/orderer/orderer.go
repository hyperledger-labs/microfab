/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
)

// Orderer represents a loaded orderer definition.
type Orderer struct {
	organization   *organization.Organization
	mspID          string
	identity       *identity.Identity
	directory      string
	apiPort        int32
	apiURL         *url.URL
	operationsPort int32
	operationsURL  *url.URL
	command        *exec.Cmd
}

// New creates a new orderer.
func New(organization *organization.Organization, directory string, apiPort int32, apiURL string, operationsPort int32, operationsURL string) (*Orderer, error) {
	identityName := fmt.Sprintf("%s Orderer", organization.Name())
	identity, err := identity.New(identityName, identity.WithOrganizationalUnit("orderer"), identity.UsingSigner(organization.CA()))
	if err != nil {
		return nil, err
	}
	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	parsedOperationsURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, err
	}
	return &Orderer{organization, organization.MSPID(), identity, directory, apiPort, parsedAPIURL, operationsPort, parsedOperationsURL, nil}, nil
}

// Organization returns the organization of the orderer.
func (o *Orderer) Organization() *organization.Organization {
	return o.organization
}

// MSPID returns the MSP ID of the orderer.
func (o *Orderer) MSPID() string {
	return o.mspID
}

// APIHostname returns the hostname of the orderer.
func (o *Orderer) APIHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return o.apiURL.Hostname()
}

// APIHost returns the host (hostname:port) of the orderer.
func (o *Orderer) APIHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", o.apiPort)
	}
	return o.apiURL.Host
}

// APIPort returns the API port of the orderer.
func (o *Orderer) APIPort(internal bool) int32 {
	if internal {
		return o.apiPort
	}
	port, _ := strconv.Atoi(o.apiURL.Port())
	return int32(port)
}

// APIURL returns the API URL of the orderer.
func (o *Orderer) APIURL(internal bool) *url.URL {
	if internal {
		url, _ := url.Parse(fmt.Sprintf("grpc://localhost:%d", o.apiPort))
		return url
	}
	return o.apiURL
}

// OperationsHost returns the host (hostname:port) of the orderer.
func (o *Orderer) OperationsHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", o.operationsPort)
	}
	return o.operationsURL.Host
}

// OperationsPort returns the operations port of the orderer.
func (o *Orderer) OperationsPort(internal bool) int32 {
	if internal {
		return o.operationsPort
	}
	port, _ := strconv.Atoi(o.operationsURL.Port())
	return int32(port)
}

// OperationsURL returns the operations URL of the orderer.
func (o *Orderer) OperationsURL(internal bool) *url.URL {
	if internal {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", o.operationsPort))
		return url
	}
	return o.operationsURL
}
