/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// InstallChaincode installs the specified chaincode package onto the peer.
func (p *Peer) InstallChaincode(scds []byte) error {
	txID := txid.New(p.connMSPID, p.connIdentity)
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, "", txID)
	cche := &peer.ChaincodeHeaderExtension{
		ChaincodeId: &peer.ChaincodeID{
			Name: "lscc",
		},
	}
	channelHeader.Extension = util.MarshalOrPanic(cche)
	signatureHeader := protoutil.BuildSignatureHeader(txID)
	header := &common.Header{
		ChannelHeader:   util.MarshalOrPanic(channelHeader),
		SignatureHeader: util.MarshalOrPanic(signatureHeader),
	}
	cciSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type: peer.ChaincodeSpec_GOLANG,
			ChaincodeId: &peer.ChaincodeID{
				Name: "lscc",
			},
			Input: &peer.ChaincodeInput{
				Args: [][]byte{
					[]byte("install"),
					scds,
				},
			},
		},
	}
	ccpp := &peer.ChaincodeProposalPayload{
		Input: util.MarshalOrPanic(cciSpec),
	}
	proposal := &peer.Proposal{
		Header:  util.MarshalOrPanic(header),
		Payload: util.MarshalOrPanic(ccpp),
	}
	proposalBytes := util.MarshalOrPanic(proposal)
	signature := p.connIdentity.Sign(proposalBytes)
	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	response, err := p.ProcessProposal(signedProposal)
	if err != nil {
		return err
	} else if response.Response.Status != int32(common.Status_SUCCESS) {
		return fmt.Errorf("Bad proposal response: status %d, mesage %s", response.Response.Status, response.Response.Message)
	}
	return nil
}
