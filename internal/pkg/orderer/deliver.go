/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/blocks"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

// Deliver requests one or more blocks from the orderer.
func (c *Connection) Deliver(envelope *common.Envelope, callback blocks.DeliverCallback) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	deliverClient, err := orderer.NewAtomicBroadcastClient(c.clientConn).Deliver(ctx)
	if err != nil {
		return err
	}
	eof := make(chan bool)
	responses := make(chan *orderer.DeliverResponse)
	errors := make(chan error)
	go func() {
		for {
			response, err := deliverClient.Recv()
			if err == io.EOF {
				eof <- true
				return
			} else if err != nil {
				errors <- err
			} else {
				responses <- response
			}
		}
	}()
	err = deliverClient.Send(envelope)
	if err != nil {
		return err
	}
	err = deliverClient.CloseSend()
	if err != nil {
		return err
	}
	done := false
	for !done {
		select {
		case err = <-errors:
			return err
		case response := <-responses:
			switch response.GetType().(type) {
			case *orderer.DeliverResponse_Status:
				if response.GetStatus() != common.Status_SUCCESS {
					return fmt.Errorf("Bad status returned by orderer: %v", response.GetStatus())
				}
			case *orderer.DeliverResponse_Block:
				block := response.GetBlock()
				err = callback(block)
				if err != nil {
					return err
				}
			}
		case <-eof:
			done = true
		}
	}
	return nil
}
