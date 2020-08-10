/*
 * SPDX-License-Identifier: Apache-2.0
 */

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity/certificate"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity/privatekey"
)

// Identity represents a loaded identity (X509 certificate and ECDSA private key pair).
type Identity struct {
	name        string
	certificate *certificate.Certificate
	privateKey  *privatekey.PrivateKey
	ca          *certificate.Certificate
	isCA        bool
}

type jsonIdentity struct {
	Name        string  `json:"name"`
	Certificate string  `json:"cert"`
	PrivateKey  string  `json:"private_key"`
	CA          *string `json:"ca"`
	IsCA        *bool   `json:"is_ca"`
}

type newIdentity struct {
	Template *x509.Certificate
	Parent   *x509.Certificate
	Signee   *ecdsa.PublicKey
	Signer   *ecdsa.PrivateKey
}

// Option is a type representing an option for creating a new identity.
type Option func(*newIdentity)

// UsingSigner uses the specified identity to sign the new identity.
func UsingSigner(signer *Identity) Option {
	return func(o *newIdentity) {
		o.Template.AuthorityKeyId = signer.certificate.Certificate().SubjectKeyId
		o.Parent = signer.certificate.Certificate()
		o.Signer = signer.privateKey.PrivateKey()
	}
}

// WithIsCA indicates whether or not the new identity is a CA.
func WithIsCA(isCA bool) Option {
	return func(o *newIdentity) {
		if isCA {
			o.Template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
			o.Template.IsCA = true
		}
	}
}

// WithOrganizationalUnit sets the OU field in the new identity.
func WithOrganizationalUnit(organizationalUnit string) Option {
	return func(o *newIdentity) {
		o.Template.Subject.OrganizationalUnit = []string{organizationalUnit}
	}
}

// New creates a new identity.
func New(name string, opts ...Option) (*Identity, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	notBefore := time.Now().Add(-5 * time.Minute).UTC()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	identity := &newIdentity{
		Template: &x509.Certificate{
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			SerialNumber:          serialNumber,
			BasicConstraintsValid: true,
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			Subject: pkix.Name{
				CommonName: name,
			},
			IsCA: false,
		},
	}
	identity.Parent = identity.Template
	identity.Signee = &privateKey.PublicKey
	identity.Signer = privateKey
	for _, opt := range opts {
		opt(identity)
	}
	publicKeyBytes := elliptic.Marshal(identity.Signee.Curve, identity.Signee.X, identity.Signee.Y)
	subjectKeyID := sha256.Sum256(publicKeyBytes)
	identity.Template.SubjectKeyId = subjectKeyID[:]
	bytes, err := x509.CreateCertificate(rand.Reader, identity.Template, identity.Parent, identity.Signee, identity.Signer)
	if err != nil {
		return nil, err
	}
	cert, err := certificate.FromBytes(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: bytes}))
	if err != nil {
		return nil, err
	}
	bytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	pk, err := privatekey.FromBytes(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes}))
	if err != nil {
		return nil, err
	}
	var ca *certificate.Certificate
	if identity.Template != identity.Parent {
		ca, err = certificate.FromBytes(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: identity.Parent.Raw}))
		if err != nil {
			return nil, err
		}
	}
	isCA := identity.Template.IsCA
	return &Identity{name, cert, pk, ca, isCA}, nil
}

// FromBase64 loads an identity from a base64 encoded string.
func FromBase64(data string) (*Identity, error) {
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return FromBytes(bytes)
}

// FromBytes loads an identity from JSON data.
func FromBytes(data []byte) (*Identity, error) {
	parsedJSON := jsonIdentity{}
	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return nil, err
	}
	cert, err := certificate.FromBase64(parsedJSON.Certificate)
	if err != nil {
		return nil, err
	}
	pk, err := privatekey.FromBase64(parsedJSON.PrivateKey)
	if err != nil {
		return nil, err
	}
	var ca *certificate.Certificate
	if parsedJSON.CA != nil {
		ca, err = certificate.FromBase64(*parsedJSON.CA)
		if err != nil {
			return nil, err
		}
	}
	isCA := false
	if parsedJSON.IsCA != nil {
		isCA = *parsedJSON.IsCA
	}
	return &Identity{parsedJSON.Name, cert, pk, ca, isCA}, nil
}

// ToBytes saves the identity to JSON data.
func (i *Identity) ToBytes() ([]byte, error) {
	cert := i.certificate.ToBase64()
	pk := i.privateKey.ToBase64()
	var ca *string
	if i.ca != nil {
		tmp := i.ca.ToBase64()
		ca = &tmp
	}
	var isCA *bool
	if i.isCA {
		tmp := true
		isCA = &tmp
	}
	serializedJSON := jsonIdentity{i.name, cert, pk, ca, isCA}
	return json.Marshal(serializedJSON)
}

// ToBase64 saves the identity to a base64 encoded string.
func (i *Identity) ToBase64() (string, error) {
	bytes, err := i.ToBytes()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Name returns the name of the identity.
func (i *Identity) Name() string {
	return i.name
}

// Certificate returns the loaded X509 certificate.
func (i *Identity) Certificate() *certificate.Certificate {
	return i.certificate
}

// PrivateKey returns the loaded ECDSA private key.
func (i *Identity) PrivateKey() *privatekey.PrivateKey {
	return i.privateKey
}

// CA returns the loaded X509 CA.
func (i *Identity) CA() *certificate.Certificate {
	return i.ca
}

// Sign returns a signature of the SHA256 hash over the specified data.
func (i *Identity) Sign(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, d := range data {
		hasher.Write(d)
	}
	hash := hasher.Sum(nil)
	signature, err := i.PrivateKey().PrivateKey().Sign(rand.Reader, hash, crypto.SHA256)
	if err != nil {
		panic(err)
	}
	return i.preventMallebility(signature)
}

func (i *Identity) preventMallebility(signature []byte) []byte {
	var parts struct {
		R, S *big.Int
	}
	_, err := asn1.Unmarshal(signature, &parts)
	if err != nil {
		panic(err)
	}
	halfOrder := new(big.Int).Rsh(elliptic.P256().Params().N, 1)
	if parts.S.Cmp(halfOrder) == 1 {
		parts.S.Sub(i.PrivateKey().PublicKey().Params().N, parts.S)
		signature, err = asn1.Marshal(parts)
		if err != nil {
			panic(err)
		}
	}
	return signature
}
