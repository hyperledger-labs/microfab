/*
 * SPDX-License-Identifier: Apache-2.0
 */

package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/IBM-Blockchain/microfab/internal/app/microfabd"
	"github.com/IBM-Blockchain/microfab/internal/pkg/channel"
	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/orderer"
	"github.com/IBM-Blockchain/microfab/internal/pkg/peer"
	"github.com/IBM-Blockchain/microfab/pkg/client"
)

var _ = Describe("Integration", func() {

	var testDirectory string
	var testMicrofab *microfabd.Microfab
	var testClient *client.Client
	var adminIdentity *identity.Identity
	var peerConnections []*peer.Connection
	var ordererConnection *orderer.Connection

	BeforeSuite(func() {
		var err error
		testDirectory, err = ioutil.TempDir("", "microfab-it")
		Expect(err).NotTo(HaveOccurred())
		testConfig := map[string]interface{}{
			"directory": testDirectory,
			"couchdb":   false,
		}
		serializedConfig, err := json.Marshal(testConfig)
		Expect(err).NotTo(HaveOccurred())
		wd, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("MICROFAB_HOME", filepath.Join(wd, ".."))
		os.Setenv("MICROFAB_CONFIG", string(serializedConfig))
		Expect(err).NotTo(HaveOccurred())
		testMicrofab, err = microfabd.New()
		Expect(err).NotTo(HaveOccurred())
		err = testMicrofab.Start()
		Expect(err).NotTo(HaveOccurred())
		testURL, err := url.Parse("http://console.127-0-0-1.nip.io:8080")
		Expect(err).NotTo(HaveOccurred())
		testClient, err = client.New(testURL)
		Expect(err).NotTo(HaveOccurred())
		err = testClient.Ping()
		Expect(err).NotTo(HaveOccurred())
		tempIdentity, err := testClient.GetIdentity("Org1")
		Expect(err).NotTo(HaveOccurred())
		adminIdentity, err = identity.FromClient(tempIdentity)
		Expect(err).NotTo(HaveOccurred())
		tempPeer, err := testClient.GetPeer("Org1")
		Expect(err).NotTo(HaveOccurred())
		peerConnection, err := peer.ConnectClient(tempPeer, tempPeer.MSPID, adminIdentity)
		Expect(err).NotTo(HaveOccurred())
		peerConnections = []*peer.Connection{peerConnection}
		tempOrderingService, err := testClient.GetOrderingService()
		Expect(err).NotTo(HaveOccurred())
		ordererConnection, err = orderer.ConnectClient(tempOrderingService, tempPeer.MSPID, adminIdentity)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		for _, peerConnection := range peerConnections {
			peerConnection.Close()
		}
		testMicrofab.Stop()
		err := os.RemoveAll(testDirectory)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Go chaincode", func() {

		When("a Go chaincode is deployed", func() {
			It("should be available for transactions", func() {
				pkg, err := ioutil.ReadFile("data/asset-transfer-basic-go.tgz")
				Expect(err).NotTo(HaveOccurred())
				for _, peerConnection := range peerConnections {
					packageID, err := peerConnection.InstallChaincode(pkg)
					Expect(err).NotTo(HaveOccurred())
					err = channel.ApproveChaincodeDefinition([]*peer.Connection{peerConnection}, ordererConnection, "channel1", 1, "atb-go", "1.0.0", packageID)
					Expect(err).NotTo(HaveOccurred())
				}
				err = channel.CommitChaincodeDefinition(peerConnections, ordererConnection, "channel1", 1, "atb-go", "1.0.0")
				Expect(err).NotTo(HaveOccurred())
				_, err = channel.SubmitTransaction(peerConnections, ordererConnection, "channel1", "atb-go", "InitLedger")
				Expect(err).NotTo(HaveOccurred())
				data, err := channel.EvaluateTransaction(peerConnections, ordererConnection, "channel1", "atb-go", "GetAllAssets")
				Expect(err).NotTo(HaveOccurred())
				assets := []map[string]interface{}{}
				err = json.Unmarshal(data, &assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(assets).To(HaveLen(6))
			})
		})

	})

	Context("Java chaincode", func() {

		When("a Java chaincode is deployed", func() {
			It("should be available for transactions", func() {
				pkg, err := ioutil.ReadFile("data/asset-transfer-basic-java.tgz")
				Expect(err).NotTo(HaveOccurred())
				for _, peerConnection := range peerConnections {
					packageID, err := peerConnection.InstallChaincode(pkg)
					Expect(err).NotTo(HaveOccurred())
					err = channel.ApproveChaincodeDefinition([]*peer.Connection{peerConnection}, ordererConnection, "channel1", 1, "atb-java", "1.0.0", packageID)
					Expect(err).NotTo(HaveOccurred())
				}
				err = channel.CommitChaincodeDefinition(peerConnections, ordererConnection, "channel1", 1, "atb-java", "1.0.0")
				Expect(err).NotTo(HaveOccurred())
				_, err = channel.SubmitTransaction(peerConnections, ordererConnection, "channel1", "atb-java", "InitLedger")
				Expect(err).NotTo(HaveOccurred())
				data, err := channel.EvaluateTransaction(peerConnections, ordererConnection, "channel1", "atb-java", "GetAllAssets")
				Expect(err).NotTo(HaveOccurred())
				assets := []map[string]interface{}{}
				err = json.Unmarshal(data, &assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(assets).To(HaveLen(6))
			})
		})

	})

	Context("JavaScript chaincode", func() {

		When("a JavaScript chaincode is deployed", func() {
			It("should be available for transactions", func() {
				pkg, err := ioutil.ReadFile("data/asset-transfer-basic-javascript.tgz")
				Expect(err).NotTo(HaveOccurred())
				for _, peerConnection := range peerConnections {
					packageID, err := peerConnection.InstallChaincode(pkg)
					Expect(err).NotTo(HaveOccurred())
					err = channel.ApproveChaincodeDefinition([]*peer.Connection{peerConnection}, ordererConnection, "channel1", 1, "atb-javascript", "1.0.0", packageID)
					Expect(err).NotTo(HaveOccurred())
				}
				err = channel.CommitChaincodeDefinition(peerConnections, ordererConnection, "channel1", 1, "atb-javascript", "1.0.0")
				Expect(err).NotTo(HaveOccurred())
				_, err = channel.SubmitTransaction(peerConnections, ordererConnection, "channel1", "atb-javascript", "InitLedger")
				Expect(err).NotTo(HaveOccurred())
				data, err := channel.EvaluateTransaction(peerConnections, ordererConnection, "channel1", "atb-javascript", "GetAllAssets")
				Expect(err).NotTo(HaveOccurred())
				assets := []map[string]interface{}{}
				err = json.Unmarshal(data, &assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(assets).To(HaveLen(6))
			})
		})

	})

	Context("TypeScript chaincode", func() {

		When("a TypeScript chaincode is deployed", func() {
			It("should be available for transactions", func() {
				pkg, err := ioutil.ReadFile("data/asset-transfer-basic-typescript.tgz")
				Expect(err).NotTo(HaveOccurred())
				for _, peerConnection := range peerConnections {
					packageID, err := peerConnection.InstallChaincode(pkg)
					Expect(err).NotTo(HaveOccurred())
					err = channel.ApproveChaincodeDefinition([]*peer.Connection{peerConnection}, ordererConnection, "channel1", 1, "atb-typescript", "1.0.0", packageID)
					Expect(err).NotTo(HaveOccurred())
				}
				err = channel.CommitChaincodeDefinition(peerConnections, ordererConnection, "channel1", 1, "atb-typescript", "1.0.0")
				Expect(err).NotTo(HaveOccurred())
				_, err = channel.SubmitTransaction(peerConnections, ordererConnection, "channel1", "atb-typescript", "InitLedger")
				Expect(err).NotTo(HaveOccurred())
				data, err := channel.EvaluateTransaction(peerConnections, ordererConnection, "channel1", "atb-typescript", "GetAllAssets")
				Expect(err).NotTo(HaveOccurred())
				assets := []map[string]interface{}{}
				err = json.Unmarshal(data, &assets)
				Expect(err).NotTo(HaveOccurred())
				Expect(assets).To(HaveLen(6))
			})
		})

	})

})
