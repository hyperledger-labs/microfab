//go:build tools
// +build tools

/*
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "golang.org/x/lint/golint"
	_ "sourcegraph.com/sqs/goreturns"
)
