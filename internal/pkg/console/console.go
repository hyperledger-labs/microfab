/*
 * SPDX-License-Identifier: Apache-2.0
 */

package console

import (
	gotls "crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/IBM-Blockchain/microfab/internal/pkg/ca"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/gorilla/mux"
)

var logger = log.New(os.Stdout, fmt.Sprintf("[%16s] ", "console"), log.LstdFlags)

type jsonHealth struct {
}

type jsonOptions struct {
	DefaultAuthority      string `json:"grpc.default_authority"`
	SSLTargetNameOverride string `json:"grpc.ssl_target_name_override"`
	RequestTimeout        int    `json:"request-timeout"`
}

type jsonTLSCACerts struct {
	PEM string `json:"pem"`
}

type jsonTLSCACertsList struct {
	PEM []string `json:"pem"`
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
	PEM               []byte       `json:"pem,omitempty"`
	TLSCARootCert     []byte       `json:"tls_ca_root_cert,omitempty"`
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
	PEM               []byte       `json:"pem,omitempty"`
	TLSCARootCert     []byte       `json:"tls_ca_root_cert,omitempty"`
}

type jsonCA struct {
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
	PEM               []byte       `json:"pem,omitempty"`
	TLSCert           []byte       `json:"tls_cert,omitempty"`
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
	Hide        bool   `json:"hide"`
}

type components map[string]interface{}

// Console represents an instance of a console.
type Console struct {
	httpServer       *http.Server
	staticComponents components
	orderer          *orderer.Orderer
	peers            []*peer.Peer
	cas              []*ca.CA
	port             int
	url              *url.URL
}

