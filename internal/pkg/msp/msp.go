/*
 * SPDX-License-Identifier: Apache-2.0
 */

package msp

import (
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity/certificate"
)

// MSP represents a loaded MSP definition.
type MSP struct {
	mspID             string
	rootCertificates  []*certificate.Certificate
	adminCertificates []*certificate.Certificate
}

// New creates a new MSP.
func New(mspID string, rootCertificates, adminCertificates []*certificate.Certificate) (*MSP, error) {
	return &MSP{mspID, rootCertificates, adminCertificates}, nil
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
