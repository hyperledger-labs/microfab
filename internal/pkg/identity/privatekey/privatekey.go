/*
 * SPDX-License-Identifier: Apache-2.0
 */

package privatekey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// PrivateKey represents a loaded ECDSA private key.
type PrivateKey struct {
	privateKey *ecdsa.PrivateKey
	bytes      []byte
}

// FromBytes loads an ECDSA private key from PEM data.
func FromBytes(data []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(data)
	temp, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := temp.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("The specified private key is not an ECDSA private key")
	}
	return &PrivateKey{privateKey, data}, nil
}

// PrivateKey returns the ECDSA private key.
func (p *PrivateKey) PrivateKey() *ecdsa.PrivateKey {
	return p.privateKey
}

// PublicKey returns the ECDSA public key.
func (p *PrivateKey) PublicKey() *ecdsa.PublicKey {
	return &p.privateKey.PublicKey
}

// Bytes returns the bytes of the ECDSA private key.
func (p *PrivateKey) Bytes() []byte {
	return p.bytes
}
