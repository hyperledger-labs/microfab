/*
 * SPDX-License-Identifier: Apache-2.0
 */

package txid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// TransactionID represents a generated transaction ID.
type TransactionID struct {
	mspID    string
	identity *identity.Identity
	txID     string
	nonce    []byte
}

// New creates a new transaction ID for the specified MSP ID and identity.
func New(mspID string, identity *identity.Identity) *TransactionID {
	nonce := make([]byte, 24)
	_, err := rand.Read(nonce)
	if err != nil {
		panic(err)
	}
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   mspID,
		IdBytes: identity.Certificate().Bytes(),
	}
	sha := sha256.New()
	sha.Write(nonce)
	sha.Write(util.MarshalOrPanic(serializedIdentity))
	hash := sha.Sum(nil)
	txID := hex.EncodeToString(hash)
	return &TransactionID{mspID, identity, txID, nonce}
}

// MSPID returns the MSP ID used to create this transaction ID.
func (t *TransactionID) MSPID() string {
	return t.mspID
}

// Identity returns the identity used to create this transaction ID.
func (t *TransactionID) Identity() *identity.Identity {
	return t.identity
}

// String returns the transaction ID as a string.
func (t *TransactionID) String() string {
	return t.txID
}

// Nonce returns the nonce used to create this transaction ID.
func (t *TransactionID) Nonce() []byte {
	return t.nonce
}
