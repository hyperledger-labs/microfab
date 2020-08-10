/*
 * SPDX-License-Identifier: Apache-2.0
 */

package txid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTxid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TxID Suite")
}
