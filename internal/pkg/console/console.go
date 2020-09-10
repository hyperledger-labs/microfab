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

	"github.com/IBM-Blockchain/microfab/internal/pkg/ca"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/gorilla/mux"
)

type jsonHealth struct {
}

type jsonOptions struct {
	DefaultAuthority      string `json:"grpc.default_authority"`
	SSLTargetNameOverride string `json:"grpc.ssl_target_name_override"`
	RequestTimeout        int    `json:"request-timeout"`
}

type jsonPeer struct {
	ID                string       `json:"id"`
	DisplayName       string       `json:"display_name"`
	Type              string       `json:"type"`
	APIURL            string       `json:"api_url"`
	APIOptions        *jsonOptions `json:"api_options"`
	ChaincodeURL      string       `json:"chaincode_url"`
	ChaincodeOptions  *jsonOptions `json:"chaincode_options"`
	OperationsURL     string       `json:"operations_url"`
	OperationsOptions *jsonOptions `json:"operations_options"`
	MSPID             string       `json:"msp_id"`
	Wallet            string       `json:"wallet"`
	Identity          string       `json:"identity"`
}

type jsonOrderer struct {
	ID                string       `json:"id"`
	DisplayName       string       `json:"display_name"`
	Type              string       `json:"type"`
	APIURL            string       `json:"api_url"`
	APIOptions        *jsonOptions `json:"api_options"`
	OperationsURL     string       `json:"operations_url"`
	OperationsOptions *jsonOptions `json:"operations_options"`
	MSPID             string       `json:"msp_id"`
	Wallet            string       `json:"wallet"`
	Identity          string       `json:"identity"`
}

type jsonIdentity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Certificate []byte `json:"cert"`
	PrivateKey  []byte `json:"private_key"`
	CA          []byte `json:"ca"`
	MSPID       string `json:"msp_id"`
	Wallet      string `json:"wallet"`
}

type components map[string]interface{}

// Console represents an instance of a console.
type Console struct {
	httpServer       *http.Server
	staticComponents components
	orderer          *orderer.Orderer
	peers            []*peer.Peer
	cas              []*ca.CA
	port             int32
	url              *url.URL
}

// New creates a new instance of a console.
func New(organizations []*organization.Organization, orderer *orderer.Orderer, peers []*peer.Peer, cas []*ca.CA, port int32, curl string) (*Console, error) {
	staticComponents := components{}
	for _, organization := range organizations {
		orgName := organization.Name()
		lowerOrgName := strings.ToLower(orgName)
		id := fmt.Sprintf("%sadmin", lowerOrgName)
		admin := organization.Admin()
		staticComponents[id] = &jsonIdentity{
			ID:          id,
			DisplayName: admin.Name(),
			Type:        "identity",
			Certificate: admin.Certificate().Bytes(),
			PrivateKey:  admin.PrivateKey().Bytes(),
			CA:          admin.CA().Bytes(),
			MSPID:       organization.MSPID(),
			Wallet:      organization.Name(),
		}
	}
	parsedURL, err := url.Parse(curl)
	if err != nil {
		return nil, err
	}
	console := &Console{
		staticComponents: staticComponents,
		port:             port,
		url:              parsedURL,
		orderer:          orderer,
		peers:            peers,
		cas:              cas,
	}
	router := mux.NewRouter()
	router.HandleFunc("/ak/api/v1/health", console.getHealth).Methods("GET")
	router.HandleFunc("/ak/api/v1/components", console.getComponents).Methods("GET")
	router.HandleFunc("/ak/api/v1/components/{id}", console.getComponent).Methods("GET")
	HTTPServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	console.httpServer = HTTPServer
	return console, nil
}

// Start starts the console.
func (c *Console) Start() error {
	return c.httpServer.ListenAndServe()
}

// Stop stops the console.
func (c *Console) Stop() error {
	return c.httpServer.Close()
}

// Port returns the port of the console.
func (c *Console) Port() int32 {
	return c.port
}

// URL returns the URL of the console.
func (c *Console) URL() *url.URL {
	return c.url
}

func (c *Console) getHealth(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(&jsonHealth{})
}

func (c *Console) getComponents(rw http.ResponseWriter, req *http.Request) {
	components := []interface{}{}
	for _, component := range c.staticComponents {
		components = append(components, component)
	}
	for _, component := range c.getDynamicComponents(req) {
		components = append(components, component)
	}
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(components)
}

