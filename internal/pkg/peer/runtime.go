/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/IBM-Blockchain/microfab/internal/pkg/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var logger = log.New(os.Stdout, fmt.Sprintf("[%16s] ", "console"), log.LstdFlags)

// Start starts the peer.
func (p *Peer) Start(timeout time.Duration) error {
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
	go func() {
		reader := bufio.NewReader(pipe)
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanLines)
		id := strings.ToLower(p.identity.Name())
		id = strings.ReplaceAll(id, " ", "")
		logger := log.New(os.Stdout, fmt.Sprintf("[%16s] ", id), 0)
		for scanner.Scan() {
			logger.Println(scanner.Text())
			logFile.WriteString(scanner.Text())
		}
		pipe.Close()
		logFile.Close()
	}()
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
	timeoutCh := time.After(timeout)
	tick := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-timeoutCh:
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

func (p *Peer) createDirectories() error {
	directories := []string{
		p.directory,
		path.Join(p.directory, "config"),
		path.Join(p.directory, "data"),
		path.Join(p.directory, "logs"),
		path.Join(p.directory, "msp"),
		path.Join(p.directory, "tls"),
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

	logger.Printf("Creating peer with gossip URL %s", p.gossipURL)
	gossip["bootstrap"] = p.gossipURL.Host //p.p.apiURL.Host
	gossip["useLeaderElection"] = false
	gossip["orgLeader"] = true
	gossip["endpoint"] = p.gossipURL.Host
	gossip["externalEndpoint"] = p.gossipURL.Host

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
			"path": path.Join(homeDirectory, "builders", "ccaas"),
			"name": "chaincode-as-a-service-builder",
			"propagateEnvironment": []string{
				"CHAINCODE_AS_A_SERVICE_BUILDER_CONFIG",
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
	ledger, ok := config["ledger"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing ledger section")
	}
	state, ok := ledger["state"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing ledger.state section")
	}

	snapshots, ok := ledger["snapshots"].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("core.yaml missing ledger.snapshots section")
	}
	snapshots["rootDir"] = path.Join(dataDirectory, "snapshots")

	if p.couchDB {
		state["stateDatabase"] = "CouchDB"
		couchDBConfig, ok := state["couchDBConfig"].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("core.yaml missing ledger.state.couchDBConfig section")
		}
		couchDBConfig["couchDBAddress"] = fmt.Sprintf("localhost:%d", p.couchDBPort)
		couchDBConfig["username"] = "admin"
		couchDBConfig["password"] = "adminpw"

	}
	if p.tls != nil {
		tlsDirectory := path.Join(p.directory, "tls")
		certFile := path.Join(tlsDirectory, "cert.pem")
		keyFile := path.Join(tlsDirectory, "key.pem")
		caFile := path.Join(tlsDirectory, "ca.pem")
		tls, ok := peer["tls"].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("core.yaml missing peer.tls section")
		}
		tls["enabled"] = true
		tls["cert"] = map[string]string{"file": certFile}
		tls["key"] = map[string]string{"file": keyFile}

		tls["rootcert"] = map[string]string{"file": caFile}
		tls["clientRootCAs"] = map[string]string{"file": caFile}
		tls, ok = operations["tls"].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("core.yaml missing operations.tls section")
		}
		tls["enabled"] = true
		tls["cert"] = map[string]string{"file": certFile}
		tls["key"] = map[string]string{"file": keyFile}

		if err := ioutil.WriteFile(certFile, p.tls.Certificate().Bytes(), 0644); err != nil {
			return err
		}
		if err := ioutil.WriteFile(keyFile, p.tls.PrivateKey().Bytes(), 0644); err != nil {
			return err
		}
		if err := ioutil.WriteFile(caFile, p.tls.CA().Bytes(), 0644); err != nil {
			return err
		}
	}
	configData, err = yaml.Marshal(config)
	if err != nil {
		return err
	}
	configFile = path.Join(p.directory, "config", "core.yaml")
	return ioutil.WriteFile(configFile, configData, 0644)
}

func (p *Peer) hasStarted() bool {
	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := cli.Get(fmt.Sprintf("%s/healthz", p.OperationsURL(true)))
	if err != nil {
		log.Printf("error waiting for peer: %v\n", err)
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	connection, err := Connect(p, p.mspID, p.organization.Admin())
	if err != nil {
		return false
	}
	defer connection.Close()
	_, err = connection.ListChannels()
	if err != nil {
		return false
	}
	return true
}
