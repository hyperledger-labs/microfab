/*
 * SPDX-License-Identifier: Apache-2.0
 */

package channel

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/blocks"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	fpeer "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

// ApproveChaincodeDefinition approves a chaincode definition on a channel.
func ApproveChaincodeDefinition(peers []*peer.Connection, o *orderer.Connection, channel string, sequence int64, name string, version string, packageID string) error {
	arg := &lifecycle.ApproveChaincodeDefinitionForMyOrgArgs{
		Sequence: sequence,
		Name:     name,
		Version:  version,
		Source: &lifecycle.ChaincodeSource{
			Type: &lifecycle.ChaincodeSource_LocalPackage{
				LocalPackage: &lifecycle.ChaincodeSource_Local{
					PackageId: packageID,
				},
			},
		},
	}
	proposal, responses, endorsements, err := executeTransaction(peers, o, channel, "_lifecycle", "ApproveChaincodeDefinitionForMyOrg", util.MarshalOrPanic(arg))
	if err != nil {
		return err
	}
	err = orderTransaction(peers, o, channel, proposal, responses, endorsements)
	if err != nil {
		return err
	}
	return nil
}

// CommitChaincodeDefinition commits a chaincode definition on a channel.
func CommitChaincodeDefinition(peers []*peer.Connection, o *orderer.Connection, channel string, sequence int64, name string, version string) error {
	arg := &lifecycle.CommitChaincodeDefinitionArgs{
		Sequence: sequence,
		Name:     name,
		Version:  version,
	}
	proposal, responses, endorsements, err := executeTransaction(peers, o, channel, "_lifecycle", "CommitChaincodeDefinition", util.MarshalOrPanic(arg))
	if err != nil {
		return err
	}
	err = orderTransaction(peers, o, channel, proposal, responses, endorsements)
	if err != nil {
		return err
	}
	return nil
}

// EvaluateTransaction evaluates a transaction for a chaincode definition on a channel.
func EvaluateTransaction(peers []*peer.Connection, o *orderer.Connection, channel, chaincode, function string, args ...string) ([]byte, error) {
	byteArgs := [][]byte{}
	for _, arg := range args {
		byteArgs = append(byteArgs, []byte(arg))
	}
	_, responses, _, err := executeTransaction(peers, o, channel, chaincode, function, byteArgs...)
	if err != nil {
		return nil, err
	}
	return responses[0].Response.Payload, nil
}

// SubmitTransaction submits a transaction for a chaincode definition on a channel.
func SubmitTransaction(peers []*peer.Connection, o *orderer.Connection, channel, chaincode, function string, args ...string) ([]byte, error) {
	byteArgs := [][]byte{}
	for _, arg := range args {
		byteArgs = append(byteArgs, []byte(arg))
	}
	proposal, responses, endorsements, err := executeTransaction(peers, o, channel, chaincode, function, byteArgs...)
	if err != nil {
		return nil, err
	}
	err = orderTransaction(peers, o, channel, proposal, responses, endorsements)
	if err != nil {
		return nil, err
	}
	return responses[0].Response.Payload, nil
}

