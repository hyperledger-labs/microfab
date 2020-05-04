/*
 * SPDX-License-Identifier: Apache-2.0
 */

package console

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/IBM-Blockchain/fablet/internal/pkg/orderer"
	"github.com/IBM-Blockchain/fablet/internal/pkg/organization"
	"github.com/IBM-Blockchain/fablet/internal/pkg/peer"
	"github.com/gorilla/mux"
)

type Health struct {
}

type Options struct {
	DefaultAuthority      string `json:"grpc.default_authority"`
	SSLTargetNameOverride string `json:"grpc.ssl_target_name_override"`
}

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

type Orderer struct {
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

type Identity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Certificate string `json:"cert"`
	PrivateKey  string `json:"private_key"`
	MSPID       string `json:"msp_id"`
	Wallet      string `json:"wallet"`
}

type Components map[string]interface{}

type Console struct {
	HTTPServer       *http.Server
	StaticComponents Components
	Orderer          *orderer.Orderer
	Peers            []*peer.Peer
	port             int32
	url              *url.URL
}

func New(organizations []*organization.Organization, orderer *orderer.Orderer, peers []*peer.Peer, port int32, url_ string) (*Console, error) {
	staticComponents := Components{}
	for _, organization := range organizations {
		orgName := organization.Name()
		lowerOrgName := strings.ToLower(orgName)
		id := fmt.Sprintf("%sadmin", lowerOrgName)
		admin := organization.Admin()
		staticComponents[id] = &Identity{
			ID:          id,
			DisplayName: admin.Name(),
			Type:        "identity",
			Certificate: admin.Certificate().ToBase64(),
			PrivateKey:  admin.PrivateKey().ToBase64(),
			MSPID:       organization.MSP().ID(),
			Wallet:      organization.Name(),
		}
	}
	parsedURL, err := url.Parse(url_)
	if err != nil {
		return nil, err
	}
	console := &Console{
		StaticComponents: staticComponents,
		port:             port,
		url:              parsedURL,
		Orderer:          orderer,
		Peers:            peers,
	}
	router := mux.NewRouter()
	router.HandleFunc("/ak/api/v1/health", console.GetHealth).Methods("GET")
	router.HandleFunc("/ak/api/v1/components", console.GetComponents).Methods("GET")
	router.HandleFunc("/ak/api/v1/components/{id}", console.GetComponent).Methods("GET")
	HTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	console.HTTPServer = HTTPServer
	return console, nil
}

func (c *Console) Start() error {
	return c.HTTPServer.ListenAndServe()
}

func (c *Console) Stop() error {
	return c.HTTPServer.Close()
}

func (c *Console) GetHealth(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(&Health{})
}

func (c *Console) GetComponents(rw http.ResponseWriter, req *http.Request) {
	components := []interface{}{}
	for _, component := range c.StaticComponents {
		components = append(components, component)
	}
	for _, component := range c.getDynamicComponents(req) {
		components = append(components, component)
	}
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(components)
}

func (c *Console) GetComponent(rw http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	component, ok := c.StaticComponents[id]
	if !ok {
		component, ok = c.getDynamicComponents(req)[id]
		if !ok {
			rw.WriteHeader(404)
			return
		}
	}
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(component)
}

// Port returns the port of the console.
func (c *Console) Port() int32 {
	return c.port
}

// URL returns the URL of the console.
func (c *Console) URL() *url.URL {
	return c.url
}

func (c *Console) getDynamicURL(req *http.Request, target *url.URL) string {
	usingDNS := req.Host == c.url.Host
	if usingDNS {
		return target.String()
	}
	updatedTarget, _ := url.Parse(target.String())
	updatedTarget.Host = req.Host
	return updatedTarget.String()
}

func (c *Console) getDynamicComponents(req *http.Request) Components {
	dynamicComponents := Components{}
	dynamicComponents["orderer"] = &Orderer{
		ID:          "orderer",
		DisplayName: "Orderer",
		Type:        "fabric-orderer",
		APIURL:      c.getDynamicURL(req, c.Orderer.APIURL()),
		APIOptions: &Options{
			DefaultAuthority:      c.Orderer.APIURL().Host,
			SSLTargetNameOverride: c.Orderer.APIURL().Host,
		},
		OperationsURL: c.getDynamicURL(req, c.Orderer.OperationsURL()),
		OperationsOptions: &Options{
			DefaultAuthority:      c.Orderer.OperationsURL().Host,
			SSLTargetNameOverride: c.Orderer.OperationsURL().Host,
		},
		MSPID:    "OrdererMSP",
		Identity: c.Orderer.Organization().Admin().Name(),
		Wallet:   c.Orderer.Organization().Name(),
	}
	for _, peer := range c.Peers {
		orgName := peer.Organization().Name()
		lowerOrgName := strings.ToLower(orgName)
		id := fmt.Sprintf("%speer", lowerOrgName)
		dynamicComponents[id] = &Peer{
			ID:          id,
			DisplayName: fmt.Sprintf("%s Peer", orgName),
			Type:        "fabric-peer",
			APIURL:      c.getDynamicURL(req, peer.APIURL()),
			APIOptions: &Options{
				DefaultAuthority:      peer.APIURL().Host,
				SSLTargetNameOverride: peer.APIURL().Host,
			},
			ChaincodeURL: c.getDynamicURL(req, peer.ChaincodeURL()),
			ChaincodeOptions: &Options{
				DefaultAuthority:      peer.ChaincodeURL().Host,
				SSLTargetNameOverride: peer.ChaincodeURL().Host,
			},
			OperationsURL: c.getDynamicURL(req, peer.OperationsURL()),
			OperationsOptions: &Options{
				DefaultAuthority:      peer.OperationsURL().Host,
				SSLTargetNameOverride: peer.OperationsURL().Host,
			},
			MSPID:    peer.MSPID(),
			Identity: peer.Organization().Admin().Name(),
			Wallet:   peer.Organization().Name(),
		}
		id = fmt.Sprintf("%sgateway", lowerOrgName)
		dynamicComponents[id] = map[string]interface{}{
			"id":           id,
			"display_name": fmt.Sprintf("%s Gateway", orgName),
			"type":         "gateway",
			"name":         fmt.Sprintf("%s Gateway", orgName),
			"version":      "1.0",
			"wallet":       peer.Organization().Name(),
			"client": map[string]interface{}{
				"organization": peer.Organization().Name(),
				"connection": map[string]interface{}{
					"timeout": map[string]interface{}{
						"peer": map[string]interface{}{
							"endorser": "300",
						},
						"orderer": "300",
					},
				},
			},
			"organizations": map[string]interface{}{
				peer.Organization().Name(): map[string]interface{}{
					"mspid": peer.MSPID(),
					"peers": []interface{}{
						peer.Host(),
					},
				},
			},
			"peers": map[string]interface{}{
				peer.Host(): map[string]interface{}{
					"url": c.getDynamicURL(req, peer.APIURL()),
					"grpcOptions": map[string]interface{}{
						"grpc.default_authority":        peer.Host(),
						"grpc.ssl_target_name_override": peer.Host(),
					},
				},
			},
		}
	}
	return dynamicComponents
}
