/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer_test

import (
	"io/ioutil"

	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("the peer package", func() {

	var testOrganization *organization.Organization
	var testDirectory string

	BeforeEach(func() {
		var err error
		testOrganization, err = organization.New("Org1")
		Expect(err).NotTo(HaveOccurred())
		testDirectory, err = ioutil.TempDir("", "ut-peer")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("peer.New()", func() {

		When("called", func() {
			It("creates a new peer", func() {
				p, err := peer.New(testOrganization, testDirectory, 7051, "grpc://org1peer-api.127-0-0-1.nip.io:8080", 7052, "grpc://org1peer-chaincode.127-0-0-1.nip.io:8080", 8443, "http://org1peer-operations.127-0-0-1.nip.io:8080", false, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(p.Organization()).To(Equal(testOrganization))
				Expect(p.MSPID()).To(Equal(testOrganization.MSPID()))
				Expect(p.APIHostname(false)).To(Equal("org1peer-api.127-0-0-1.nip.io"))
				Expect(p.APIHostname(true)).To(Equal("localhost"))
				Expect(p.APIHost(false)).To(Equal("org1peer-api.127-0-0-1.nip.io:8080"))
				Expect(p.APIHost(true)).To(Equal("localhost:7051"))
				Expect(p.APIPort(false)).To(BeEquivalentTo(8080))
				Expect(p.APIPort(true)).To(BeEquivalentTo(7051))
				Expect(p.APIURL(false).String()).To(BeEquivalentTo("grpc://org1peer-api.127-0-0-1.nip.io:8080"))
				Expect(p.APIURL(true).String()).To(BeEquivalentTo("grpc://localhost:7051"))
				Expect(p.ChaincodeHost(false)).To(Equal("org1peer-chaincode.127-0-0-1.nip.io:8080"))
				Expect(p.ChaincodeHost(true)).To(Equal("localhost:7052"))
				Expect(p.ChaincodePort(false)).To(BeEquivalentTo(8080))
				Expect(p.ChaincodePort(true)).To(BeEquivalentTo(7052))
				Expect(p.ChaincodeURL(false).String()).To(BeEquivalentTo("grpc://org1peer-chaincode.127-0-0-1.nip.io:8080"))
				Expect(p.ChaincodeURL(true).String()).To(BeEquivalentTo("grpc://localhost:7052"))
				Expect(p.OperationsHost(false)).To(Equal("org1peer-operations.127-0-0-1.nip.io:8080"))
				Expect(p.OperationsHost(true)).To(Equal("localhost:8443"))
				Expect(p.OperationsPort(false)).To(BeEquivalentTo(8080))
				Expect(p.OperationsPort(true)).To(BeEquivalentTo(8443))
				Expect(p.OperationsURL(false).String()).To(BeEquivalentTo("http://org1peer-operations.127-0-0-1.nip.io:8080"))
				Expect(p.OperationsURL(true).String()).To(BeEquivalentTo("http://localhost:8443"))
			})
		})

		When("called with an invalid API URL", func() {
			It("returns an error", func() {
				_, err := peer.New(testOrganization, testDirectory, 7051, "!@£$%^&*()_+", 7052, "grpc://org1peer-chaincode.127-0-0-1.nip.io:8080", 8443, "http://org1peer-operations.127-0-0-1.nip.io:8080", false, 0)
				Expect(err).To(HaveOccurred())
			})
		})

		When("called with an invalid chaincode URL", func() {
			It("returns an error", func() {
				_, err := peer.New(testOrganization, testDirectory, 7051, "grpc://org1peer-api.127-0-0-1.nip.io:8080", 7052, "!@£$%^&*()_+", 8443, "http://org1peer-operations.127-0-0-1.nip.io:8080", false, 0)
				Expect(err).To(HaveOccurred())
			})
		})

		When("called with an invalid operations URL", func() {
			It("returns an error", func() {
				_, err := peer.New(testOrganization, testDirectory, 7051, "grpc://org1peer-api.127-0-0-1.nip.io:8080", 7052, "grpc://org1peer-chaincode.127-0-0-1.nip.io:8080", 8443, "!@£$%^&*()_+", false, 0)
				Expect(err).To(HaveOccurred())
			})
		})

	})

})
