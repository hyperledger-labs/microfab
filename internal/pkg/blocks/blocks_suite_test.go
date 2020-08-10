/*
 * SPDX-License-Identifier: Apache-2.0
 */

package blocks_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes/deliverer.go --fake-name Deliverer . Deliverer

func TestBlocks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blocks Suite")
}
