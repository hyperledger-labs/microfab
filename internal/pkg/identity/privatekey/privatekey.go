/*
 * SPDX-License-Identifier: Apache-2.0
 */

package privatekey

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

// PrivateKey represents a loaded ECDSA private key.
type PrivateKey struct {
	privateKey *ecdsa.PrivateKey
	bytes      []byte
}

// FromFile loads an ECDSA private key from a PEM file.
func FromFile(file string) (*PrivateKey, error) {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return FromBytes(bytes)
}

// FromBase64 loads an ECDSA private key from a base64 encoded string.
func FromBase64(data string) (*PrivateKey, error) {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return FromBytes(bytes)
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

// ToBase64 saves the ECDSA private key to a base64 encoded string.
func (p *PrivateKey) ToBase64() string {
	return base64.StdEncoding.EncodeToString(p.bytes)
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
