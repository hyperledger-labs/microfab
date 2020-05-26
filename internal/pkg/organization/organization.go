/*
 * SPDX-License-Identifier: Apache-2.0
 */

package organization

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity/certificate"
	"github.com/IBM-Blockchain/microfab/internal/pkg/msp"
)

// Organization represents a loaded organization definition.
type Organization struct {
	name  string
	ca    *identity.Identity
	admin *identity.Identity
	msp   *msp.MSP
}

type jsonOrganization struct {
	Name  string `json:"name"`
	CA    string `json:"ca"`
	Admin string `json:"admin"`
	MSP   string `json:"msp"`
}

// New creates a new organization.
func New(name string) (*Organization, error) {
	caName := fmt.Sprintf("%s CA", name)
	ca, err := identity.New(caName, identity.WithIsCA(true))
	if err != nil {
		return nil, err
	}
	adminName := fmt.Sprintf("%s Admin", name)
	admin, err := identity.New(adminName, identity.WithOrganizationalUnit("admin"), identity.UsingSigner(ca))
	if err != nil {
		return nil, err
	}
	safeRegex := regexp.MustCompile("[^a-zA-Z0-9]+")
	safeName := safeRegex.ReplaceAllString(name, "")
	mspID := fmt.Sprintf("%sMSP", safeName)
	msp, err := msp.New(mspID, []*certificate.Certificate{ca.Certificate()}, []*certificate.Certificate{admin.Certificate()})
	if err != nil {
		return nil, err
	}
	return &Organization{name, ca, admin, msp}, nil
}

// FromFile loads an organization from a JSON file.
func FromFile(file string) (*Organization, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

// FromBytes loads an organization from JSON data.
func FromBytes(data []byte) (*Organization, error) {
	parsedJSON := jsonOrganization{}
	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return nil, err
	}
	ca, err := identity.FromBase64(parsedJSON.CA)
	if err != nil {
		return nil, err
	}
	admin, err := identity.FromBase64(parsedJSON.Admin)
	if err != nil {
		return nil, err
	}
	msp, err := msp.FromBase64(parsedJSON.MSP)
	if err != nil {
		return nil, err
	}
	return &Organization{parsedJSON.Name, ca, admin, msp}, nil
}

// ToBytes saves the organization to JSON data.
func (o *Organization) ToBytes() ([]byte, error) {
	ca, err := o.ca.ToBase64()
	if err != nil {
		return nil, err
	}
	admin, err := o.admin.ToBase64()
	if err != nil {
		return nil, err
	}
	msp, err := o.msp.ToBase64()
	if err != nil {
		return nil, err
	}
	serializedJSON := jsonOrganization{o.name, ca, admin, msp}
	return json.Marshal(serializedJSON)
}

// ToFile saves the organization to a JSON file.
func (o *Organization) ToFile(file string) error {
	data, err := o.ToBytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, 0644)
}

// Name returns the name of the organization.
func (o *Organization) Name() string {
	return o.name
}

// CA returns the CA for the organization.
func (o *Organization) CA() *identity.Identity {
	return o.ca
}

// Admin returns the admin identity for the organization.
func (o *Organization) Admin() *identity.Identity {
	return o.admin
}

// MSP returns the MSP for the organization.
func (o *Organization) MSP() *msp.MSP {
	return o.msp
}
