//go:build tools
// +build tools

/*
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	_ "github.com/go-task/slim-sprig"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/sqs/goreturns"
	_ "golang.org/x/lint/golint"
)
