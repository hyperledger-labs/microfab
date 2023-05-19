/*
Copyright IBM Corp. 2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger-labs/microfab/internal/pkg/blocks"
	"github.com/hyperledger-labs/microfab/internal/pkg/configtxlator"
	"github.com/hyperledger/fabric-protos-go/common"
)

// GetConfigEnvelope gets the latest config envelope from the specified channel.
func GetConfigEnvelope(deliverer blocks.Deliverer, channel string) (*common.ConfigEnvelope, error) {
	configBlock, err := blocks.GetConfigBlock(deliverer, channel)
	if err != nil {
		return nil, err
	}
	if len(configBlock.Data.Data) != 1 {
		return nil, fmt.Errorf("Config block must only contain one transaction")
	}
	envelope := &common.Envelope{}
	err = proto.Unmarshal(configBlock.Data.Data[0], envelope)
	if err != nil {
		return nil, err
	}
	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, err
	}
	configEnvelope := &common.ConfigEnvelope{}
	err = proto.Unmarshal(payload.Data, configEnvelope)
	if err != nil {
		return nil, err
	}
	return configEnvelope, nil
}

// GetConfig gets the latest config from the specified channel.
func GetConfig(deliverer blocks.Deliverer, channel string) (*common.Config, error) {
	configEnvelope, err := GetConfigEnvelope(deliverer, channel)
	if err != nil {
		return nil, err
	}
	return configEnvelope.GetConfig(), nil
}

// GenerateConfigUpdate generates a config update calculated by comparing the specified configs.
func GenerateConfigUpdate(originalConfig *common.Config, newConfig *common.Config) (*common.ConfigUpdate, error) {
	configUpdate, err := configtxlator.Compute(originalConfig, newConfig)
	if err != nil {
		return nil, err
	}
	return configUpdate, nil
}
