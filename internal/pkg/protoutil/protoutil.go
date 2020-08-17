/*
 * SPDX-License-Identifier: Apache-2.0
 */

package protoutil

import (
	"crypto/sha256"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// GenerateTimestamp generates a new timestamp with the current time.
func GenerateTimestamp() *timestamp.Timestamp {
	now := time.Now()
	seconds := now.Unix()
	nanos := int32(now.UnixNano() - (seconds * 1000000000))
	return &timestamp.Timestamp{
		Seconds: seconds,
		Nanos:   nanos,
	}
}

// BuildChannelHeader builds a channel header for the specified channel and transaction ID.
func BuildChannelHeader(headerType common.HeaderType, channel string, txID *txid.TransactionID) *common.ChannelHeader {
	timestamp := GenerateTimestamp()
	return &common.ChannelHeader{
		Type:        int32(headerType),
		Version:     1,
		ChannelId:   channel,
		TxId:        txID.String(),
		Timestamp:   timestamp,
		TlsCertHash: txID.Identity().Certificate().Hash(),
	}
}

// BuildSignatureHeader builds a signature header for the specified transaction ID.
func BuildSignatureHeader(txID *txid.TransactionID) *common.SignatureHeader {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   txID.MSPID(),
		IdBytes: txID.Identity().Certificate().Bytes(),
	}
	return &common.SignatureHeader{
		Creator: util.MarshalOrPanic(serializedIdentity),
		Nonce:   txID.Nonce(),
	}
}

// BuildHeader builds a header for the specified header type, channel, and transaction ID.
func BuildHeader(headerType common.HeaderType, channel string, txID *txid.TransactionID) *common.Header {
	channelHeader := BuildChannelHeader(headerType, channel, txID)
	signatureHeader := BuildSignatureHeader(txID)
	return &common.Header{
		ChannelHeader:   util.MarshalOrPanic(channelHeader),
		SignatureHeader: util.MarshalOrPanic(signatureHeader),
	}
}

// BuildPayload builds a payload for the specified header and data.
func BuildPayload(header *common.Header, data proto.Message) *common.Payload {
	return &common.Payload{
		Header: header,
		Data:   util.MarshalOrPanic(data),
	}
}

// BuildEnvelope builds an envelope for the specified payload and signs it.
func BuildEnvelope(payload *common.Payload, identity *identity.Identity) *common.Envelope {
	payloadBytes := util.MarshalOrPanic(payload)
	signature := identity.Sign(payloadBytes)
	return &common.Envelope{
		Payload:   payloadBytes,
		Signature: signature,
	}
}

// BuildGenesisBlock builds a genesis block containing the specified envelope.
func BuildGenesisBlock(envelope *common.Envelope) *common.Block {
	data := util.MarshalOrPanic(envelope)
	dataHash := sha256.Sum256(data)
	blockHeader := &common.BlockHeader{
		Number:       0,
		PreviousHash: nil,
		DataHash:     dataHash[:],
	}
	blockMetadata := &common.BlockMetadata{
		Metadata: [][]byte{
			{},
			{},
			{},
			{},
			{},
		},
	}
	blockData := &common.BlockData{
		Data: [][]byte{
			data,
		},
	}
	return &common.Block{
		Header:   blockHeader,
		Metadata: blockMetadata,
		Data:     blockData,
	}
}

// BuildImplicitMetaPolicy builds an implicit meta policy for the specified rule and sub policy.
func BuildImplicitMetaPolicy(rule common.ImplicitMetaPolicy_Rule, subPolicy string) *common.Policy {
	return &common.Policy{
		Type: int32(common.Policy_IMPLICIT_META),
		Value: util.MarshalOrPanic(&common.ImplicitMetaPolicy{
			Rule:      rule,
			SubPolicy: subPolicy,
		}),
	}
}

// BuildImplicitMetaConfigPolicy builds an implicit meta config policy for the specified rule and subpolicy.
func BuildImplicitMetaConfigPolicy(rule common.ImplicitMetaPolicy_Rule, subPolicy string) *common.ConfigPolicy {
	return &common.ConfigPolicy{
		ModPolicy: "Admins",
		Policy:    BuildImplicitMetaPolicy(rule, subPolicy),
	}
}

