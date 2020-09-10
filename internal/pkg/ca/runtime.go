/*
 * SPDX-License-Identifier: Apache-2.0
 */

package ca

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Start starts the peer.
func (c *CA) Start() error {
	logsDirectory := filepath.Join(c.directory, "logs")
	if err := os.MkdirAll(logsDirectory, 0755); err != nil {
		return err
	}
	certfile := filepath.Join(c.directory, "ca-cert.pem")
	keyfile := filepath.Join(c.directory, "ca-key.pem")
	if err := ioutil.WriteFile(certfile, c.identity.Certificate().Bytes(), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyfile, c.identity.PrivateKey().Bytes(), 0644); err != nil {
		return err
	}
	cmd := exec.Command(
		"fabric-ca-server",
		"start",
		"--boot",
		"admin:adminpw",
		"--ca.certfile",
		certfile,
		"--ca.keyfile",
		keyfile,
		"--port",
		strconv.Itoa(int(c.apiPort)),
		"--operations.listenaddress",
		fmt.Sprintf("0.0.0.0:%d", c.operationsPort),
	)
	cmd.Dir = c.directory
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	logFile, err := os.OpenFile(path.Join(logsDirectory, "ca.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
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
		id := strings.ToLower(c.identity.Name())
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
	c.command = cmd
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
			c.Stop()
			return errors.New("timeout whilst waiting for CA to start")
		case err := <-errchan:
			c.Stop()
			return errors.WithMessage(err, "failed to start CA")
		case <-tick:
			if c.hasStarted() {
				return nil
			}
		}
	}
}

// Stop stops the peer.
func (c *CA) Stop() error {
	if c.command != nil {
		err := c.command.Process.Kill()
		if err != nil {
			return errors.WithMessage(err, "failed to stop peer")
		}
		c.command = nil
	}
	return nil
}

func (c *CA) hasStarted() bool {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", c.operationsPort))
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
