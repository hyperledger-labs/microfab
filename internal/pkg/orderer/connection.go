/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = log.New(os.Stdout, fmt.Sprintf("[%16s] ", "console"), log.LstdFlags)

// Connection represents a connection to a orderer.
type Connection struct {
	orderer    *Orderer
	clientConn *grpc.ClientConn
	mspID      string
	identity   *identity.Identity
}

// Connect opens a connection to the orderer.
func Connect(orderer *Orderer, mspID string, identity *identity.Identity) (*Connection, error) {
	var clientConn *grpc.ClientConn
	var err error
	if orderer.tls != nil {
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
		clientConn, err = grpc.Dial(fmt.Sprintf("localhost:%d", orderer.apiPort), grpc.WithTransportCredentials(creds))
	} else {
		clientConn, err = grpc.Dial(fmt.Sprintf("localhost:%d", orderer.apiPort), grpc.WithInsecure())
	}
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

// ConnectClient opens a connection to the orderer using a client orderer object.
func ConnectClient(orderer *client.OrderingService, mspID string, identity *identity.Identity, tlsEnabled bool) (*Connection, error) {
	parsedURL, err := url.Parse(orderer.APIURL)
	if err != nil {
		return nil, err
	}

	var clientConn *grpc.ClientConn

	if tlsEnabled {
		logger.Printf("Using TLS")
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})

		clientConn, err = grpc.Dial(parsedURL.Host, grpc.WithTransportCredentials(creds), grpc.WithAuthority(orderer.APIOptions.DefaultAuthority))
	} else {
		logger.Printf("Not Using TLS")
		clientConn, err = grpc.Dial(parsedURL.Host, grpc.WithInsecure(), grpc.WithAuthority(orderer.APIOptions.DefaultAuthority))
	}

	// clientConn, err := grpc.Dial(parsedURL.Host, grpc.WithInsecure(), grpc.WithAuthority(orderer.APIOptions.DefaultAuthority))
	if err != nil {
		return nil, err
	}
	return &Connection{
		orderer:    nil,
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
