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

const testIdentityText = `{
    "name": "Org1 Admin",
    "cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNQakNDQWVXZ0F3SUJBZ0lVWmJrVndnamV3TC9sTnhOYU1qQzFsb1VkMFNvd0NnWUlLb1pJemowRUF3SXcKV2pFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVFzd0NRWURWUVFERXdKallUQWVGdzB5Ck1EQTRNRE14TlRFMk1EQmFGdzB5TVRBNE1ETXhOVEl4TURCYU1DUXhEakFNQmdOVkJBc1RCV0ZrYldsdU1SSXcKRUFZRFZRUURFd2x2Y21jeFlXUnRhVzR3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVJTYUxaRApjTmMxb3NybU4wdkRzTzloa0dYMVBhUjV0NzJYNitCV3hDaURNWWZyTGNYczlkOW9LNU9oQkJWWkJ0dEkwWVR6Cm41aTN6cnZxSjhXOXFsOURvNEcrTUlHN01BNEdBMVVkRHdFQi93UUVBd0lIZ0RBTUJnTlZIUk1CQWY4RUFqQUEKTUIwR0ExVWREZ1FXQkJSYVAzVjNhbUplZmdaV1hWUUhyOUR5UFJ1RDZqQWZCZ05WSFNNRUdEQVdnQlR4TEwxbApxalExU1ovcG5wb1JuekFqUjBHbHF6QmJCZ2dxQXdRRkJnY0lBUVJQZXlKaGRIUnljeUk2ZXlKb1ppNUJabVpwCmJHbGhkR2x2YmlJNklpSXNJbWhtTGtWdWNtOXNiRzFsYm5SSlJDSTZJbTl5WnpGaFpHMXBiaUlzSW1obUxsUjUKY0dVaU9pSmhaRzFwYmlKOWZUQUtCZ2dxaGtqT1BRUURBZ05IQURCRUFpQXdzZ1pkaGlnVlB2ZTF2d3VJQTEydgpRRTdJaDFYeE1KWkJoZjAzL3pZTnVBSWdZRER5ckVCREp1N2JhS0hsbGNHWkRXK0pST29tS2dDejVuMFRKZStYCmYrTT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=",
    "ca": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNDVENDQWErZ0F3SUJBZ0lVWjdDSzJGYkthODR2Tmt1T241bzkwdWlyMHR3d0NnWUlLb1pJemowRUF3SXcKV2pFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVFzd0NRWURWUVFERXdKallUQWVGdzB5Ck1EQTRNRE14TlRFME1EQmFGdzB6TlRBM016RXhOVEUwTURCYU1Gb3hDekFKQmdOVkJBWVRBbFZUTVJjd0ZRWUQKVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFVU1CSUdBMVVFQ2hNTFNIbHdaWEpzWldSblpYSXhEekFOQmdOVgpCQXNUQmtaaFluSnBZekVMTUFrR0ExVUVBeE1DWTJFd1dUQVRCZ2NxaGtqT1BRSUJCZ2dxaGtqT1BRTUJCd05DCkFBVFoxd25tNkxzN2c0c2VYbkZUYm1uWlFZNFpJNWo1Z2dBMFh4L0g1UFhxYVFVN09LaTZaUTBoZk13NVNGT3QKdzIyRldDTjJ2WFBzWE1jYjF0WUFKYWt3bzFNd1VUQU9CZ05WSFE4QkFmOEVCQU1DQVFZd0R3WURWUjBUQVFILwpCQVV3QXdFQi96QWRCZ05WSFE0RUZnUVU4U3k5WmFvME5VbWY2WjZhRVo4d0kwZEJwYXN3RHdZRFZSMFJCQWd3CkJvY0Vmd0FBQVRBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlFQXNEOUFrcENEQnhKN2ZrSkhUK0kxWHZPSUVDSXAKYnNQenFOMTd5a1JNOGNFQ0lEL0RhYk5CeWRaUTEvRGJVdFZTUjVQSk1uS202b0dkSjMrWkQ0Vm43QWJqCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K",
    "hsm": false,
    "private_key": "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZ2RUdDdlUGtNcG0rODV1QkIKMFZqQzZ4RUlUekxJSlZDWmhNT3NMQ0ZhT3I2aFJBTkNBQVJTYUxaRGNOYzFvc3JtTjB2RHNPOWhrR1gxUGFSNQp0NzJYNitCV3hDaURNWWZyTGNYczlkOW9LNU9oQkJWWkJ0dEkwWVR6bjVpM3pydnFKOFc5cWw5RAotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg=="
}`

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

	Context("blocks.GetConfigBlock", func() {

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
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_FAIL_IF_NOT_READY))
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
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_FAIL_IF_NOT_READY))
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

	Context("blocks.GetNewestBlock", func() {

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
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_FAIL_IF_NOT_READY))
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

	Context("blocks.GetSpecificBlock", func() {

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
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_FAIL_IF_NOT_READY))
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

	Context("blocks.GetGenesisBlock", func() {

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
				Expect(seekInfo.Behavior).To(Equal(orderer.SeekInfo_FAIL_IF_NOT_READY))
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
