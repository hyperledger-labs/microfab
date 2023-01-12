/*
 * SPDX-License-Identifier: Apache-2.0
 */

package txid_test

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/hyperledger-labs/microfab/internal/pkg/identity"
	"github.com/hyperledger-labs/microfab/internal/pkg/txid"
	"github.com/hyperledger-labs/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/msp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("the txid package", func() {

	var testIdentity *identity.Identity

	BeforeEach(func() {
		var err error
		testIdentity, err = identity.New("Org1Admin")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("txid.New()", func() {

		When("called", func() {
			It("creates a new transaction ID", func() {
				transactionID := txid.New("Org1MSP", testIdentity)
				Expect(transactionID.MSPID()).To(Equal("Org1MSP"))
				Expect(transactionID.Identity()).To(Equal(testIdentity))
				nonce := transactionID.Nonce()
				Expect(nonce).To(HaveLen(24))
				serializedIdentity := &msp.SerializedIdentity{
					Mspid:   "Org1MSP",
					IdBytes: testIdentity.Certificate().Bytes(),
				}
				sha := sha256.New()
				sha.Write(nonce)
				sha.Write(util.MarshalOrPanic(serializedIdentity))
				hash := sha.Sum(nil)
				expectedString := hex.EncodeToString(hash)
				Expect(transactionID.String()).To(Equal(expectedString))
			})
		})

	})

})
