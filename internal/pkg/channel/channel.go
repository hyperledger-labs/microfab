/*
 * SPDX-License-Identifier: Apache-2.0
 */

package channel

import (
	"fmt"

	"github.com/IBM-Blockchain/microfab/internal/pkg/config"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/msp"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	fmsp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type channelOperation struct {
	config   *common.Config
	mspID    string
	identity *identity.Identity
}

// Option is a type representing an option for creating or updating a channel.
type Option func(*channelOperation) error

// AddMSPID adds the specified MSP ID to the channel.
func AddMSPID(mspID string) Option {
	return func(operation *channelOperation) error {
		operation.config.GetChannelGroup().Groups["Application"].Groups[mspID] = &common.ConfigGroup{}
		return nil
	}
}

// AddMSPIDs adds the specified MSP IDs to the channel.
func AddMSPIDs(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := AddMSPID(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// AddMSP adds the specified MSP to the channel.
func AddMSP(msp *msp.MSP) Option {
	return func(operation *channelOperation) error {
		rootCerts := [][]byte{}
		for _, rootCert := range msp.RootCertificates() {
			rootCerts = append(rootCerts, rootCert.Bytes())
		}
		tlsRootCerts := [][]byte{}
		adminCerts := [][]byte{}
		for _, adminCert := range msp.AdminCertificates() {
			adminCerts = append(adminCerts, adminCert.Bytes())
		}
		emptySignaturePolicy := &common.Policy{
			Type: int32(common.Policy_SIGNATURE),
			Value: util.MarshalOrPanic(&common.SignaturePolicyEnvelope{
				Identities: []*fmsp.MSPPrincipal{},
				Rule: &common.SignaturePolicy{
					Type: &common.SignaturePolicy_NOutOf_{
						NOutOf: &common.SignaturePolicy_NOutOf{
							N:     1,
							Rules: []*common.SignaturePolicy{},
						},
					},
				},
			}),
		}
		mspGroup := &common.ConfigGroup{
			Groups:    map[string]*common.ConfigGroup{},
			ModPolicy: "Admins",
			Policies: map[string]*common.ConfigPolicy{
				"Admins": {
					ModPolicy: "Admins",
					Policy:    proto.Clone(emptySignaturePolicy).(*common.Policy),
				},
				"Writers": {
					ModPolicy: "Admins",
					Policy:    proto.Clone(emptySignaturePolicy).(*common.Policy),
				},
				"Readers": {
					ModPolicy: "Admins",
					Policy:    proto.Clone(emptySignaturePolicy).(*common.Policy),
				},
			},
			Values: map[string]*common.ConfigValue{
				"MSP": {
					ModPolicy: "Admins",
					Value: util.MarshalOrPanic(&fmsp.MSPConfig{
						Config: util.MarshalOrPanic(&fmsp.FabricMSPConfig{
							Name: msp.ID(),
							CryptoConfig: &fmsp.FabricCryptoConfig{
								SignatureHashFamily:            "SHA2",
								IdentityIdentifierHashFunction: "SHA256",
							},
							RootCerts:    rootCerts,
							TlsRootCerts: tlsRootCerts,
							Admins:       adminCerts,
						}),
					}),
				},
			},
		}
		addToPolicy(mspGroup.Policies["Admins"].Policy, msp.ID(), fmsp.MSPRole_ADMIN)
		addToPolicy(mspGroup.Policies["Writers"].Policy, msp.ID(), fmsp.MSPRole_MEMBER)
		addToPolicy(mspGroup.Policies["Readers"].Policy, msp.ID(), fmsp.MSPRole_MEMBER)
		operation.config.GetChannelGroup().Groups["Application"].Groups[msp.ID()] = mspGroup
		return nil
	}
}

// AddMSPs adds the specified MSPs to the channel.
func AddMSPs(msps ...*msp.MSP) Option {
	return func(operation *channelOperation) error {
		for _, msp := range msps {
			err := AddMSP(msp)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// RemoveMSPID removes the specified MSP from the channel.
func RemoveMSPID(mspID string) Option {
	return func(operation *channelOperation) error {
		delete(operation.config.GetChannelGroup().Groups["Application"].Groups, mspID)
		return nil
	}
}

// RemoveMSPIDs removes the specified MSP IDs from the channel.
func RemoveMSPIDs(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := RemoveMSPID(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// AddAdmin adds the specified MSP ID to the admins policy of the channel.
func AddAdmin(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Admins"].GetPolicy()
		err := addToPolicy(policy, mspID, fmsp.MSPRole_ADMIN)
		if err != nil {
			return err
		}
		return nil
	}
}

// AddAdmins adds the specified MSP IDs to the admins policy of the channel.
func AddAdmins(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := AddAdmin(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// RemoveAdmin removes the specified MSP ID from the admins policy of the channel.
func RemoveAdmin(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Admins"].GetPolicy()
		err := removeFromPolicy(policy, mspID)
		if err != nil {
			return err
		}
		return nil
	}
}

// RemoveAdmins removes the specified MSP IDs from the admins policy of the channel.
func RemoveAdmins(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := RemoveAdmin(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// AddWriter adds the specified MSP ID to the writers policy of the channel.
func AddWriter(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Writers"].GetPolicy()
		err := addToPolicy(policy, mspID, fmsp.MSPRole_MEMBER)
		if err != nil {
			return err
		}
		return nil
	}
}

// AddWriters adds the specified MSP IDs to the writers policy of the channel.
func AddWriters(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := AddWriter(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// RemoveWriter removes the specified MSP ID from the writers policy of the channel.
func RemoveWriter(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Writers"].GetPolicy()
		err := removeFromPolicy(policy, mspID)
		if err != nil {
			return err
		}
		return nil
	}
}

// RemoveWriters removes the specified MSP IDs from the writers policy of the channel.
func RemoveWriters(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := RemoveWriter(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// AddReader adds the specified MSP ID to the readers policy of the channel.
func AddReader(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Readers"].GetPolicy()
		err := addToPolicy(policy, mspID, fmsp.MSPRole_MEMBER)
		if err != nil {
			return err
		}
		return nil
	}
}

// AddReaders adds the specified MSP IDs to the readers policy of the channel.
func AddReaders(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := AddReader(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// RemoveReader removes the specified MSP ID from the readers policy of the channel.
func RemoveReader(mspID string) Option {
	return func(operation *channelOperation) error {
		policy := operation.config.GetChannelGroup().Groups["Application"].Policies["Readers"].GetPolicy()
		err := removeFromPolicy(policy, mspID)
		if err != nil {
			return err
		}
		return nil
	}
}

// RemoveReaders removes the specified MSP IDs from the readers policy of the channel.
func RemoveReaders(mspIDs ...string) Option {
	return func(operation *channelOperation) error {
		for _, mspID := range mspIDs {
			err := RemoveReader(mspID)(operation)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// AddAnchorPeer adds the specified anchor peer to the channel.
func AddAnchorPeer(mspID string, hostname string, port int32) Option {
	return func(operation *channelOperation) error {
		msp, ok := operation.config.GetChannelGroup().Groups["Application"].Groups[mspID]
		if !ok {
			return fmt.Errorf("The channel does not contain an MSP with ID %s", mspID)
		}
		cv, ok := msp.Values["AnchorPeers"]
		if !ok {
			msp.Values["AnchorPeers"] = &common.ConfigValue{
				ModPolicy: "Admins",
				Value: util.MarshalOrPanic(&peer.AnchorPeers{
					AnchorPeers: []*peer.AnchorPeer{},
				}),
			}
			cv = msp.Values["AnchorPeers"]
		}
		aps := &peer.AnchorPeers{}
		proto.Unmarshal(cv.Value, aps)
		aps.AnchorPeers = append(aps.AnchorPeers, &peer.AnchorPeer{
			Host: hostname,
			Port: port,
		})
		cv.Value = util.MarshalOrPanic(aps)
		return nil
	}
}

// WithCapabilityLevel set the specified capability level for the channel.
func WithCapabilityLevel(capabilityLevel string) Option {
	return func(operation *channelOperation) error {
		operation.config.GetChannelGroup().Groups["Application"].Values["Capabilities"].Value = util.MarshalOrPanic(&common.Capabilities{
			Capabilities: map[string]*common.Capability{
				capabilityLevel: {},
			},
		})
		return nil
	}
}

// UsingMSPID uses the specified MSP ID to create or update the channel.
func UsingMSPID(mspID string) Option {
	return func(operation *channelOperation) error {
		operation.mspID = mspID
		return nil
	}
}

// UsingIdentity uses the specified identity to create or update the channel.
func UsingIdentity(identity *identity.Identity) Option {
	return func(operation *channelOperation) error {
		operation.identity = identity
		return nil
	}
}

// CreateChannel creates a new channel on the specified ordering service.
func CreateChannel(o *orderer.Connection, channel string, opts ...Option) error {
	_ = &common.Policy{
		Type: int32(common.Policy_SIGNATURE),
		Value: util.MarshalOrPanic(&common.SignaturePolicyEnvelope{
			Identities: []*fmsp.MSPPrincipal{},
			Rule: &common.SignaturePolicy{
				Type: &common.SignaturePolicy_NOutOf_{
					NOutOf: &common.SignaturePolicy_NOutOf{
						N:     1,
						Rules: []*common.SignaturePolicy{},
					},
				},
			},
		}),
	}
	configUpdate := &common.ConfigUpdate{
		ChannelId: channel,
		ReadSet: &common.ConfigGroup{
			Groups: map[string]*common.ConfigGroup{
				"Application": {
					Groups: map[string]*common.ConfigGroup{},
				},
			},
			Values: map[string]*common.ConfigValue{
				"Consortium": {
					Value: util.MarshalOrPanic(&common.Consortium{
						Name: "SampleConsortium",
					}),
				},
			},
		},
		WriteSet: &common.ConfigGroup{
			Groups: map[string]*common.ConfigGroup{
				"Application": {
					Groups:    map[string]*common.ConfigGroup{},
					ModPolicy: "Admins",
					Policies: map[string]*common.ConfigPolicy{
						"Admins":               protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Admins"),
						"Endorsement":          protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Endorsement"),
						"LifecycleEndorsement": protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Endorsement"),
						"Readers":              protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Readers"),
						"Writers":              protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Writers"),
					},
					Values: map[string]*common.ConfigValue{
						"Capabilities": {
							ModPolicy: "Admins",
							Value: util.MarshalOrPanic(&common.Capabilities{
								Capabilities: map[string]*common.Capability{
									"V2_0": {},
								},
							}),
						},
					},
					Version: 1,
				},
			},
			Values: map[string]*common.ConfigValue{
				"Consortium": {
					Value: util.MarshalOrPanic(&common.Consortium{
						Name: "SampleConsortium",
					}),
				},
			},
		},
	}
	operation := &channelOperation{
		&common.Config{
			ChannelGroup: configUpdate.WriteSet,
		},
		o.MSPID(),
		o.Identity(),
	}
	for _, opt := range opts {
		err := opt(operation)
		if err != nil {
			return err
		}
	}
	for mspID := range configUpdate.WriteSet.Groups["Application"].Groups {
		configUpdate.ReadSet.Groups["Application"].Groups[mspID] = &common.ConfigGroup{}
	}
	return createOrUpdateChannel(o, operation.mspID, operation.identity, configUpdate)
}

// UpdateChannel updates an existing channel on the specified ordering service.
func UpdateChannel(o *orderer.Connection, channel string, opts ...Option) error {
	originalConfig, err := config.GetConfig(o, channel)
	if err != nil {
		return err
	}
	newConfig := proto.Clone(originalConfig).(*common.Config)
	operation := &channelOperation{
		newConfig,
		o.MSPID(),
		o.Identity(),
	}
	for _, opt := range opts {
		err := opt(operation)
		if err != nil {
			return err
		}
	}
	configUpdate, err := config.GenerateConfigUpdate(originalConfig, newConfig)
	if err != nil {
		return err
	}
	configUpdate.ChannelId = channel
	return createOrUpdateChannel(o, operation.mspID, operation.identity, configUpdate)
}

func createOrUpdateChannel(o *orderer.Connection, mspID string, identity *identity.Identity, configUpdate *common.ConfigUpdate) error {
	txID := txid.New(mspID, identity)
	header := protoutil.BuildHeader(common.HeaderType_CONFIG_UPDATE, configUpdate.ChannelId, txID)
	configUpdateBytes := util.MarshalOrPanic(configUpdate)
	signature := identity.Sign(header.SignatureHeader, configUpdateBytes)
	configUpdateEnvelope := &common.ConfigUpdateEnvelope{
		ConfigUpdate: configUpdateBytes,
		Signatures: []*common.ConfigSignature{
			{
				SignatureHeader: header.SignatureHeader,
				Signature:       signature,
			},
		},
	}
	payload := protoutil.BuildPayload(header, configUpdateEnvelope)
	envelope := protoutil.BuildEnvelope(payload, txID)
	err := o.Broadcast(envelope)
	if err != nil {
		return err
	}
	return nil
}

func addToPolicy(policy *common.Policy, mspID string, role fmsp.MSPRole_MSPRoleType) error {
	spe := &common.SignaturePolicyEnvelope{}
	err := proto.Unmarshal(policy.Value, spe)
	if err != nil {
		return err
	}
	idx := len(spe.Identities)
	spe.Identities = append(spe.Identities, &fmsp.MSPPrincipal{
		PrincipalClassification: fmsp.MSPPrincipal_ROLE,
		Principal: util.MarshalOrPanic(&fmsp.MSPRole{
			MspIdentifier: mspID,
			Role:          role,
		}),
	})
	spe.Rule.GetNOutOf().Rules = append(spe.Rule.GetNOutOf().Rules, &common.SignaturePolicy{
		Type: &common.SignaturePolicy_SignedBy{
			SignedBy: int32(idx),
		},
	})
	policy.Value = util.MarshalOrPanic(spe)
	return nil
}

func removeFromPolicy(policy *common.Policy, mspID string) error {
	spe := &common.SignaturePolicyEnvelope{}
	err := proto.Unmarshal(policy.Value, spe)
	if err != nil {
		return err
	}
	identities := []*fmsp.MSPPrincipal{}
	rules := []*common.SignaturePolicy{}
	for _, identity := range spe.Identities {
		role := &fmsp.MSPRole{}
		err := proto.Unmarshal(identity.Principal, role)
		if err != nil {
			return err
		}
		if role.MspIdentifier != mspID {
			idx := len(identities)
			identities = append(identities, identity)
			rules = append(rules, &common.SignaturePolicy{
				Type: &common.SignaturePolicy_SignedBy{
					SignedBy: int32(idx),
				},
			})
		}
	}
	spe.Identities = identities
	spe.Rule.GetNOutOf().Rules = rules
	policy.Value = util.MarshalOrPanic(spe)
	return nil
}
