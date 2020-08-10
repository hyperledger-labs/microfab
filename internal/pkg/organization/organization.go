/*
 * SPDX-License-Identifier: Apache-2.0
 */

package organization

import (
	"fmt"
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
