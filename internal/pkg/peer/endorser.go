/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"context"
	"time"

	"github.com/hyperledger/fabric-protos-go/peer"
)

// ProcessProposal asks the peer to endorse the specified proposal.
func (c *Connection) ProcessProposal(signedProposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
	endorserClient := peer.NewEndorserClient(c.clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := endorserClient.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, err
	}
	return response, nil
}
