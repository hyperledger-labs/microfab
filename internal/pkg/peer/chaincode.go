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
	"github.com/hyperledger/fabric-protos-go/peer/lifecycle"
)

// InstallChaincode installs the specified chaincode package onto the peer.
func (c *Connection) InstallChaincode(pkg []byte) (string, error) {
	txID := txid.New(c.mspID, c.identity)
	channelHeader := protoutil.BuildChannelHeader(common.HeaderType_ENDORSER_TRANSACTION, "", txID)
	cche := &peer.ChaincodeHeaderExtension{
		ChaincodeId: &peer.ChaincodeID{
			Name: "_lifecycle",
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
				Name: "_lifecycle",
			},
			Input: &peer.ChaincodeInput{
				Args: [][]byte{
					[]byte("InstallChaincode"),
					util.MarshalOrPanic(&lifecycle.InstallChaincodeArgs{
						ChaincodeInstallPackage: pkg,
					}),
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
	signature := c.identity.Sign(proposalBytes)
	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	response, err := c.ProcessProposal(signedProposal)
	if err != nil {
		return "", err
	} else if response.Response.Status != int32(common.Status_SUCCESS) {
		return "", fmt.Errorf("Bad proposal response: status %d, mesage %s", response.Response.Status, response.Response.Message)
	}
	result := &lifecycle.InstallChaincodeResult{}
	util.UnmarshalOrPanic(response.Response.Payload, result)
	return result.PackageId, nil
}
