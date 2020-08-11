/*
 * SPDX-License-Identifier: Apache-2.0
 */

package organization

import (
	"fmt"
	"regexp"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
)

// Organization represents a loaded organization definition.
type Organization struct {
	name  string
	ca    *identity.Identity
	admin *identity.Identity
	mspID string
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
	return &Organization{name, ca, admin, mspID}, nil
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

// MSPID returns the MSP ID for the organization.
func (o *Organization) MSPID() string {
	return o.mspID
}
