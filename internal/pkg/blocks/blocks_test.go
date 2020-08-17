/*
 * SPDX-License-Identifier: Apache-2.0
 */

package blocks_test

import (
	"errors"

	"github.com/IBM-Blockchain/microfab/internal/pkg/blocks"
	"github.com/IBM-Blockchain/microfab/internal/pkg/blocks/fakes"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/orderer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getPayload(envelope *common.Envelope) *common.Payload {
	payload := &common.Payload{}
	util.UnmarshalOrPanic(envelope.Payload, payload)
	return payload
}

func getHeader(envelope *common.Envelope) *common.Header {
	payload := getPayload(envelope)
	return payload.Header
}

func getChannelHeader(envelope *common.Envelope) *common.ChannelHeader {
	header := getHeader(envelope)
	channelHeader := &common.ChannelHeader{}
	util.UnmarshalOrPanic(header.ChannelHeader, channelHeader)
	return channelHeader
}

func getSignatureHeader(envelope *common.Envelope) *common.SignatureHeader {
	header := getHeader(envelope)
	signatureHeader := &common.SignatureHeader{}
	util.UnmarshalOrPanic(header.SignatureHeader, signatureHeader)
	return signatureHeader
}

func getCreator(envelope *common.Envelope) *msp.SerializedIdentity {
	signatureHeader := getSignatureHeader(envelope)
	creator := &msp.SerializedIdentity{}
	util.UnmarshalOrPanic(signatureHeader.Creator, creator)
	return creator
}

func getSeekInfo(envelope *common.Envelope) *orderer.SeekInfo {
	payload := getPayload(envelope)
	seekInfo := &orderer.SeekInfo{}
	util.UnmarshalOrPanic(payload.Data, seekInfo)
	return seekInfo
}

var _ = Describe("the blocks package", func() {

	var testIdentity *identity.Identity
	var fakeBlock *common.Block
	var fakeDeliverer *fakes.Deliverer

	BeforeEach(func() {
		var err error
		testIdentity, err = identity.New("Org1Admin")
		Expect(err).NotTo(HaveOccurred())
		fakeBlock = &common.Block{}
		fakeDeliverer = &fakes.Deliverer{}
		fakeDeliverer.MSPIDReturns("Org1MSP")
		fakeDeliverer.IdentityReturns(testIdentity)
		fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
			return callback(fakeBlock)
		})
	})

	Context("blocks.GetConfigBlock()", func() {

		var fakeLatestBlock *common.Block
		var fakeConfigBlock *common.Block

		BeforeEach(func() {
			firstBlock := true
			fakeLatestBlock = &common.Block{
				Metadata: &common.BlockMetadata{
					Metadata: [][]byte{
						{},
						util.MarshalOrPanic(&common.Metadata{
							Value: util.MarshalOrPanic(&common.LastConfig{
								Index: 1337,
							}),
						}),
						{},
						{},
						{},
					},
				},
			}
			fakeConfigBlock = &common.Block{}
			fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
				if firstBlock {
					firstBlock = false
					return callback(fakeLatestBlock)
				}
				return callback(fakeConfigBlock)
			})
		})

		When("called for a channel", func() {
			It("returns the config block", func() {
				block, err := blocks.GetConfigBlock(fakeDeliverer, "mychannel")
				Expect(err).NotTo(HaveOccurred())
				Expect(block).To(Equal(fakeConfigBlock))
				Expect(fakeDeliverer.DeliverCallCount()).To(Equal(2))
				envelope, _ := fakeDeliverer.DeliverArgsForCall(0)
				channelHeader := getChannelHeader(envelope)
				Expect(channelHeader.Type).To(BeEquivalentTo(common.HeaderType_DELIVER_SEEK_INFO))
				Expect(channelHeader.ChannelId).To(Equal("mychannel"))
				creator := getCreator(envelope)
				Expect(creator.Mspid).To(Equal("Org1MSP"))
				Expect(creator.IdBytes).To(Equal(testIdentity.Certificate().Bytes()))
				seekInfo := getSeekInfo(envelope)
				Expect(seekInfo.Start.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Newest{}))
				Expect(seekInfo.Stop.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Newest{}))
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_BLOCK_UNTIL_READY))
				envelope, _ = fakeDeliverer.DeliverArgsForCall(1)
				channelHeader = getChannelHeader(envelope)
				Expect(channelHeader.Type).To(BeEquivalentTo(common.HeaderType_DELIVER_SEEK_INFO))
				Expect(channelHeader.ChannelId).To(Equal("mychannel"))
				creator = getCreator(envelope)
				Expect(creator.Mspid).To(Equal("Org1MSP"))
				Expect(creator.IdBytes).To(Equal(testIdentity.Certificate().Bytes()))
				seekInfo = getSeekInfo(envelope)
				Expect(seekInfo.Start.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				Expect(seekInfo.Stop.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				start := seekInfo.Start.Type.(*orderer.SeekPosition_Specified)
				Expect(start.Specified.Number).To(BeEquivalentTo(1337))
				stop := seekInfo.Stop.Type.(*orderer.SeekPosition_Specified)
				Expect(stop.Specified.Number).To(BeEquivalentTo(1337))
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_BLOCK_UNTIL_READY))
			})
		})

		When("the deliver request fails", func() {
			It("returns the error", func() {
				fakeDeliverer.DeliverReturns(errors.New("fake error"))
				block, err := blocks.GetConfigBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("fake error"))
				Expect(block).To(BeNil())
			})
		})

		When("no blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					return nil
				})
				block, err := blocks.GetConfigBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("No blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

		When("multiple blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					if err := callback(fakeBlock); err != nil {
						return err
					}
					if err := callback(fakeBlock); err != nil {
						return err
					}
					return nil
				})
				block, err := blocks.GetConfigBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("Multiple blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

	})

	Context("blocks.GetNewestBlock()", func() {

		When("called for a channel", func() {
			It("returns the newest block", func() {
				block, err := blocks.GetNewestBlock(fakeDeliverer, "mychannel")
				Expect(err).NotTo(HaveOccurred())
				Expect(block).To(Equal(fakeBlock))
				Expect(fakeDeliverer.DeliverCallCount()).To(Equal(1))
				envelope, _ := fakeDeliverer.DeliverArgsForCall(0)
				channelHeader := getChannelHeader(envelope)
				Expect(channelHeader.Type).To(BeEquivalentTo(common.HeaderType_DELIVER_SEEK_INFO))
				Expect(channelHeader.ChannelId).To(Equal("mychannel"))
				creator := getCreator(envelope)
				Expect(creator.Mspid).To(Equal("Org1MSP"))
				Expect(creator.IdBytes).To(Equal(testIdentity.Certificate().Bytes()))
				seekInfo := getSeekInfo(envelope)
				Expect(seekInfo.Start.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Newest{}))
				Expect(seekInfo.Stop.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Newest{}))
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_BLOCK_UNTIL_READY))
			})
		})

		When("the deliver request fails", func() {
			It("returns the error", func() {
				fakeDeliverer.DeliverReturns(errors.New("fake error"))
				block, err := blocks.GetNewestBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("fake error"))
				Expect(block).To(BeNil())
			})
		})

		When("no blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					return nil
				})
				block, err := blocks.GetNewestBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("No blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

		When("multiple blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					if err := callback(fakeBlock); err != nil {
						return err
					}
					if err := callback(fakeBlock); err != nil {
						return err
					}
					return nil
				})
				block, err := blocks.GetNewestBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("Multiple blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

	})

	Context("blocks.GetSpecificBlock()", func() {

		When("called for a channel", func() {
			It("returns the specified block", func() {
				block, err := blocks.GetSpecificBlock(fakeDeliverer, "mychannel", 1337)
				Expect(err).NotTo(HaveOccurred())
				Expect(block).To(Equal(fakeBlock))
				Expect(fakeDeliverer.DeliverCallCount()).To(Equal(1))
				envelope, _ := fakeDeliverer.DeliverArgsForCall(0)
				channelHeader := getChannelHeader(envelope)
				Expect(channelHeader.Type).To(BeEquivalentTo(common.HeaderType_DELIVER_SEEK_INFO))
				Expect(channelHeader.ChannelId).To(Equal("mychannel"))
				creator := getCreator(envelope)
				Expect(creator.Mspid).To(Equal("Org1MSP"))
				Expect(creator.IdBytes).To(Equal(testIdentity.Certificate().Bytes()))
				seekInfo := getSeekInfo(envelope)
				Expect(seekInfo.Start.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				Expect(seekInfo.Stop.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				start := seekInfo.Start.Type.(*orderer.SeekPosition_Specified)
				Expect(start.Specified.Number).To(BeEquivalentTo(1337))
				stop := seekInfo.Stop.Type.(*orderer.SeekPosition_Specified)
				Expect(stop.Specified.Number).To(BeEquivalentTo(1337))
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_BLOCK_UNTIL_READY))
			})
		})

		When("the deliver request fails", func() {
			It("returns the error", func() {
				fakeDeliverer.DeliverReturns(errors.New("fake error"))
				block, err := blocks.GetSpecificBlock(fakeDeliverer, "mychannel", 1337)
				Expect(err).To(MatchError("fake error"))
				Expect(block).To(BeNil())
			})
		})

		When("no blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					return nil
				})
				block, err := blocks.GetSpecificBlock(fakeDeliverer, "mychannel", 1337)
				Expect(err).To(MatchError("No blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

		When("multiple blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					if err := callback(fakeBlock); err != nil {
						return err
					}
					if err := callback(fakeBlock); err != nil {
						return err
					}
					return nil
				})
				block, err := blocks.GetSpecificBlock(fakeDeliverer, "mychannel", 1337)
				Expect(err).To(MatchError("Multiple blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

	})

	Context("blocks.GetGenesisBlock()", func() {

		When("called for a channel", func() {
			It("returns the genesis block", func() {
				block, err := blocks.GetGenesisBlock(fakeDeliverer, "mychannel")
				Expect(err).NotTo(HaveOccurred())
				Expect(block).To(Equal(fakeBlock))
				Expect(fakeDeliverer.DeliverCallCount()).To(Equal(1))
				envelope, _ := fakeDeliverer.DeliverArgsForCall(0)
				channelHeader := getChannelHeader(envelope)
				Expect(channelHeader.Type).To(BeEquivalentTo(common.HeaderType_DELIVER_SEEK_INFO))
				Expect(channelHeader.ChannelId).To(Equal("mychannel"))
				creator := getCreator(envelope)
				Expect(creator.Mspid).To(Equal("Org1MSP"))
				Expect(creator.IdBytes).To(Equal(testIdentity.Certificate().Bytes()))
				seekInfo := getSeekInfo(envelope)
				Expect(seekInfo.Start.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				Expect(seekInfo.Stop.Type).To(BeAssignableToTypeOf(&orderer.SeekPosition_Specified{}))
				start := seekInfo.Start.Type.(*orderer.SeekPosition_Specified)
				Expect(start.Specified.Number).To(BeEquivalentTo(0))
				stop := seekInfo.Stop.Type.(*orderer.SeekPosition_Specified)
				Expect(stop.Specified.Number).To(BeEquivalentTo(0))
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_BLOCK_UNTIL_READY))
			})
		})

		When("the deliver request fails", func() {
			It("returns the error", func() {
				fakeDeliverer.DeliverReturns(errors.New("fake error"))
				block, err := blocks.GetGenesisBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("fake error"))
				Expect(block).To(BeNil())
			})
		})

		When("no blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					return nil
				})
				block, err := blocks.GetGenesisBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("No blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

		When("multiple blocks are returned", func() {
			It("returns an error", func() {
				fakeDeliverer.DeliverCalls(func(_ *common.Envelope, callback blocks.DeliverCallback) error {
					if err := callback(fakeBlock); err != nil {
						return err
					}
					if err := callback(fakeBlock); err != nil {
						return err
					}
					return nil
				})
				block, err := blocks.GetGenesisBlock(fakeDeliverer, "mychannel")
				Expect(err).To(MatchError("Multiple blocks returned by seek info request"))
				Expect(block).To(BeNil())
			})
		})

	})

})
