/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"google.golang.org/grpc"
)

// Connection represents a connection to a peer.
type Connection struct {
	peer       *Peer
	clientConn *grpc.ClientConn
	mspID      string
	identity   *identity.Identity
}

// Connect opens a connection to the peer.
func Connect(peer *Peer, mspID string, identity *identity.Identity) (*Connection, error) {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", peer.apiPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Connection{
		peer:       peer,
		clientConn: clientConn,
		mspID:      mspID,
		identity:   identity,
	}, nil
}

// IsConnected returns true if the connection is open to the peer.
func (c *Connection) IsConnected() bool {
	return c.clientConn != nil
}

// Close closes the connection to the peer.
func (c *Connection) Close() error {
	c.identity = nil
	if c.clientConn != nil {
		err := c.clientConn.Close()
		c.clientConn = nil
		return err
	}
	return nil
}

// MSPID gets the MSP ID used for the connection.
func (c *Connection) MSPID() string {
	return c.mspID
}

// Identity gets the identity used for the connection.
func (c *Connection) Identity() *identity.Identity {
	return c.identity
}
