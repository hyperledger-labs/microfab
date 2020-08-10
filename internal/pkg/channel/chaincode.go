/*
 * SPDX-License-Identifier: Apache-2.0
 */

package channel

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	fpeer "github.com/hyperledger/fabric-protos-go/peer"
)

// InstantiateChaincode instantiates a chaincode on the channel.
func InstantiateChaincode(peers []*peer.Connection, o *orderer.Orderer, channel string, name string, version string, args [][]byte) error {
	return instantiateOrUpgradeChaincode(peers, o, channel, name, version, args, "deploy")
}

// UpgradeChaincode upgrades a chaincode on the channel.
func UpgradeChaincode(peers []*peer.Connection, o *orderer.Orderer, channel string, name string, version string, args [][]byte) error {
	return instantiateOrUpgradeChaincode(peers, o, channel, name, version, args, "upgrade")
}

func instantiateOrUpgradeChaincode(peers []*peer.Connection, o *orderer.Orderer, channel string, name string, version string, args [][]byte, lsccFunction string) error {
	firstPeer := peers[0]
	txID := txid.New(firstPeer.MSPID(), firstPeer.Identity())
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, channel, txID)
	cche := &fpeer.ChaincodeHeaderExtension{
		ChaincodeId: &fpeer.ChaincodeID{
			Name: "lscc",
		},
	}
	channelHeader.Extension = util.MarshalOrPanic(cche)
	signatureHeader := protoutil.BuildSignatureHeader(txID)
	header := &common.Header{
		ChannelHeader:   util.MarshalOrPanic(channelHeader),
		SignatureHeader: util.MarshalOrPanic(signatureHeader),
	}
	cdsSpec := &fpeer.ChaincodeDeploymentSpec{
		ChaincodeSpec: &fpeer.ChaincodeSpec{
			Type: fpeer.ChaincodeSpec_GOLANG,
			ChaincodeId: &fpeer.ChaincodeID{
				Name:    name,
				Version: version,
			},
			Input: &fpeer.ChaincodeInput{
				Args: args,
			},
		},
	}
	cciSpec := &fpeer.ChaincodeInvocationSpec{
		ChaincodeSpec: &fpeer.ChaincodeSpec{
			Type: fpeer.ChaincodeSpec_GOLANG,
			ChaincodeId: &fpeer.ChaincodeID{
				Name: "lscc",
			},
			Input: &fpeer.ChaincodeInput{
				Args: [][]byte{
					[]byte(lsccFunction),
					[]byte(channel),
					util.MarshalOrPanic(cdsSpec),
					{}, // Endorsement policy
					[]byte("escc"),
					[]byte("vscc"),
				},
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
			return err
		} else if response.Response.Status != int32(common.Status_SUCCESS) {
			return fmt.Errorf("Bad proposal response: status %d, mesage %s", response.Response.Status, response.Response.Message)
		}
		responses = append(responses, response)
		endorsements = append(endorsements, response.Endorsement)
	}
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
	envelope := protoutil.BuildEnvelope(payload, txID)
	return o.Broadcast(envelope)
}
