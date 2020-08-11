/*
 * SPDX-License-Identifier: Apache-2.0
 */

package peer

import (
	"fmt"
	"io"
	"io/ioutil"
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
		{
			"path": path.Join(homeDirectory, "builders", "wasm"),
			"name": "wasm",
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
