/*
 * SPDX-License-Identifier: Apache-2.0
 */

package orderer

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/protoutil"
	"github.com/IBM-Blockchain/microfab/internal/pkg/txid"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/pkg/errors"
)

// Orderer represents a loaded orderer definition.
type Orderer struct {
	organization   *organization.Organization
	mspID          string
	identity       *identity.Identity
	directory      string
	apiPort        int
	apiURL         *url.URL
	operationsPort int
	operationsURL  *url.URL
	command        *exec.Cmd
}

// New creates a new orderer.
func New(organization *organization.Organization, directory string, apiPort int, apiURL string, operationsPort int, operationsURL string) (*Orderer, error) {
	identityName := fmt.Sprintf("%s Orderer", organization.Name())
	identity, err := identity.New(identityName, identity.WithOrganizationalUnit("orderer"), identity.UsingSigner(organization.CA()))
	if err != nil {
		return nil, err
	}
	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	parsedOperationsURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, err
	}
	return &Orderer{organization, organization.MSPID(), identity, directory, apiPort, parsedAPIURL, operationsPort, parsedOperationsURL, nil}, nil
}

// Organization returns the organization of the orderer.
func (o *Orderer) Organization() *organization.Organization {
	return o.organization
}

// MSPID returns the MSP ID of the orderer.
func (o *Orderer) MSPID() string {
	return o.mspID
}

// APIPort returns the API port of the orderer.
func (o *Orderer) APIPort() int {
	return o.apiPort
}

// APIURL returns the API URL of the orderer.
func (o *Orderer) APIURL() *url.URL {
	return o.apiURL
}

// OperationsPort returns the operations port of the orderer.
func (o *Orderer) OperationsPort() int {
	return o.operationsPort
}

// OperationsURL returns the operations URL of the orderer.
func (o *Orderer) OperationsURL() *url.URL {
	return o.operationsURL
}

// Host returns the host (hostname:port) of the orderer.
func (o *Orderer) Host() string {
	return o.apiURL.Host
}

// Hostname returns the hostname of the orderer.
func (o *Orderer) Hostname() string {
	return o.apiURL.Hostname()
}

// Port returns the port of the orderer.
func (o *Orderer) Port() int32 {
	port, _ := strconv.Atoi(o.apiURL.Port())
	return int32(port)
}

func (o *Orderer) createDirectories() error {
	directories := []string{
		o.directory,
		path.Join(o.directory, "config"),
		path.Join(o.directory, "data"),
		path.Join(o.directory, "logs"),
		path.Join(o.directory, "msp"),
	}
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Orderer) hasStarted() bool {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", o.operationsPort))
	if err != nil {
		return false
	}
	return resp.StatusCode == 200
}

func (o *Orderer) createGenesisBlock(consortium []*organization.Organization) error {
	txID := txid.New(o.mspID, o.identity)
	header := protoutil.BuildHeader(common.HeaderType_CONFIG, "testchainid", txID)
	config := &common.Config{
		ChannelGroup: &common.ConfigGroup{
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
				},
				"Consortiums": {
					Groups: map[string]*common.ConfigGroup{
						"SampleConsortium": {
							Groups:    map[string]*common.ConfigGroup{},
							ModPolicy: "/Channel/Orderer/Admins",
							Values: map[string]*common.ConfigValue{
								"ChannelCreationPolicy": {
									ModPolicy: "/Channel/Orderer/Admins",
									Value:     util.MarshalOrPanic(protoutil.BuildImplicitMetaPolicy(common.ImplicitMetaPolicy_ANY, "Admins")),
								},
							},
						},
					},
					ModPolicy: "/Channel/Orderer/Admins",
				},
				"Orderer": {
					Groups:    map[string]*common.ConfigGroup{},
					ModPolicy: "Admins",
					Policies: map[string]*common.ConfigPolicy{
						"Admins":          protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Admins"),
						"BlockValidation": protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Writers"),
						"Readers":         protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Readers"),
						"Writers":         protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Writers"),
					},
					Values: map[string]*common.ConfigValue{
						"BatchSize": {
							ModPolicy: "Admins",
							Value: util.MarshalOrPanic(&orderer.BatchSize{
								AbsoluteMaxBytes:  103809024,
								MaxMessageCount:   10,
								PreferredMaxBytes: 524288,
							}),
						},
						"BatchTimeout": {
							ModPolicy: "Admins",
							Value: util.MarshalOrPanic(&orderer.BatchTimeout{
								Timeout: "100ms",
							}),
						},
						"Capabilities": {
							ModPolicy: "Admins",
							Value: util.MarshalOrPanic(&common.Capabilities{
								Capabilities: map[string]*common.Capability{
									"V2_0": {},
								},
							}),
						},
						"ChannelRestrictions": {
							ModPolicy: "Admins",
							Value:     nil,
						},
						"ConsensusType": {
							ModPolicy: "Admins",
							Value: util.MarshalOrPanic(&orderer.ConsensusType{
								Metadata: nil,
								State:    orderer.ConsensusType_STATE_NORMAL,
								Type:     "solo",
							}),
						},
					},
				},
			},
			ModPolicy: "Admins",
			Policies: map[string]*common.ConfigPolicy{
				"Admins":  protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Admins"),
				"Readers": protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Readers"),
				"Writers": protoutil.BuildImplicitMetaConfigPolicy(common.ImplicitMetaPolicy_ANY, "Writers"),
			},
			Values: map[string]*common.ConfigValue{
				"BlockDataHashingStructure": {
					ModPolicy: "Admins",
					Value: util.MarshalOrPanic(&common.BlockDataHashingStructure{
						Width: 4294967295,
					}),
				},
				"Capabilities": {
					ModPolicy: "Admins",
					Value: util.MarshalOrPanic(&common.Capabilities{
						Capabilities: map[string]*common.Capability{
							"V2_0": {},
						},
					}),
				},
				"HashingAlgorithm": {
					ModPolicy: "Admins",
					Value: util.MarshalOrPanic(&common.HashingAlgorithm{
						Name: "SHA256",
					}),
				},
				"OrdererAddresses": {
					ModPolicy: "/Channel/Orderer/Admins",
					Value: util.MarshalOrPanic(&common.OrdererAddresses{
						Addresses: []string{
							o.apiURL.Host,
						},
					}),
				},
			},
		},
		Sequence: 0,
	}
	config.ChannelGroup.Groups["Orderer"].Groups[o.organization.MSPID()] = protoutil.BuildConfigGroupFromOrganization(o.organization)
	for _, organization := range consortium {
		config.ChannelGroup.Groups["Consortiums"].Groups["SampleConsortium"].Groups[organization.MSPID()] = protoutil.BuildConfigGroupFromOrganization(organization)
	}
	configEnvelope := &common.ConfigEnvelope{
		Config:     config,
		LastUpdate: nil,
	}
	payload := protoutil.BuildPayload(header, configEnvelope)
	envelope := protoutil.BuildEnvelope(payload, txID)
	genesisBlock := protoutil.BuildGenesisBlock(envelope)
	data := util.MarshalOrPanic(genesisBlock)
	configDirectory := path.Join(o.directory, "config")
	return ioutil.WriteFile(path.Join(configDirectory, "genesisblock"), data, 0644)
}