// New creates a new instance of a console.
func New(port int, curl string) (*Console, error) {
	parsedURL, err := url.Parse(curl)
	if err != nil {
		return nil, err
	}
	console := &Console{
		staticComponents: components{},
		port:             port,
		url:              parsedURL,
		orderer:          nil,
		peers:            []*peer.Peer{},
		cas:              []*ca.CA{},
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
	logger.Print("Created new console...")
	return console, nil
}

// EnableTLS enables TLS for this console.
func (c *Console) EnableTLS(tls *identity.Identity) error {
	certificate, err := gotls.X509KeyPair(tls.Certificate().Bytes(), tls.PrivateKey().Bytes())
	if err != nil {
		return err
	}
	c.httpServer.TLSConfig = &gotls.Config{
		Certificates: []gotls.Certificate{certificate},
	}
	return nil
}

// RegisterOrganization registers the specified organization with the console.
func (c *Console) RegisterOrganization(organization *organization.Organization) {
	logger.Printf("RegisterOrganization %v", organization)
	for _, identity := range organization.GetIdentities() {
		identityHide := identity != organization.Admin()
		id := strings.ToLower(identity.Name())
		id = strings.ReplaceAll(id, " ", "")
		c.staticComponents[id] = &jsonIdentity{
			ID:          id,
			DisplayName: identity.Name(),
			Type:        "identity",
			Certificate: identity.Certificate().Bytes(),
			PrivateKey:  identity.PrivateKey().Bytes(),
			CA:          identity.CA().Bytes(),
			MSPID:       organization.MSPID(),
			Wallet:      organization.Name(),
			Hide:        identityHide,
		}
	}
}

// RegisterOrderer registers the specified orderer with the console.
func (c *Console) RegisterOrderer(orderer *orderer.Orderer) {
	c.orderer = orderer
}

// RegisterPeer registers the specified peer with the console.
func (c *Console) RegisterPeer(peer *peer.Peer) {
	c.peers = append(c.peers, peer)
}

// RegisterCA registers the specified CA with the console.
func (c *Console) RegisterCA(ca *ca.CA) {
	c.cas = append(c.cas, ca)
}

// Start starts the console.
func (c *Console) Start() error {
	if c.httpServer.TLSConfig != nil {
		return c.httpServer.ListenAndServeTLS("", "")
	}
	return c.httpServer.ListenAndServe()
}

// Stop stops the console.
func (c *Console) Stop() error {
	return c.httpServer.Close()
}

// Port returns the port of the console.
func (c *Console) Port() int {
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
	logger.Print("Getting components for REST response")
	logger.Printf("%+v", c.staticComponents)
	logger.Printf("%+v", c.getDynamicComponents(req))
	components := []interface{}{}
	for _, component := range c.staticComponents {
		components = append(components, component)
	}
	for _, component := range c.getDynamicComponents(req) {
		components = append(components, component)
	}
	rw.Header().Add("Content-Type", "application/json")
	logger.Printf("components== %+v", components)
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

func (c *Console) getOrderer(req *http.Request) *jsonOrderer {
	result := &jsonOrderer{
		ID:          "orderer",
		DisplayName: "Orderer",
		Type:        "fabric-orderer",
		APIURL:      c.getDynamicURL(req, c.orderer.APIURL(false)),
		APIOptions: &jsonOptions{
			DefaultAuthority:      c.orderer.APIHost(false),
			SSLTargetNameOverride: c.orderer.APIHostname(false),
			RequestTimeout:        300 * 1000,
		},
		OperationsURL: c.getDynamicURL(req, c.orderer.OperationsURL(false)),
		OperationsOptions: &jsonOptions{
			DefaultAuthority:      c.orderer.OperationsHost(false),
			SSLTargetNameOverride: c.orderer.OperationsHostname(false),
			RequestTimeout:        300 * 1000,
		},
		MSPID:    "OrdererMSP",
		Identity: c.orderer.Organization().Admin().Name(),
		Wallet:   c.orderer.Organization().Name(),
	}
	if tls := c.orderer.TLS(); tls != nil {
		result.PEM = tls.CA().Bytes()
		result.TLSCARootCert = tls.CA().Bytes()
	}
	return result
}

func (c *Console) getPeer(req *http.Request, peer *peer.Peer) *jsonPeer {
	orgName := peer.Organization().Name()
	lowerOrgName := strings.ToLower(orgName)
	id := fmt.Sprintf("%speer", lowerOrgName)
	result := &jsonPeer{
		ID:          id,
		DisplayName: fmt.Sprintf("%s Peer", orgName),
		Type:        "fabric-peer",
		APIURL:      c.getDynamicURL(req, peer.APIURL(false)),
		APIOptions: &jsonOptions{
			DefaultAuthority:      peer.APIHost(false),
			SSLTargetNameOverride: peer.APIHostname(false),
			RequestTimeout:        300 * 1000,
		},
		ChaincodeURL: c.getDynamicURL(req, peer.ChaincodeURL(false)),
		ChaincodeOptions: &jsonOptions{
			DefaultAuthority:      peer.ChaincodeHost(false),
			SSLTargetNameOverride: peer.ChaincodeHostname(false),
			RequestTimeout:        300 * 1000,
		},
		OperationsURL: c.getDynamicURL(req, peer.OperationsURL(false)),
		OperationsOptions: &jsonOptions{
			DefaultAuthority:      peer.OperationsHost(false),
			SSLTargetNameOverride: peer.OperationsHostname(false),
			RequestTimeout:        300 * 1000,
		},
		MSPID:    peer.MSPID(),
		Identity: peer.Organization().Admin().Name(),
		Wallet:   peer.Organization().Name(),
	}
	if tls := peer.TLS(); tls != nil {
		result.PEM = tls.CA().Bytes()
		result.TLSCARootCert = tls.CA().Bytes()
	}
	return result
}

func (c *Console) getPeers(req *http.Request) []*jsonPeer {
	result := []*jsonPeer{}
	for _, peer := range c.peers {
		result = append(result, c.getPeer(req, peer))
	}
	return result
}

func (c *Console) getGateway(req *http.Request, peer *peer.Peer) map[string]interface{} {
	orgName := peer.Organization().Name()
	lowerOrgName := strings.ToLower(orgName)
	id := fmt.Sprintf("%sgateway", lowerOrgName)
	var ca *ca.CA
	for _, temp := range c.cas {
		if temp.Organization().Name() == peer.Organization().Name() {
			ca = temp
			break
		}
	}
	p := map[string]interface{}{
		"url": c.getDynamicURL(req, peer.APIURL(false)),
		"grpcOptions": map[string]interface{}{
			"grpc.default_authority":        peer.APIHost(false),
			"grpc.ssl_target_name_override": peer.APIHostname(false),
		},
	}
	if tls := peer.TLS(); tls != nil {
		p["tlsCACerts"] = map[string]string{
			"pem": string(tls.CA().Bytes()),
		}
	}
	result := map[string]interface{}{
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
			peer.APIHost(false): p,
		},
	}
	if ca != nil {
		organizations := result["organizations"].(map[string]interface{})
		organization := organizations[ca.Organization().Name()].(map[string]interface{})
		organization["certificateAuthorities"] = []interface{}{
			ca.APIHost(false),
		}
		c := map[string]interface{}{
			"url": c.getDynamicURL(req, ca.APIURL(false)),
		}
		if tls := ca.TLS(); tls != nil {
			c["tlsCACerts"] = map[string][]string{
				"pem": {string(tls.CA().Bytes())},
			}
		}
		result["certificateAuthorities"] = map[string]interface{}{
			ca.APIHost(false): c,
		}
	}
	return result
}

func (c *Console) getGateways(req *http.Request) []map[string]interface{} {
	result := []map[string]interface{}{}
	for _, peer := range c.peers {
		result = append(result, c.getGateway(req, peer))
	}
	return result
}

func (c *Console) getCA(req *http.Request, ca *ca.CA) *jsonCA {
	orgName := ca.Organization().Name()
	lowerOrgName := strings.ToLower(orgName)
	id := fmt.Sprintf("%sca", lowerOrgName)
	result := &jsonCA{
		ID:          id,
		DisplayName: fmt.Sprintf("%s CA", orgName),
		Type:        "fabric-ca",
		APIURL:      c.getDynamicURL(req, ca.APIURL(false)),
		APIOptions: &jsonOptions{
			DefaultAuthority:      ca.APIHost(false),
			SSLTargetNameOverride: ca.APIHostname(false),
			RequestTimeout:        300 * 1000,
		},
		OperationsURL: c.getDynamicURL(req, ca.OperationsURL(false)),
		OperationsOptions: &jsonOptions{
			DefaultAuthority:      ca.OperationsHost(false),
			SSLTargetNameOverride: ca.OperationsHostname(false),
			RequestTimeout:        300 * 1000,
		},
		MSPID:    ca.Organization().MSPID(),
		Identity: ca.Organization().CAAdmin().Name(),
		Wallet:   ca.Organization().Name(),
	}
	if tls := ca.TLS(); tls != nil {
		result.PEM = tls.CA().Bytes()
		result.TLSCert = tls.CA().Bytes()
	}
	return result
}

func (c *Console) getCAs(req *http.Request) []*jsonCA {
	result := []*jsonCA{}
	for _, ca := range c.cas {
		result = append(result, c.getCA(req, ca))
	}
	return result
}

func (c *Console) getDynamicComponents(req *http.Request) components {
	dynamicComponents := components{}
	dynamicComponents["orderer"] = c.getOrderer(req)
	peers := c.getPeers(req)
	for _, peer := range peers {
		dynamicComponents[peer.ID] = peer
	}
	gateways := c.getGateways(req)
	for _, gateway := range gateways {
		id := gateway["id"].(string)
		dynamicComponents[id] = gateway
	}
	cas := c.getCAs(req)
	for _, ca := range cas {
		dynamicComponents[ca.ID] = ca
	}
	return dynamicComponents
}
