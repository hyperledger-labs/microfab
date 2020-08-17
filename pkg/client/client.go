/*
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// Client represents a Microfab client.
type Client struct {
	url *url.URL
}

// Options represents connection options for a peer or ordering service.
type Options struct {
	DefaultAuthority      string `json:"grpc.default_authority"`
	SSLTargetNameOverride string `json:"grpc.ssl_target_name_override"`
	RequestTimeout        int    `json:"request-timeout"`
}

// Peer represents a peer running inside Microfab.
type Peer struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"display_name"`
	Type              string   `json:"type"`
	APIURL            string   `json:"api_url"`
	APIOptions        *Options `json:"api_options"`
	ChaincodeURL      string   `json:"chaincode_url"`
	ChaincodeOptions  *Options `json:"chaincode_options"`
	OperationsURL     string   `json:"operations_url"`
	OperationsOptions *Options `json:"operations_options"`
	MSPID             string   `json:"msp_id"`
	Wallet            string   `json:"wallet"`
	Identity          string   `json:"identity"`
}

// OrderingService represents an ordering service running inside Microfab.
type OrderingService struct {
	ID                string   `json:"id"`
	DisplayName       string   `json:"display_name"`
	Type              string   `json:"type"`
	APIURL            string   `json:"api_url"`
	APIOptions        *Options `json:"api_options"`
	OperationsURL     string   `json:"operations_url"`
	OperationsOptions *Options `json:"operations_options"`
	MSPID             string   `json:"msp_id"`
	Wallet            string   `json:"wallet"`
	Identity          string   `json:"identity"`
}

// Identity represents an identity used for managing components inside Microfab.
type Identity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Certificate []byte `json:"cert"`
	PrivateKey  []byte `json:"private_key"`
	CA          []byte `json:"ca"`
	MSPID       string `json:"msp_id"`
	Wallet      string `json:"wallet"`
}

// New creates a new Microfab client.
func New(url *url.URL) (*Client, error) {
	return &Client{
		url: url,
	}, nil
}

// Ping tests the connection to Microfab.
func (c *Client) Ping() error {
	target := c.url.ResolveReference(&url.URL{Path: "/ak/api/v1/health"})
	resp, err := http.Get(target.String())
	if err != nil {
		return err
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.Errorf("Microfab returned HTTP %s", resp.Status)
	}
	return nil
}

// GetOrganizations gets the names of all of the organizations.
func (c *Client) GetOrganizations() ([]string, error) {
	components, err := c.getComponents()
	if err != nil {
		return nil, err
	}
	organizationNames := map[string]bool{}
	for _, component := range components {
		wallet, ok := component["wallet"]
		if !ok {
			continue
		}
		organizationNames[wallet.(string)] = true
	}
	result := []string{}
	for organizationName := range organizationNames {
		result = append(result, organizationName)
	}
	return result, nil
}

// GetPeer gets the peer for the specified organization.
func (c *Client) GetPeer(organization string) (*Peer, error) {
	components, err := c.getComponents()
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		ctype, ok := component["type"]
		if !ok {
			continue
		}
		wallet, ok := component["wallet"]
		if !ok {
			continue
		} else if ctype == "fabric-peer" && wallet == organization {
			data, err := json.Marshal(component)
			if err != nil {
				return nil, err
			}
			result := &Peer{}
			err = json.Unmarshal(data, result)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	return nil, errors.Errorf("Microfab does not have a peer for organization %s", organization)
}

// GetOrderingService gets the ordering service.
func (c *Client) GetOrderingService() (*OrderingService, error) {
	components, err := c.getComponents()
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		ctype, ok := component["type"]
		if !ok {
			continue
		} else if ctype == "fabric-orderer" {
			data, err := json.Marshal(component)
			if err != nil {
				return nil, err
			}
			result := &OrderingService{}
			err = json.Unmarshal(data, result)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	return nil, errors.New("Microfab does not have an ordering service")
}

// GetIdentity gets the identity for the specified organization.
func (c *Client) GetIdentity(organization string) (*Identity, error) {
	components, err := c.getComponents()
	if err != nil {
		return nil, err
	}
	for _, component := range components {
		ctype, ok := component["type"]
		if !ok {
			continue
		}
		wallet, ok := component["wallet"]
		if !ok {
			continue
		} else if ctype == "identity" && wallet == organization {
			data, err := json.Marshal(component)
			if err != nil {
				return nil, err
			}
			result := &Identity{}
			err = json.Unmarshal(data, result)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	return nil, errors.Errorf("Microfab does not have an admin identity for organization %s", organization)
}

func (c *Client) getComponents() ([]map[string]interface{}, error) {
	target := c.url.ResolveReference(&url.URL{Path: "/ak/api/v1/components"})
	resp, err := http.Get(target.String())
	if err != nil {
		return nil, err
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.Errorf("Microfab returned HTTP %s", resp.Status)
	}
	components := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&components)
	if err != nil {
		return nil, err
	}
	return components, nil
}