func (c *Console) getComponent(rw http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["id"]
	component, ok := c.staticComponents[id]
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

func (c *Console) getDynamicURL(req *http.Request, target *url.URL) string {
	usingDNS := req.Host == c.url.Host
	if usingDNS {
		return target.String()
	}
	updatedTarget, _ := url.Parse(target.String())
	updatedTarget.Host = req.Host
	return updatedTarget.String()
}

func (c *Console) getDynamicComponents(req *http.Request) components {
	dynamicComponents := components{}
	dynamicComponents["orderer"] = &jsonOrderer{
		ID:          "orderer",
		DisplayName: "Orderer",
		Type:        "fabric-orderer",
		APIURL:      c.getDynamicURL(req, c.orderer.APIURL(false)),
		APIOptions: &jsonOptions{
			DefaultAuthority:      c.orderer.APIHost(false),
			SSLTargetNameOverride: c.orderer.APIHost(false),
			RequestTimeout:        300 * 1000,
		},
		OperationsURL: c.getDynamicURL(req, c.orderer.OperationsURL(false)),
		OperationsOptions: &jsonOptions{
			DefaultAuthority:      c.orderer.OperationsHost(false),
			SSLTargetNameOverride: c.orderer.OperationsHost(false),
			RequestTimeout:        300 * 1000,
		},
		MSPID:    "OrdererMSP",
		Identity: c.orderer.Organization().Admin().Name(),
		Wallet:   c.orderer.Organization().Name(),
	}
	for _, peer := range c.peers {
		orgName := peer.Organization().Name()
		lowerOrgName := strings.ToLower(orgName)
		id := fmt.Sprintf("%speer", lowerOrgName)
		dynamicComponents[id] = &jsonPeer{
			ID:          id,
			DisplayName: fmt.Sprintf("%s Peer", orgName),
			Type:        "fabric-peer",
			APIURL:      c.getDynamicURL(req, peer.APIURL(false)),
			APIOptions: &jsonOptions{
				DefaultAuthority:      peer.APIHost(false),
				SSLTargetNameOverride: peer.APIHost(false),
				RequestTimeout:        300 * 1000,
			},
			ChaincodeURL: c.getDynamicURL(req, peer.ChaincodeURL(false)),
			ChaincodeOptions: &jsonOptions{
				DefaultAuthority:      peer.ChaincodeHost(false),
				SSLTargetNameOverride: peer.ChaincodeHost(false),
				RequestTimeout:        300 * 1000,
			},
			OperationsURL: c.getDynamicURL(req, peer.OperationsURL(false)),
			OperationsOptions: &jsonOptions{
				DefaultAuthority:      peer.OperationsHost(false),
				SSLTargetNameOverride: peer.OperationsHost(false),
				RequestTimeout:        300 * 1000,
			},
			MSPID:    peer.MSPID(),
			Identity: peer.Organization().Admin().Name(),
			Wallet:   peer.Organization().Name(),
		}
		var ca *ca.CA
		for _, temp := range c.cas {
			if temp.Organization().Name() == peer.Organization().Name() {
				ca = temp
				break
			}
		}
		id = fmt.Sprintf("%sgateway", lowerOrgName)
		gateway := map[string]interface{}{
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
						peer.APIHost(false),
					},
				},
			},
			"peers": map[string]interface{}{
				peer.APIHost(false): map[string]interface{}{
					"url": c.getDynamicURL(req, peer.APIURL(false)),
					"grpcOptions": map[string]interface{}{
						"grpc.default_authority":        peer.APIHost(false),
						"grpc.ssl_target_name_override": peer.APIHost(false),
					},
				},
			},
		}
		if ca != nil {
			organizations := gateway["organizations"].(map[string]interface{})
			organization := organizations[ca.Organization().Name()].(map[string]interface{})
			organization["certificateAuthorities"] = []interface{}{
				ca.APIHost(false),
			}
			gateway["certificateAuthorities"] = map[string]interface{}{
				ca.APIHost(false): map[string]interface{}{
					"url": c.getDynamicURL(req, ca.APIURL(false)),
				},
			}
		}
		dynamicComponents[id] = gateway
	}
	for _, ca := range c.cas {
		orgName := ca.Organization().Name()
		lowerOrgName := strings.ToLower(orgName)
		id := fmt.Sprintf("%sca", lowerOrgName)
		dynamicComponents[id] = &jsonPeer{
			ID:          id,
			DisplayName: fmt.Sprintf("%s CA", orgName),
			Type:        "fabric-ca",
			APIURL:      c.getDynamicURL(req, ca.APIURL(false)),
			APIOptions: &jsonOptions{
				DefaultAuthority:      ca.APIHost(false),
				SSLTargetNameOverride: ca.APIHost(false),
				RequestTimeout:        300 * 1000,
			},
			OperationsURL: c.getDynamicURL(req, ca.OperationsURL(false)),
			OperationsOptions: &jsonOptions{
				DefaultAuthority:      ca.OperationsHost(false),
				SSLTargetNameOverride: ca.OperationsHost(false),
				RequestTimeout:        300 * 1000,
			},
			Identity: ca.Organization().Admin().Name(),
			Wallet:   ca.Organization().Name(),
		}
	}
	return dynamicComponents
}
