/*
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import "github.com/IBM-Blockchain/microfab/internal/pkg/identity"

// Node represents a connection to a node.
type Node interface {
	Connect(string, *identity.Identity) error
	Connected() bool
	ConnectionMSPID() string
	ConnectionIdentity() *identity.Identity
	Close() error
	Host() string
	Hostname() string
	Port() int32
}
