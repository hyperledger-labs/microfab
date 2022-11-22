/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"
	"net/url"
	"os/exec"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
)

// Peer represents a loaded peer definition.
type Peer struct {
	organization   *organization.Organization
	identity       *identity.Identity
	mspID          string
	directory      string
	microfabPort   int32
	apiPort        int32
	apiURL         *url.URL
	chaincodePort  int32
	chaincodeURL   *url.URL
	operationsPort int32
	operationsURL  *url.URL
	couchDB        bool
	couchDBPort    int32
	gossipPort     int32
	gossipURL      *url.URL
	command        *exec.Cmd
	tls            *identity.Identity
}

// New creates a new peer.
func New(organization *organization.Organization, directory string, microfabPort int32, apiPort int32, apiURL string, chaincodePort int32, chaincodeURL string, operationsPort int32, operationsURL string, couchDB bool, couchDBPort int32, gossipPort int32, gossipURL string) (*Peer, error) {
	identityName := fmt.Sprintf("%s Peer", organization.Name())
	identity, err := identity.New(identityName, identity.WithOrganizationalUnit("peer"), identity.UsingSigner(organization.CA()))
	if err != nil {
		return nil, err
	}
	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	parsedChaincodeURL, err := url.Parse(chaincodeURL)
	if err != nil {
		return nil, err
	}
	parsedOperationsURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, err
	}
	parsedGossipURL, err := url.Parse(gossipURL)
	if err != nil {
		return nil, err
	}

	return &Peer{organization, identity, organization.MSPID(), directory, microfabPort, apiPort, parsedAPIURL, chaincodePort, parsedChaincodeURL, operationsPort, parsedOperationsURL, couchDB, couchDBPort, gossipPort, parsedGossipURL, nil, nil}, nil
}

// TLS gets the TLS identity for this peer.
func (p *Peer) TLS() *identity.Identity {
	return p.tls
}

// EnableTLS enables TLS for this peer.
func (p *Peer) EnableTLS(tls *identity.Identity) {
	p.tls = tls
}

// Organization returns the organization of the peer.
func (p *Peer) Organization() *organization.Organization {
	return p.organization
}

// MSPID returns the MSP ID of the peer.
func (p *Peer) MSPID() string {
	return p.mspID
}

// APIHostname returns the hostname of the peer.
func (p *Peer) APIHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return p.apiURL.Hostname()
}

// APIHost returns the host (hostname:port) of the peer.
func (p *Peer) APIHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.apiPort)
	}
	return fmt.Sprintf("%s:%d", p.APIHostname(false), p.microfabPort)
}

// APIPort returns the API port of the peer.
func (p *Peer) APIPort(internal bool) int32 {
	if internal {
		return p.apiPort
	}

	return p.microfabPort
}

// APIURL returns the API URL of the peer.
func (p *Peer) APIURL(internal bool) *url.URL {
	scheme := "grpc"
	if p.tls != nil {
		scheme = "grpcs"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, p.apiPort))
		return url
	}

	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, p.APIHost(false)))
	return url
}

// ChaincodeHostname returns the hostname of the peer.
func (p *Peer) ChaincodeHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return p.chaincodeURL.Hostname()
}

// ChaincodeHost returns the host (hostname:port) of the peer.
func (p *Peer) ChaincodeHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.chaincodePort)
	}
	return fmt.Sprintf("%s:%d", p.ChaincodeHostname(false), p.microfabPort)
}

// ChaincodePort returns the chaincode port of the peer.
func (p *Peer) ChaincodePort(internal bool) int32 {
	if internal {
		return p.chaincodePort
	}
	return p.microfabPort
}

// ChaincodeURL returns the chaincode URL of the peer.
func (p *Peer) ChaincodeURL(internal bool) *url.URL {
	scheme := "grpc"
	if p.tls != nil {
		scheme = "grpcs"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, p.chaincodePort))
		return url
	}
	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, p.ChaincodeHost(false)))
	return url
}

// OperationsHostname returns the hostname of the peer.
func (p *Peer) OperationsHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return p.operationsURL.Hostname()
}

// OperationsHost returns the host (hostname:port) of the peer.
func (p *Peer) OperationsHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.operationsPort)
	}
	return fmt.Sprintf("%s:%d", p.OperationsHostname(false), p.microfabPort)
}

// OperationsPort returns the operations port of the peer.
func (p *Peer) OperationsPort(internal bool) int32 {
	if internal {
		return p.operationsPort
	}
	return p.microfabPort
}

// OperationsURL returns the operations URL of the peer.
func (p *Peer) OperationsURL(internal bool) *url.URL {
	scheme := "http"
	if p.tls != nil {
		scheme = "https"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, p.operationsPort))
		return url
	}
	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, p.OperationsHost(false)))
	return url
}

func (p *Peer) GossipHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.gossipPort)
	}
	return fmt.Sprintf("%s:%d", p.GossipHostname(false), p.microfabPort)
}

func (p *Peer) GossipHostname(internal bool) string {
	if internal {
		return "localhost"
	}
	return p.gossipURL.Hostname()
}

func (p *Peer) GossipURL(internal bool) *url.URL {
	scheme := "http"
	if p.tls != nil {
		scheme = "https"
	}
	if internal {
		url, _ := url.Parse(fmt.Sprintf("%s://localhost:%d", scheme, p.gossipPort))
		return url
	}
	url, _ := url.Parse(fmt.Sprintf("%s://%s", scheme, p.GossipHost(false)))
	return url
}
