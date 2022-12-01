/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer_test

import (
	"io/ioutil"

	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("the orderer package", func() {

	var testOrganization *organization.Organization
	var testDirectory string

	BeforeEach(func() {
		var err error
		testOrganization, err = organization.New("Org1", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		testDirectory, err = ioutil.TempDir("", "ut-peer")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("orderer.New()", func() {

		When("called", func() {
			It("creates a new peer", func() {
				p, err := orderer.New(testOrganization, testDirectory, 8080, 7051, "grpc://orderer-api.127-0-0-1.nip.io:8080", 8443, "http://orderer-operations.127-0-0-1.nip.io:8080")
				Expect(err).NotTo(HaveOccurred())
				Expect(p.Organization()).To(Equal(testOrganization))
				Expect(p.MSPID()).To(Equal(testOrganization.MSPID()))
				Expect(p.APIHostname(false)).To(Equal("orderer-api.127-0-0-1.nip.io"))
				Expect(p.APIHostname(true)).To(Equal("localhost"))
				Expect(p.APIHost(false)).To(Equal("orderer-api.127-0-0-1.nip.io:8080"))
				Expect(p.APIHost(true)).To(Equal("localhost:7051"))
				Expect(p.APIPort(false)).To(BeEquivalentTo(8080))
				Expect(p.APIPort(true)).To(BeEquivalentTo(7051))
				Expect(p.APIURL(false).String()).To(BeEquivalentTo("grpc://orderer-api.127-0-0-1.nip.io:8080"))
				Expect(p.APIURL(true).String()).To(BeEquivalentTo("grpc://localhost:7051"))
				Expect(p.OperationsHost(false)).To(Equal("orderer-operations.127-0-0-1.nip.io:8080"))
				Expect(p.OperationsHost(true)).To(Equal("localhost:8443"))
				Expect(p.OperationsPort(false)).To(BeEquivalentTo(8080))
				Expect(p.OperationsPort(true)).To(BeEquivalentTo(8443))
				Expect(p.OperationsURL(false).String()).To(BeEquivalentTo("http://orderer-operations.127-0-0-1.nip.io:8080"))
				Expect(p.OperationsURL(true).String()).To(BeEquivalentTo("http://localhost:8443"))
			})
		})

		When("called with an invalid API URL", func() {
			It("returns an error", func() {
				_, err := orderer.New(testOrganization, testDirectory, 8080, 7051, "!@£$%^&*()_+", 8443, "http://orderer-operations.127-0-0-1.nip.io:8080")
				Expect(err).To(HaveOccurred())
			})
		})

		When("called with an invalid operations URL", func() {
			It("returns an error", func() {
				_, err := orderer.New(testOrganization, testDirectory, 8080, 7051, "grpc://orderer-api.127-0-0-1.nip.io:8080", 8443, "!@£$%^&*()_+")
				Expect(err).To(HaveOccurred())
			})
		})

	})

})
