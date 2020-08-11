/*
 * SPDX-License-Identifier: Apache-2.0
 */

package organization_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOrganization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Organization Suite")
}
