/*
 * SPDX-License-Identifier: Apache-2.0
 */

package msp

import (
	"encoding/base64"
	"encoding/json"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity/certificate"
)

// MSP represents a loaded MSP definition.
type MSP struct {
	mspID             string
	rootCertificates  []*certificate.Certificate
	adminCertificates []*certificate.Certificate
}

type jsonMSP struct {
	MSPID             string   `json:"msp_id"`
	RootCertificates  []string `json:"root_certs"`
	AdminCertificates []string `json:"admins"`
}

// New creates a new MSP.
func New(mspID string, rootCertificates, adminCertificates []*certificate.Certificate) (*MSP, error) {
	return &MSP{mspID, rootCertificates, adminCertificates}, nil
}

// FromBase64 loads an MSP from a base64 encoded string.
func FromBase64(data string) (*MSP, error) {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return FromBytes(bytes)
}

// FromBytes loads an MSP from JSON data.
func FromBytes(data []byte) (*MSP, error) {
	parsedJSON := &jsonMSP{}
	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return nil, err
	}
	rootCertificates, err := parseCertificates(parsedJSON.RootCertificates)
	if err != nil {
		return nil, err
	}
	adminCertificates, err := parseCertificates(parsedJSON.AdminCertificates)
	if err != nil {
		return nil, err
	}
	return &MSP{parsedJSON.MSPID, rootCertificates, adminCertificates}, nil
}

// ToBytes saves the MSP to JSON data.
func (m *MSP) ToBytes() ([]byte, error) {
	rootCertificates := serializeCertificates(m.rootCertificates)
	adminCertificates := serializeCertificates(m.adminCertificates)
	serializedJSON := &jsonMSP{m.mspID, rootCertificates, adminCertificates}
	return json.Marshal(serializedJSON)
}

// ToBase64 saves the MSP to a base64 encoded string.
func (m *MSP) ToBase64() (string, error) {
	bytes, err := m.ToBytes()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// ID returns the ID of the MSP.
func (m *MSP) ID() string {
	return m.mspID
}

// RootCertificates returns the root certificates of the MSP.
func (m *MSP) RootCertificates() []*certificate.Certificate {
	return m.rootCertificates
}

// AdminCertificates returns the admin certificates of the MSP.
func (m *MSP) AdminCertificates() []*certificate.Certificate {
	return m.adminCertificates
}

func parseCertificates(certs []string) ([]*certificate.Certificate, error) {
	result := []*certificate.Certificate{}
	for _, cert := range certs {
		certificate, err := certificate.FromBase64(cert)
		if err != nil {
			return nil, err
		}
		result = append(result, certificate)
	}
	return result, nil
}

func serializeCertificates(certs []*certificate.Certificate) []string {
	result := []string{}
	for _, cert := range certs {
		result = append(result, cert.ToBase64())
	}
	return result
}
