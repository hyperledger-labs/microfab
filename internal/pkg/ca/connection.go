/*
 * SPDX-License-Identifier: Apache-2.0
 */

package ca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/hyperledger-labs/microfab/internal/pkg/identity"
	"github.com/hyperledger-labs/microfab/pkg/client"
	"github.com/pkg/errors"
)

// Connection represents a connection to a certificate authority.
type Connection struct {
	ca *CA
}

// Connect opens a connection to the certificate authority.
func Connect(ca *CA) (*Connection, error) {
	return &Connection{ca: ca}, nil
}

// Close closes the connection to the certificate authority.
func (c *Connection) Close() error {
	return nil
}

// Enroll enrolls an identity using the certificate authority.
func (c *Connection) Enroll(name, id, secret string) (*identity.Identity, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: id,
		},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		return nil, err
	}
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
	request := map[string]interface{}{
		"certificate_request": string(csrPEM),
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/enroll", c.ca.APIURL(true)), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(id, secret)
	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		return nil, errors.Errorf("Failed to enroll using certificate authority: %s", resp.Status)
	}
	response := map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	success, ok := response["success"].(bool)
	if !success || !ok {
		return nil, errors.Errorf("Failed to enroll using certificate authority: %v", response)
	}
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("Invalid response to enroll using certificate authority: %v", response)
	}
	b64cert, ok := result["Cert"].(string)
	if !ok {
		return nil, errors.Errorf("Invalid response to enroll using certificate authority: %v", response)
	}
	cert, err := base64.StdEncoding.DecodeString(b64cert)
	if err != nil {
		return nil, err
	}
	pk, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return identity.FromClient(&client.Identity{
		DisplayName: name,
		Certificate: cert,
		PrivateKey:  pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pk}),
		CA:          c.ca.identity.Certificate().Bytes(),
	})
}
