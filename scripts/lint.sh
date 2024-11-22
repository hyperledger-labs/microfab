#!/usr/bin/env bash
#
# SPDX-License-Identifier: Apache-2.0
#
set -euxo pipefail
GOFMT="$(gofmt -l -s .)"
test -z "${GOFMT}"
go run golang.org/x/lint/golint -set_exit_status ./...
GORETURNS="$(go run github.com/sqs/goreturns -l cmd internal)"
test -z "${GORETURNS}"
go vet ./...
