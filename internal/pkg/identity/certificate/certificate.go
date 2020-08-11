/*
 * SPDX-License-Identifier: Apache-2.0
 */

package certificate

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
)

// Certificate represents a loaded X509 certificate.
type Certificate struct {
	certificate *x509.Certificate
	bytes       []byte
}

// FromBytes loads an X509 certificate from PEM data.
func FromBytes(data []byte) (*Certificate, error) {
	block, _ := pem.Decode(data)
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &Certificate{certificate, data}, nil
}

// Certificate returns the X509 certificate.
func (c *Certificate) Certificate() *x509.Certificate {
	return c.certificate
}

// Bytes returns the bytes of the X509 certificate.
func (c *Certificate) Bytes() []byte {
	return c.bytes
}

// Hash returns a SHA256 hash over the bytes of the X509 certificate.
func (c *Certificate) Hash() []byte {
	sha := sha256.New()
	sha.Write(c.bytes)
	return sha.Sum(nil)
}
