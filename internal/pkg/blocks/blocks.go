/*
 * SPDX-License-Identifier: Apache-2.0
 */

package blocks

import (
	"fmt"

	"github.com/IBM-Blockchain/fablet/internal/pkg/node"
	"github.com/IBM-Blockchain/fablet/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/fablet/internal/pkg/txid"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// DeliverCallback is a callback function called for every block returned by Deliver.
type DeliverCallback func(*common.Block) error

// DeliverFilteredCallback is a callback function called for every filtered block returned by DeliverFiltered.
type DeliverFilteredCallback func(*peer.FilteredBlock) error

// Deliverer can be implemented by types that can deliver one or more blocks.
type Deliverer interface {
	node.Node
	Deliver(envelope *common.Envelope, callback DeliverCallback) error
}

// FilteredDeliverer can be implemented by types that can deliver one or more blocks.
type FilteredDeliverer interface {
	node.Node
	DeliverFiltered(envelope *common.Envelope, callback DeliverFilteredCallback) error
}

// GetConfigBlock gets the latest config block from the specified channel.
func GetConfigBlock(deliverer Deliverer, channel string) (*common.Block, error) {
	newestBlock, err := GetNewestBlock(deliverer, channel)
	if err != nil {
		return nil, err
	}
	metadataBytes := newestBlock.GetMetadata().GetMetadata()[common.BlockMetadataIndex_LAST_CONFIG]
	metadata := &common.Metadata{}
	err = proto.Unmarshal(metadataBytes, metadata)
	if err != nil {
		return nil, err
	}
	lastConfig := &common.LastConfig{}
	err = proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		return nil, err
	}
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
		Behavior: orderer.SeekInfo_FAIL_IF_NOT_READY,
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
		Behavior: orderer.SeekInfo_FAIL_IF_NOT_READY,
	}
	return getBlock(deliverer, channel, seekInfo)
}

// GetGenesisBlock gets the genesis block from the specified channel.
func GetGenesisBlock(deliverer Deliverer, channel string) (*common.Block, error) {
	return GetSpecificBlock(deliverer, channel, 0)
}

func buildEnvelope(deliverer Deliverer, channel string, seekInfo *orderer.SeekInfo) *common.Envelope {
	txID := txid.New(deliverer.ConnectionMSPID(), deliverer.ConnectionIdentity())
	header := protoutil.BuildHeader(common.HeaderType_DELIVER_SEEK_INFO, channel, txID)
	payload := protoutil.BuildPayload(header, seekInfo)
	return protoutil.BuildEnvelope(payload, txID)
}

func getBlocks(deliverer Deliverer, channel string, seekInfo *orderer.SeekInfo) ([]*common.Block, error) {
	envelope := buildEnvelope(deliverer, channel, seekInfo)
	result := make([]*common.Block, 0)
	err := deliverer.Deliver(envelope, func(block *common.Block) error {
		result = append(result, block)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getBlock(deliverer Deliverer, channel string, seekInfo *orderer.SeekInfo) (*common.Block, error) {
	envelope := buildEnvelope(deliverer, channel, seekInfo)
	var result *common.Block
	err := deliverer.Deliver(envelope, func(block *common.Block) error {
		if result != nil {
			return fmt.Errorf("No blocks or too many blocks returned by seek info request")
		}
		result = block
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