// BuildSignaturePolicyEnvelope builds a signature policy envelope for the specified MSP ID and role.
func BuildSignaturePolicyEnvelope(mspID string, role msp.MSPRole_MSPRoleType) *common.SignaturePolicyEnvelope {
	return &common.SignaturePolicyEnvelope{
		Identities: []*msp.MSPPrincipal{
			{
				PrincipalClassification: msp.MSPPrincipal_ROLE,
				Principal: util.MarshalOrPanic(&msp.MSPRole{
					MspIdentifier: mspID,
					Role:          role,
				}),
			},
		},
		Rule: &common.SignaturePolicy{
			Type: &common.SignaturePolicy_NOutOf_{
				NOutOf: &common.SignaturePolicy_NOutOf{
					N: 1,
					Rules: []*common.SignaturePolicy{
						{
							Type: &common.SignaturePolicy_SignedBy{
								SignedBy: 0,
							},
						},
					},
				},
			},
		},
	}
}

// BuildFabricCryptoConfig builds the default Fabric crypto configuration.
func BuildFabricCryptoConfig() *msp.FabricCryptoConfig {
	return &msp.FabricCryptoConfig{
		IdentityIdentifierHashFunction: "SHA256",
		SignatureHashFamily:            "SHA2",
	}
}

// BuildFabricMSPConfig builds the Fabric MSP configuration for an organization.
func BuildFabricMSPConfig(organization *organization.Organization) *msp.FabricMSPConfig {
	mspConfig := &msp.FabricMSPConfig{
		Name: organization.MSPID(),
		RootCerts: [][]byte{
			organization.CA().Certificate().Bytes(),
		},
		IntermediateCerts:    [][]byte{},
		TlsRootCerts:         [][]byte{},
		TlsIntermediateCerts: [][]byte{},
		Admins: [][]byte{
			organization.Admin().Certificate().Bytes(),
		},
		RevocationList:                [][]byte{},
		SigningIdentity:               nil,
		OrganizationalUnitIdentifiers: []*msp.FabricOUIdentifier{},
		CryptoConfig:                  BuildFabricCryptoConfig(),
		FabricNodeOus:                 BuildFabricNodeOUs(),
	}
	return mspConfig
}

// BuildFabricNodeOUs builds the default Fabric NodeOU configuration.
func BuildFabricNodeOUs() *msp.FabricNodeOUs {
	return &msp.FabricNodeOUs{
		Enable: true,
		AdminOuIdentifier: &msp.FabricOUIdentifier{
			OrganizationalUnitIdentifier: "admin",
		},
		ClientOuIdentifier: &msp.FabricOUIdentifier{
			OrganizationalUnitIdentifier: "client",
		},
		PeerOuIdentifier: &msp.FabricOUIdentifier{
			OrganizationalUnitIdentifier: "peer",
		},
		OrdererOuIdentifier: &msp.FabricOUIdentifier{
			OrganizationalUnitIdentifier: "orderer",
		},
	}
}

// BuildConfigGroupFromOrganization builds a config group from an organization.
func BuildConfigGroupFromOrganization(organization *organization.Organization) *common.ConfigGroup {
	signaturePolicyEnvelope := BuildSignaturePolicyEnvelope(organization.MSPID(), msp.MSPRole_MEMBER)
	configPolicy := &common.ConfigPolicy{
		ModPolicy: "Admins",
		Policy: &common.Policy{
			Type:  int32(common.Policy_SIGNATURE),
			Value: util.MarshalOrPanic(signaturePolicyEnvelope),
		},
	}
	fabricMSPConfig := BuildFabricMSPConfig(organization)
	configGroup := &common.ConfigGroup{
		Groups:    map[string]*common.ConfigGroup{},
		ModPolicy: "/Channel/Application/Admins",
		Policies: map[string]*common.ConfigPolicy{
			"Admins":      configPolicy,
			"Readers":     configPolicy,
			"Writers":     configPolicy,
			"Endorsement": configPolicy,
		},
		Values: map[string]*common.ConfigValue{
			"MSP": {
				ModPolicy: "Admins",
				Value: util.MarshalOrPanic(&msp.MSPConfig{
					Type:   0,
					Config: util.MarshalOrPanic(fabricMSPConfig),
				}),
			},
		},
	}
	return configGroup
}
