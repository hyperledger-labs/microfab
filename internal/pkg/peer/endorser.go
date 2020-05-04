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
func (p *Peer) ProcessProposal(signedProposal *peer.SignedProposal) (*peer.ProposalResponse, error) {
	endorserClient := peer.NewEndorserClient(p.clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := endorserClient.ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, err
	}
	return response, nil
}
