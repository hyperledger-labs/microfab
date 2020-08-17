/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"
	"net/url"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/pkg/client"
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

// ConnectClient opens a connection to the peer using a client peer object.
func ConnectClient(peer *client.Peer, mspID string, identity *identity.Identity) (*Connection, error) {
	parsedURL, err := url.Parse(peer.APIURL)
	if err != nil {
		return nil, err
	}
	clientConn, err := grpc.Dial(parsedURL.Host, grpc.WithInsecure(), grpc.WithAuthority(peer.APIOptions.DefaultAuthority))
	if err != nil {
		return nil, err
	}
	return &Connection{
		peer:       nil,
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
