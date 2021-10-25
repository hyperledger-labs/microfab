/*
 * SPDX-License-Identifier: Apache-2.0
 */

package ca

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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Start starts the peer.
func (c *CA) Start(timeout time.Duration) error {
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
	args := []string{
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
		"--ca.name",
		fmt.Sprintf("%sca", strings.ToLower(c.organization.Name())),
	}
	if c.tls != nil {
		tlsDirectory := path.Join(c.directory, "tls")
		if err := os.MkdirAll(tlsDirectory, 0755); err != nil {
			return err
		}
		certFile := path.Join(tlsDirectory, "cert.pem")
		keyFile := path.Join(tlsDirectory, "key.pem")
		args = append(args,
			"--tls.enabled",
			"--tls.certfile",
			certFile,
			"--tls.keyfile",
			keyFile,
			"--operations.tls.enabled",
			"--operations.tls.certfile",
			certFile,
			"--operations.tls.keyfile",
			keyFile,
		)
		if err := ioutil.WriteFile(certFile, c.tls.Certificate().Bytes(), 0644); err != nil {
			return err
		}
		if err := ioutil.WriteFile(keyFile, c.tls.PrivateKey().Bytes(), 0644); err != nil {
			return err
		}
	}
	fmt.Print(args)
	cmd := exec.Command(
		"fabric-ca-server",
		args...,
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
	timeoutCh := time.After(timeout)
	tick := time.Tick(250 * time.Millisecond)
	for {
		select {
		case <-timeoutCh:
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
	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := cli.Get(fmt.Sprintf("%s/healthz", c.OperationsURL(true)))
	if err != nil {
		log.Printf("error waiting for CA: %v\n", err)
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}