func executeTransaction(peers []*peer.Connection, o *orderer.Connection, channel, chaincode, function string, args ...[]byte) (*fpeer.Proposal, []*fpeer.ProposalResponse, []*fpeer.Endorsement, error) {
	firstPeer := peers[0]
	txID := txid.New(firstPeer.MSPID(), firstPeer.Identity())
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, channel, txID)
	cche := &fpeer.ChaincodeHeaderExtension{
		ChaincodeId: &fpeer.ChaincodeID{
			Name: chaincode,
		},
	}
	channelHeader.Extension = util.MarshalOrPanic(cche)
	signatureHeader := protoutil.BuildSignatureHeader(txID)
	header := &common.Header{
		ChannelHeader:   util.MarshalOrPanic(channelHeader),
		SignatureHeader: util.MarshalOrPanic(signatureHeader),
	}
	cciSpec := &fpeer.ChaincodeInvocationSpec{
		ChaincodeSpec: &fpeer.ChaincodeSpec{
			Type: fpeer.ChaincodeSpec_GOLANG,
			ChaincodeId: &fpeer.ChaincodeID{
				Name: chaincode,
			},
			Input: &fpeer.ChaincodeInput{
				Args: append([][]byte{[]byte(function)}, args...),
			},
		},
	}
	ccpp := &fpeer.ChaincodeProposalPayload{
		Input: util.MarshalOrPanic(cciSpec),
	}
	proposal := &fpeer.Proposal{
		Header:  util.MarshalOrPanic(header),
		Payload: util.MarshalOrPanic(ccpp),
	}
	proposalBytes := util.MarshalOrPanic(proposal)
	signature := firstPeer.Identity().Sign(proposalBytes)
	signedProposal := &fpeer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	responses := []*fpeer.ProposalResponse{}
	endorsements := []*fpeer.Endorsement{}
	for _, peer := range peers {
		response, err := peer.ProcessProposal(signedProposal)
		if err != nil {
			return nil, nil, nil, err
		} else if response.Response.Status != int32(common.Status_SUCCESS) {
			return nil, nil, nil, fmt.Errorf("Bad proposal response: status %d, mesage %s", response.Response.Status, response.Response.Message)
		}
		responses = append(responses, response)
		endorsements = append(endorsements, response.Endorsement)
	}
	return proposal, responses, endorsements, nil
}

func orderTransaction(peers []*peer.Connection, o *orderer.Connection, channel string, proposal *fpeer.Proposal, responses []*fpeer.ProposalResponse, endorsements []*fpeer.Endorsement) error {
	firstPeer := peers[0]
	header := &common.Header{}
	util.UnmarshalOrPanic(proposal.Header, header)
	channelHeader := &common.ChannelHeader{}
	util.UnmarshalOrPanic(header.ChannelHeader, channelHeader)
	txID := channelHeader.TxId
	chaincodeEndorsedAction := &fpeer.ChaincodeEndorsedAction{
		ProposalResponsePayload: responses[0].Payload,
		Endorsements:            endorsements,
	}
	chaincodeActionPayload := &fpeer.ChaincodeActionPayload{
		ChaincodeProposalPayload: proposal.Payload,
		Action:                   chaincodeEndorsedAction,
	}
	transactionAction := &fpeer.TransactionAction{
		Header:  header.SignatureHeader,
		Payload: util.MarshalOrPanic(chaincodeActionPayload),
	}
	transaction := &fpeer.Transaction{
		Actions: []*fpeer.TransactionAction{
			transactionAction,
		},
	}
	payload := protoutil.BuildPayload(header, transaction)
	envelope := protoutil.BuildEnvelope(payload, o.Identity())
	currentBlock, err := blocks.GetNewestBlock(firstPeer, channel)
	if err != nil {
		return err
	}
	nextBlockNumber := currentBlock.Header.Number + 1
	done := make(chan error)
	go func() {
		for {
			nextBlock, err := blocks.GetSpecificBlock(firstPeer, channel, nextBlockNumber)
			if err != nil {
				done <- err
				return
			}
			for _, data := range nextBlock.Data.Data {
				envelope := &common.Envelope{}
				util.UnmarshalOrPanic(data, envelope)
				payload := &common.Payload{}
				util.UnmarshalOrPanic(envelope.Payload, payload)
				channelHeader := &common.ChannelHeader{}
				util.UnmarshalOrPanic(payload.Header.ChannelHeader, channelHeader)
				if channelHeader.TxId == txID {
					done <- nil
					return
				}
			}
			nextBlockNumber++
		}
	}()
	if err := o.Broadcast(envelope); err != nil {
		return err
	}
	err = <-done
	return err
}
