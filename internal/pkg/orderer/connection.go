/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"google.golang.org/grpc"
)

// Connection represents a connection to a orderer.
type Connection struct {
	orderer    *Orderer
	clientConn *grpc.ClientConn
	mspID      string
	identity   *identity.Identity
}

// Connect opens a connection to the orderer.
func Connect(orderer *Orderer, mspID string, identity *identity.Identity) (*Connection, error) {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", orderer.apiPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Connection{
		orderer:    orderer,
		clientConn: clientConn,
		mspID:      mspID,
		identity:   identity,
	}, nil
}

// IsConnected returns true if the connection is open to the orderer.
func (c *Connection) IsConnected() bool {
	return c.clientConn != nil
}

// Close closes the connection to the orderer.
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
