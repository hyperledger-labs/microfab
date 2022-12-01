/*
 * SPDX-License-Identifier: Apache-2.0
 */

package organization_test

import (
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("the organization package", func() {

	Context("organization.New()", func() {

		When("called with a name with no special characters", func() {
			It("creates a new organization", func() {
				o, err := organization.New("Org1", nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(o.Name()).To(Equal("Org1"))
				Expect(o.MSPID()).To(Equal("Org1MSP"))
				ca := o.CA()
				Expect(ca).NotTo(BeNil())
				Expect(ca.Certificate().Certificate().Subject.CommonName).To(Equal("Org1 CA"))
				Expect(ca.Certificate().Certificate().IsCA).To(BeTrue())
				admin := o.Admin()
				Expect(admin).NotTo(BeNil())
				Expect(admin.Certificate().Certificate().Subject.CommonName).To(Equal("Org1 Admin"))
				Expect(admin.Certificate().Certificate().IsCA).To(BeFalse())
				err = admin.Certificate().Certificate().CheckSignatureFrom(ca.Certificate().Certificate())
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("called with a name with special characters", func() {
			It("creates a new organization", func() {
				o, err := organization.New("Org @ 1", nil, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(o.Name()).To(Equal("Org @ 1"))
				Expect(o.MSPID()).To(Equal("Org1MSP"))
				ca := o.CA()
				Expect(ca).NotTo(BeNil())
				Expect(ca.Certificate().Certificate().Subject.CommonName).To(Equal("Org @ 1 CA"))
				Expect(ca.Certificate().Certificate().IsCA).To(BeTrue())
				admin := o.Admin()
				Expect(admin).NotTo(BeNil())
				Expect(admin.Certificate().Certificate().Subject.CommonName).To(Equal("Org @ 1 Admin"))
				Expect(admin.Certificate().Certificate().IsCA).To(BeFalse())
				err = admin.Certificate().Certificate().CheckSignatureFrom(ca.Certificate().Certificate())
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})

})
