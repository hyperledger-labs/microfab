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

debug:
	go build -o microfabd cmd/microfabd/main.go
	MICROFAB_CONFIG='{"couchdb":false,"endorsing_organizations":[{"name": "org1" },{"name": "org2"}],"certificate_authorities":false}'	./microfabd