// Start starts the orderer.
func (o *Orderer) Start(consortium []*organization.Organization) error {
	err := o.createDirectories()
	if err != nil {
		return err
	}
	configDirectory := path.Join(o.directory, "config")
	dataDirectory := path.Join(o.directory, "data")
	logsDirectory := path.Join(o.directory, "logs")
	mspDirectory := path.Join(o.directory, "msp")
	err = util.CreateMSPDirectory(mspDirectory, o.identity)
	if err != nil {
		return err
	}
	err = o.createGenesisBlock(consortium)
	if err != nil {
		return err
	}
	cmd := exec.Command("orderer", "start")
	cmd.Env = os.Environ()
	extraEnvs := []string{
		fmt.Sprintf("ORDERER_GENERAL_LOCALMSPDIR=%s", mspDirectory),
		fmt.Sprintf("ORDERER_GENERAL_LOCALMSPID=%s", o.mspID),
		"ORDERER_GENERAL_BOOTSTRAPMETHOD=file",
		fmt.Sprintf("ORDERER_GENERAL_BOOTSTRAPFILE=%s", path.Join(configDirectory, "genesisblock")),
		fmt.Sprintf("ORDERER_FILELEDGER_LOCATION=%s", dataDirectory),
		fmt.Sprintf("ORDERER_CONSENSUS_WALDIR=%s", path.Join(dataDirectory, "etcdraft", "wal")),
		fmt.Sprintf("ORDERER_CONSENSUS_SNAPDIR=%s", path.Join(dataDirectory, "etcdraft", "snapshot")),
		"ORDERER_METRICS_PROVIDER=prometheus",
		"ORDERER_GENERAL_LISTENADDRESS=0.0.0.0",
		fmt.Sprintf("ORDERER_GENERAL_LISTENPORT=%d", o.apiPort),
		fmt.Sprintf("ORDERER_OPERATIONS_LISTENADDRESS=0.0.0.0:%d", o.operationsPort),
	}
	cmd.Env = append(cmd.Env, extraEnvs...)
	cmd.Stdin = nil
	logFile, err := os.OpenFile(path.Join(logsDirectory, "orderer.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.WithMessage(err, "failed to open orderer log file")
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return errors.WithMessage(err, "failed to open pipe")
	}
	go io.Copy(logFile, pipe)
	cmd.Stderr = cmd.Stdout
	err = cmd.Start()
	if err != nil {
		return errors.WithMessage(err, "failed to start orderer")
	}
	o.command = cmd
	errchan := make(chan error, 1)
	go func() {
		err = cmd.Wait()
		if err != nil {
			errchan <- err
		}
	}()
	timeout := time.After(10 * time.Second)
	tick := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-timeout:
			o.Stop()
			return errors.New("timeout whilst waiting for orderer to start")
		case err := <-errchan:
			o.Stop()
			return errors.WithMessage(err, "failed to start orderer")
		case <-tick:
			if o.hasStarted() {
				return nil
			}
		}
	}
}

// Stop stops the orderer.
func (o *Orderer) Stop() error {
	if o.command != nil {
		err := o.command.Process.Kill()
		if err != nil {
			return errors.WithMessage(err, "failed to stop orderer")
		}
		o.command = nil
	}
	return nil
}
