#
# SPDX-License-Identifier: Apache-2.0
#

.PHONY: all lint unit integration

all: lint unit

generate:
	go generate ./...

lint:
	./scripts/lint.sh

unit:
	go run github.com/onsi/ginkgo/ginkgo -skipPackage integration ./...

integration:
	go run github.com/onsi/ginkgo/ginkgo integration
