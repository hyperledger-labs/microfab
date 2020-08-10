/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/identity"
	"github.com/IBM-Blockchain/microfab/internal/pkg/organization"
	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// Peer represents a loaded peer definition.
type Peer struct {
	organization   *organization.Organization
	identity       *identity.Identity
	mspID          string
	directory      string
	apiPort        int32
	apiURL         *url.URL
	chaincodePort  int32
	chaincodeURL   *url.URL
	operationsPort int32
	operationsURL  *url.URL
	clientConn     *grpc.ClientConn
	connIdentity   *identity.Identity
	connMSPID      string
	command        *exec.Cmd
}

type jsonPeer struct {
	APIURL string `json:"api_url"`
	MSPID  string `json:"msp_id"`
	PEM    string `json:"pem"`
}

// New creates a new peer.
func New(organization *organization.Organization, directory string, apiPort int32, apiURL string, chaincodePort int32, chaincodeURL string, operationsPort int32, operationsURL string) (*Peer, error) {
	identityName := fmt.Sprintf("%s Peer", organization.Name())
	identity, err := identity.New(identityName, identity.WithOrganizationalUnit("peer"), identity.UsingSigner(organization.CA()))
	if err != nil {
		return nil, err
	}
	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	parsedChaincodeURL, err := url.Parse(chaincodeURL)
	if err != nil {
		return nil, err
	}
	parsedOperationsURL, err := url.Parse(operationsURL)
	if err != nil {
		return nil, err
	}
	return &Peer{organization, identity, organization.MSP().ID(), directory, apiPort, parsedAPIURL, chaincodePort, parsedChaincodeURL, operationsPort, parsedOperationsURL, nil, nil, "", nil}, nil
}

// FromBytes creates a new connection to a peer from JSON data.
func FromBytes(data []byte) (*Peer, error) {
	// parsedJSON := &jsonPeer{}
	// err := json.Unmarshal(data, &parsedJSON)
	// if err != nil {
	// 	return nil, err
	// }
	// apiURL, err := url.Parse(parsedJSON.APIURL)
	// if err != nil {
	// 	return nil, err
	// }
	// port, err := strconv.Atoi(apiURL.Port())
	// if err != nil {
	// 	return nil, err
	// }
	// secure := apiURL.Scheme == "grpcs"
	// pem, err := base64.StdEncoding.DecodeString(parsedJSON.PEM)
	// if err != nil {
	// 	return nil, err
	// }
	// return &Peer{parsedJSON.APIURL, apiURL.Host, apiURL.Hostname(), int32(port), secure, parsedJSON.MSPID, pem, nil, nil}, nil
	return nil, nil
}

// ToBytes saves the peer to JSON data.
func (p *Peer) ToBytes() ([]byte, error) {
	serializedJSON := jsonPeer{}
	return json.Marshal(serializedJSON)
}

// Organization returns the organization of the peer.
func (p *Peer) Organization() *organization.Organization {
	return p.organization
}

// MSPID returns the MSP ID of the peer.
func (p *Peer) MSPID() string {
	return p.mspID
}

// Connect opens a connection to the peer.
func (p *Peer) Connect(mspID string, identity *identity.Identity) error {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", p.apiPort), grpc.WithInsecure())
	if err != nil {
		return err
	}
	p.clientConn = clientConn
	p.connMSPID = mspID
	p.connIdentity = identity
	return nil
}

// Connected determines if a connection is open to the peer.
func (p *Peer) Connected() bool {
	return p.clientConn != nil
}

// Close closes any open connection to the peer.
func (p *Peer) Close() error {
	p.identity = nil
	if p.clientConn != nil {
		err := p.clientConn.Close()
		p.clientConn = nil
		return err
	}
	return nil
}

// ConnectionMSPID returns the MSP ID used to connect to the peer.
func (p *Peer) ConnectionMSPID() string {
	return p.mspID
}

// ConnectionIdentity returns the identity used to connect to the peer.
func (p *Peer) ConnectionIdentity() *identity.Identity {
	return p.identity
}

// APIPort returns the API port of the peer.
func (p *Peer) APIPort() int32 {
	return p.apiPort
}

// APIURL returns the API URL of the peer.
func (p *Peer) APIURL() *url.URL {
	return p.apiURL
}

// ChaincodePort returns the chaincode port of the peer.
func (p *Peer) ChaincodePort() int32 {
	return p.chaincodePort
}

// ChaincodeURL returns the chaincode URL of the peer.
func (p *Peer) ChaincodeURL() *url.URL {
	return p.chaincodeURL
}

// OperationsPort returns the operations port of the peer.
func (p *Peer) OperationsPort() int32 {
	return p.operationsPort
}

// OperationsURL returns the operations URL of the peer.
func (p *Peer) OperationsURL() *url.URL {
	return p.operationsURL
}

// Host returns the host (hostname:port) of the peer.
func (p *Peer) Host() string {
	return p.apiURL.Host
}

// Hostname returns the hostname of the peer.
func (p *Peer) Hostname() string {
	return p.apiURL.Hostname()
}

// Port returns the port of the peer.
func (p *Peer) Port() int32 {
	port, _ := strconv.Atoi(p.apiURL.Port())
	return int32(port)
}

