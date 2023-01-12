/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/hyperledger-labs/microfab/internal/pkg/identity"
	"github.com/hyperledger-labs/microfab/internal/pkg/organization"
)

// Orderer represents a loaded orderer definition.
type Orderer struct {
	organization   *organization.Organization
	mspID          string
	identity       *identity.Identity
	directory      string
	microfabPort   int32
	apiPort        int32
	apiURL         *url.URL
	operationsPort int32
	operationsURL  *url.URL
	command        *exec.Cmd
	tls            *identity.Identity
}

// New creates a new orderer.
func New(organization *organization.Organization, directory string, microFabPort int32, apiPort int32, apiURL string, operationsPort int32, operationsURL string) (*Orderer, error) {
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
	return &Orderer{organization, organization.MSPID(), identity, directory, microFabPort, apiPort, parsedAPIURL, operationsPort, parsedOperationsURL, nil, nil}, nil
}

// TLS gets the TLS identity for this orderer.
func (o *Orderer) TLS() *identity.Identity {
	return o.tls
}

// EnableTLS enables TLS for this orderer.
func (o *Orderer) EnableTLS(tls *identity.Identity) {
	o.tls = tls
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
	return fmt.Sprintf("%s:%d", o.APIHostname(false), o.microfabPort)
}

// APIPort returns the API port of the orderer.
func (o *Orderer) APIPort(internal bool) int32 {
	if internal {
		return o.apiPort
	}
	return o.microfabPort
}

// APIURL returns the API URL of the orderer.
func (o *Orderer) APIURL(internal bool) *url.URL {
	scheme := "grpc"
	if o.tls != nil {
		scheme = "grpcs"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, o.apiPort))
		return url
	}

	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, o.APIHost(false)))
	return url
}

// OperationsHostname returns the hostname of the orderer.
func (o *Orderer) OperationsHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return o.operationsURL.Hostname()
}

// OperationsHost returns the host (hostname:port) of the orderer.
func (o *Orderer) OperationsHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", o.operationsPort)
	}
	return fmt.Sprintf("%s:%d", o.OperationsHostname(false), o.microfabPort)
}

// OperationsPort returns the operations port of the orderer.
func (o *Orderer) OperationsPort(internal bool) int32 {
	if internal {
		return o.operationsPort
	}
	return o.microfabPort
}

// OperationsURL returns the operations URL of the orderer.
func (o *Orderer) OperationsURL(internal bool) *url.URL {
	scheme := "http"
	if o.tls != nil {
		scheme = "https"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, o.operationsPort))
		return url
	}
	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, o.OperationsHost(false)))
	return url
}
