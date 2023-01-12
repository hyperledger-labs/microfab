/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/hyperledger-labs/microfab/internal/pkg/identity"
	"github.com/hyperledger-labs/microfab/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	var clientConn *grpc.ClientConn
	var err error
	if peer.tls != nil {
		logger.Println("Peer TLS Enabled")
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		clientConn, err = grpc.Dial(fmt.Sprintf("localhost:%d", peer.apiPort), grpc.WithTransportCredentials(creds))
	} else {
		logger.Println("Peer _not_ TLS Enabled")
		clientConn, err = grpc.Dial(fmt.Sprintf("localhost:%d", peer.apiPort), grpc.WithInsecure())
	}
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
func ConnectClient(peer *client.Peer, mspID string, identity *identity.Identity, tlsEnabled bool) (*Connection, error) {
	parsedURL, err := url.Parse(peer.APIURL)
	if err != nil {
		return nil, err
	}
	logger.Printf("ConnectionClient Peer has parsedURL %s\n", parsedURL)

	var clientConn *grpc.ClientConn

	if tlsEnabled {
		logger.Printf("Using TLS")
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})

		clientConn, err = grpc.Dial(parsedURL.Host, grpc.WithTransportCredentials(creds), grpc.WithAuthority(peer.APIOptions.DefaultAuthority))
	} else {
		logger.Printf("Not Using TLS")
		clientConn, err = grpc.Dial(parsedURL.Host, grpc.WithInsecure(), grpc.WithAuthority(peer.APIOptions.DefaultAuthority))
	}

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