func (p *Peer) createDirectories() error {
	directories := []string{
		p.directory,
		path.Join(p.directory, "config"),
		path.Join(p.directory, "data"),
		path.Join(p.directory, "logs"),
		path.Join(p.directory, "msp"),
	}
	for _, dir := range directories {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) createConfig(dataDirectory, mspDirectory string) error {
	fabricConfigPath, ok := os.LookupEnv("FABRIC_CFG_PATH")
	if !ok {
		return fmt.Errorf("FABRIC_CFG_PATH not defined")
	}
	configFile := path.Join(fabricConfigPath, "core.yaml")
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	config := map[interface{}]interface{}{}
	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return err
	}
	peer, ok := config["peer"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing peer section")
	}
	peer["id"] = fmt.Sprintf("%speer", strings.ToLower(p.organization.Name()))
	peer["mspConfigPath"] = mspDirectory
	peer["localMspId"] = p.mspID
	peer["fileSystemPath"] = dataDirectory
	peer["address"] = fmt.Sprintf("0.0.0.0:%d", p.apiPort)
	peer["listenAddress"] = fmt.Sprintf("0.0.0.0:%d", p.apiPort)
	peer["chaincodeListenAddress"] = fmt.Sprintf("0.0.0.0:%d", p.chaincodePort)
	gossip, ok := peer["gossip"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing peer.gossip section")
	}
	gossip["bootstrap"] = p.apiURL.Host
	gossip["useLeaderElection"] = false
	gossip["orgLeader"] = true
	gossip["endpoint"] = p.apiURL.Host
	gossip["externalEndpoint"] = p.apiURL.Host
	metrics, ok := config["metrics"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing metrics section")
	}
	metrics["provider"] = "prometheus"
	operations, ok := config["operations"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing operations section")
	}
	operations["listenAddress"] = fmt.Sprintf("0.0.0.0:%d", p.operationsPort)
	vm, ok := config["vm"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing vm section")
	}
	vm["endpoint"] = ""
	chaincode, ok := config["chaincode"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing chaincode section")
	}
	homeDirectory, err := util.GetHomeDirectory()
	if err != nil {
		return err
	}
	chaincode["externalBuilders"] = []map[interface{}]interface{}{
		{
			"path": path.Join(homeDirectory, "builders", "golang"),
			"name": "golang",
			"propagateEnvironment": []string{
				"GOCACHE",
				"GOENV",
				"GOROOT",
				"HOME",
			},
		},
		{
			"path": path.Join(homeDirectory, "builders", "java"),
			"name": "java",
			"propagateEnvironment": []string{
				"HOME",
				"JAVA_HOME",
				"MAVEN_OPTS",
			},
		},
		{
			"path": path.Join(homeDirectory, "builders", "node"),
			"name": "node",
			"propagateEnvironment": []string{
				"HOME",
				"npm_config_cache",
			},
		},
		{
			"path": path.Join(homeDirectory, "builders", "external"),
			"name": "external-service-builder",
			"propagateEnvironment": []string{
				"HOME",
			},
		},
	}
	configData, err = yaml.Marshal(config)
	if err != nil {
		return err
	}
	configFile = path.Join(p.directory, "config", "core.yaml")
	return ioutil.WriteFile(configFile, configData, 0644)
}

func (p *Peer) hasStarted() bool {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", p.operationsPort))
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	err = p.Connect(p.mspID, p.organization.Admin())
	if err != nil {
		return false
	}
	defer p.Close()
	_, err = p.ListChannels()
	if err != nil {
		return false
	}
	return true
}

// Start starts the peer.
func (p *Peer) Start() error {
	err := p.createDirectories()
	if err != nil {
		return err
	}
	configDirectory := path.Join(p.directory, "config")
	dataDirectory := path.Join(p.directory, "data")
	logsDirectory := path.Join(p.directory, "logs")
	mspDirectory := path.Join(p.directory, "msp")
	err = util.CreateMSPDirectory(mspDirectory, p.identity)
	if err != nil {
		return err
	}
	err = p.createConfig(dataDirectory, mspDirectory)
	if err != nil {
		return err
	}
	cmd := exec.Command("peer", "node", "start")
	cmd.Env = os.Environ()
	extraEnvs := []string{
		fmt.Sprintf("FABRIC_CFG_PATH=%s", configDirectory),
	}
	cmd.Env = append(cmd.Env, extraEnvs...)
	cmd.Stdin = nil
	logFile, err := os.OpenFile(path.Join(logsDirectory, "peer.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(logFile, pipe)
	cmd.Stderr = cmd.Stdout
	err = cmd.Start()
	if err != nil {
		return err
	}
	p.command = cmd
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
			p.Stop()
			return errors.New("timeout whilst waiting for peer to start")
		case err := <-errchan:
			p.Stop()
			return errors.WithMessage(err, "failed to start peer")
		case <-tick:
			if p.hasStarted() {
				return nil
			}
		}
	}
}

// Stop stops the peer.
func (p *Peer) Stop() error {
	if p.command != nil {
		err := p.command.Process.Kill()
		if err != nil {
			return errors.WithMessage(err, "failed to stop peer")
		}
		p.command = nil
	}
	return nil
}
