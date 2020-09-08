/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
)

// Peer represents a loaded peer definition.
type Peer struct {
	organization   *organization.Organization
	identity       *identity.Identity
	mspID          string
	directory      string
	apiPort        int32
	apiURL         *url.URL
	chaincodePort  int32
	chaincodeURL   *url.URL
	operationsPort int32
	operationsURL  *url.URL
	couchDB        bool
	couchDBPort    int32
	command        *exec.Cmd
}

// New creates a new peer.
func New(organization *organization.Organization, directory string, apiPort int32, apiURL string, chaincodePort int32, chaincodeURL string, operationsPort int32, operationsURL string, couchDB bool, couchDBPort int32) (*Peer, error) {
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
	return &Peer{organization, identity, organization.MSPID(), directory, apiPort, parsedAPIURL, chaincodePort, parsedChaincodeURL, operationsPort, parsedOperationsURL, couchDB, couchDBPort, nil}, nil
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
	return p.apiURL.Host
}

// APIPort returns the API port of the peer.
func (p *Peer) APIPort(internal bool) int32 {
	if internal {
		return p.apiPort
	}
	port, _ := strconv.Atoi(p.apiURL.Port())
	return int32(port)
}

// APIURL returns the API URL of the peer.
func (p *Peer) APIURL(internal bool) *url.URL {
	if internal {
		url, _ := url.Parse(fmt.Sprintf("grpc://localhost:%d", p.apiPort))
		return url
	}
	return p.apiURL
}

// ChaincodeHost returns the host (hostname:port) of the peer.
func (p *Peer) ChaincodeHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.chaincodePort)
	}
	return p.chaincodeURL.Host
}

// ChaincodePort returns the chaincode port of the peer.
func (p *Peer) ChaincodePort(internal bool) int32 {
	if internal {
		return p.chaincodePort
	}
	port, _ := strconv.Atoi(p.chaincodeURL.Port())
	return int32(port)
}

// ChaincodeURL returns the chaincode URL of the peer.
func (p *Peer) ChaincodeURL(internal bool) *url.URL {
	if internal {
		url, _ := url.Parse(fmt.Sprintf("grpc://localhost:%d", p.chaincodePort))
		return url
	}
	return p.chaincodeURL
}

// OperationsHost returns the host (hostname:port) of the peer.
func (p *Peer) OperationsHost(internal bool) string {
	if internal {
		return fmt.Sprintf("localhost:%d", p.operationsPort)
	}
	return p.operationsURL.Host
}

// OperationsPort returns the operations port of the peer.
func (p *Peer) OperationsPort(internal bool) int32 {
	if internal {
		return p.operationsPort
	}
	port, _ := strconv.Atoi(p.operationsURL.Port())
	return int32(port)
}

// OperationsURL returns the operations URL of the peer.
func (p *Peer) OperationsURL(internal bool) *url.URL {
	if internal {
		url, _ := url.Parse(fmt.Sprintf("http://localhost:%d", p.operationsPort))
		return url
	}
	return p.operationsURL
}
