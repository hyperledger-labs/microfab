/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

// Broadcast sends an envelope with one or more transactions in to the orderer.
func (o *Orderer) Broadcast(envelope *common.Envelope) error {
	abClient := orderer.NewAtomicBroadcastClient(o.clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	broadcastClient, err := abClient.Broadcast(ctx)
	if err != nil {
		return err
	}
	eof := make(chan bool)
	responses := make(chan *orderer.BroadcastResponse)
	errors := make(chan error)
	go func() {
		for {
			response, err := broadcastClient.Recv()
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
	err = broadcastClient.Send(envelope)
	if err != nil {
		return err
	}
	err = broadcastClient.CloseSend()
	if err != nil {
		return err
	}
	done := false
	for !done {
		select {
		case err = <-errors:
			return err
		case response := <-responses:
			if response.GetStatus() != common.Status_SUCCESS {
				return fmt.Errorf("Bad status returned by ordering service %v", response.GetStatus())
			}
		case <-eof:
			done = true
		}
	}
	return nil
}
