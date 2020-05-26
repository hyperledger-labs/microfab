/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// JoinChannel asks the peer to join the specified channel.
func (p *Peer) JoinChannel(block *common.Block) error {
	txID := txid.New(p.connMSPID, p.connIdentity)
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_CONFIG, "", txID)
	cche := &peer.ChaincodeHeaderExtension{
		ChaincodeId: &peer.ChaincodeID{
			Name: "cscc",
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
				Name: "cscc",
			},
			Input: &peer.ChaincodeInput{
				Args: [][]byte{
					[]byte("JoinChain"),
					util.MarshalOrPanic(block),
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

// ListChannels asks the peer for the list of channels it has joined.
func (p *Peer) ListChannels() ([]string, error) {
	txID := txid.New(p.connMSPID, p.connIdentity)
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_CONFIG, "", txID)
	cche := &peer.ChaincodeHeaderExtension{
		ChaincodeId: &peer.ChaincodeID{
			Name: "cscc",
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
				Name: "cscc",
			},
			Input: &peer.ChaincodeInput{
				Args: [][]byte{
					[]byte("GetChannels"),
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
		return nil, err
	} else if response.Response.Status != int32(common.Status_SUCCESS) {
		return nil, fmt.Errorf("Bad proposal response: status %d, mesage %s", response.Response.Status, response.Response.Message)
	}
	channelQueryResponse := &peer.ChannelQueryResponse{}
	err = proto.Unmarshal(response.Response.Payload, channelQueryResponse)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, channel := range channelQueryResponse.Channels {
		result = append(result, channel.ChannelId)
	}
	return result, nil
}
