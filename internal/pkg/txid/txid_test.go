/*
 * SPDX-License-Identifier: Apache-2.0
 */

package txid_test

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/msp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const testIdentityText = `{
    "name": "Org1 Admin",
    "cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNQakNDQWVXZ0F3SUJBZ0lVWmJrVndnamV3TC9sTnhOYU1qQzFsb1VkMFNvd0NnWUlLb1pJemowRUF3SXcKV2pFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVFzd0NRWURWUVFERXdKallUQWVGdzB5Ck1EQTRNRE14TlRFMk1EQmFGdzB5TVRBNE1ETXhOVEl4TURCYU1DUXhEakFNQmdOVkJBc1RCV0ZrYldsdU1SSXcKRUFZRFZRUURFd2x2Y21jeFlXUnRhVzR3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVJTYUxaRApjTmMxb3NybU4wdkRzTzloa0dYMVBhUjV0NzJYNitCV3hDaURNWWZyTGNYczlkOW9LNU9oQkJWWkJ0dEkwWVR6Cm41aTN6cnZxSjhXOXFsOURvNEcrTUlHN01BNEdBMVVkRHdFQi93UUVBd0lIZ0RBTUJnTlZIUk1CQWY4RUFqQUEKTUIwR0ExVWREZ1FXQkJSYVAzVjNhbUplZmdaV1hWUUhyOUR5UFJ1RDZqQWZCZ05WSFNNRUdEQVdnQlR4TEwxbApxalExU1ovcG5wb1JuekFqUjBHbHF6QmJCZ2dxQXdRRkJnY0lBUVJQZXlKaGRIUnljeUk2ZXlKb1ppNUJabVpwCmJHbGhkR2x2YmlJNklpSXNJbWhtTGtWdWNtOXNiRzFsYm5SSlJDSTZJbTl5WnpGaFpHMXBiaUlzSW1obUxsUjUKY0dVaU9pSmhaRzFwYmlKOWZUQUtCZ2dxaGtqT1BRUURBZ05IQURCRUFpQXdzZ1pkaGlnVlB2ZTF2d3VJQTEydgpRRTdJaDFYeE1KWkJoZjAzL3pZTnVBSWdZRER5ckVCREp1N2JhS0hsbGNHWkRXK0pST29tS2dDejVuMFRKZStYCmYrTT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
    "ca": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNDVENDQWErZ0F3SUJBZ0lVWjdDSzJGYkthODR2Tmt1T241bzkwdWlyMHR3d0NnWUlLb1pJemowRUF3SXcKV2pFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVFzd0NRWURWUVFERXdKallUQWVGdzB5Ck1EQTRNRE14TlRFME1EQmFGdzB6TlRBM016RXhOVEUwTURCYU1Gb3hDekFKQmdOVkJBWVRBbFZUTVJjd0ZRWUQKVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFVU1CSUdBMVVFQ2hNTFNIbHdaWEpzWldSblpYSXhEekFOQmdOVgpCQXNUQmtaaFluSnBZekVMTUFrR0ExVUVBeE1DWTJFd1dUQVRCZ2NxaGtqT1BRSUJCZ2dxaGtqT1BRTUJCd05DCkFBVFoxd25tNkxzN2c0c2VYbkZUYm1uWlFZNFpJNWo1Z2dBMFh4L0g1UFhxYVFVN09LaTZaUTBoZk13NVNGT3QKdzIyRldDTjJ2WFBzWE1jYjF0WUFKYWt3bzFNd1VUQU9CZ05WSFE4QkFmOEVCQU1DQVFZd0R3WURWUjBUQVFILwpCQVV3QXdFQi96QWRCZ05WSFE0RUZnUVU4U3k5WmFvME5VbWY2WjZhRVo4d0kwZEJwYXN3RHdZRFZSMFJCQWd3CkJvY0Vmd0FBQVRBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlFQXNEOUFrcENEQnhKN2ZrSkhUK0kxWHZPSUVDSXAKYnNQenFOMTd5a1JNOGNFQ0lEL0RhYk5CeWRaUTEvRGJVdFZTUjVQSk1uS202b0dkSjMrWkQ0Vm43QWJqCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K",
    "hsm": false,
    "private_key": "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZ2RUdDdlUGtNcG0rODV1QkIKMFZqQzZ4RUlUekxJSlZDWmhNT3NMQ0ZhT3I2aFJBTkNBQVJTYUxaRGNOYzFvc3JtTjB2RHNPOWhrR1gxUGFSNQp0NzJYNitCV3hDaURNWWZyTGNYczlkOW9LNU9oQkJWWkJ0dEkwWVR6bjVpM3pydnFKOFc5cWw5RAotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
}`

var _ = Describe("the txid package", func() {

	var testIdentity *identity.Identity

	BeforeEach(func() {
		var err error
		testIdentity, err = identity.FromBytes([]byte(testIdentityText))
		Expect(err).NotTo(HaveOccurred())
	})

	Context("txid.New", func() {

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
