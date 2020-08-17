/*
 * SPDX-License-Identifier: Apache-2.0
 */

package blocks

import (
	"errors"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

// DeliverCallback is a callback function called for every block returned by Deliver.
type DeliverCallback func(*common.Block) error

// Deliverer can be implemented by types that can deliver one or more blocks.
type Deliverer interface {
	MSPID() string
	Identity() *identity.Identity
	Deliver(envelope *common.Envelope, callback DeliverCallback) error
}

// GetConfigBlock gets the latest config block from the specified channel.
func GetConfigBlock(deliverer Deliverer, channel string) (*common.Block, error) {
	newestBlock, err := GetNewestBlock(deliverer, channel)
	if err != nil {
		return nil, err
	}
	metadataBytes := newestBlock.GetMetadata().GetMetadata()[common.BlockMetadataIndex_LAST_CONFIG]
	metadata := &common.Metadata{}
	util.UnmarshalOrPanic(metadataBytes, metadata)
	lastConfig := &common.LastConfig{}
	util.UnmarshalOrPanic(metadata.Value, lastConfig)
	return GetSpecificBlock(deliverer, channel, lastConfig.Index)
}

// GetNewestBlock gets the newest block from the specified channel.
func GetNewestBlock(deliverer Deliverer, channel string) (*common.Block, error) {
	seekInfo := &orderer.SeekInfo{
		Start: &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Newest{
				Newest: &orderer.SeekNewest{},
			},
		},
		Stop: &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Newest{
				Newest: &orderer.SeekNewest{},
			},
		},
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}
	return getBlock(deliverer, channel, seekInfo)
}

// GetSpecificBlock gets the specified block from the specified channel.
func GetSpecificBlock(deliverer Deliverer, channel string, number uint64) (*common.Block, error) {
	seekInfo := &orderer.SeekInfo{
		Start: &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: number,
				},
			},
		},
		Stop: &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: number,
				},
			},
		},
		Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}
	return getBlock(deliverer, channel, seekInfo)
}

// GetGenesisBlock gets the genesis block from the specified channel.
func GetGenesisBlock(deliverer Deliverer, channel string) (*common.Block, error) {
	return GetSpecificBlock(deliverer, channel, 0)
}

func buildEnvelope(deliverer Deliverer, channel string, seekInfo *orderer.SeekInfo) *common.Envelope {
	txID := txid.New(deliverer.MSPID(), deliverer.Identity())
	header := protoutil.BuildHeader(common.HeaderType_DELIVER_SEEK_INFO, channel, txID)
	payload := protoutil.BuildPayload(header, seekInfo)
	return protoutil.BuildEnvelope(payload, deliverer.Identity())
}

func getBlock(deliverer Deliverer, channel string, seekInfo *orderer.SeekInfo) (*common.Block, error) {
	envelope := buildEnvelope(deliverer, channel, seekInfo)
	var result *common.Block
	err := deliverer.Deliver(envelope, func(block *common.Block) error {
		if result != nil {
			return errors.New("Multiple blocks returned by seek info request")
		}
		result = block
		return nil
	})
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, errors.New("No blocks returned by seek info request")
	}
	return result, nil
}
