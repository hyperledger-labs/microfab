/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOrderer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orderer Suite")
}
