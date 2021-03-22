// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Command fleet-tlw implements the TLS wiring API for Chrome OS fleet labs.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"

	"infra/cros/cmd/fleet-tlw/internal/cache"
)

var (
	port        = flag.Int("port", 0, "Port to listen to")
	proxySSHKey = flag.String("proxy-ssh-key", "", "Path to SSH key for SSH proxy servers (no auth for ExposePortToDut Proxy Mode if unset)")
)

func main() {
	if err := innerMain(); err != nil {
		log.Fatalf("fleet-tlw: %s", err)
	}
}

func innerMain() error {
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("fleet-tlw: %s", err)
	}
	s := grpc.NewServer()

	// TODO (guocb) Fetch caching backends data from UFS after migration to
	// caching cluster.
	ce, err := cache.NewDevserverEnv(cache.AutotestConfig)
	if err != nil {
		return err
	}

	proxySSHConfig, err := getSSHClientConfigForProxy(*proxySSHKey)
	if err != nil {
		return err
	}

	tlw := newTLWServer(ce, proxySSHConfig)
	tlw.registerWith(s)
	defer tlw.Close()

	// TODO(sanikak): Every time a new parameter is added to the tlw server,
	// it needs to be added to the session server. This is not ideal. A better
	// way to accomplish the same objective should be developed.
	ss := newSessionServer(ce, proxySSHConfig)
	ss.registerWith(s)
	defer ss.Close()

	c := setupSignalHandler()
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := <-c
		log.Printf("Captured %v, stopping fleet-tlw service and cleaning up...", sig)
		s.GracefulStop()
	}()
	return s.Serve(l)
}

func getSSHClientConfigForProxy(sshKeyFile string) (*ssh.ClientConfig, error) {
	sshConfig := &ssh.ClientConfig{
		User:            "chromeos-test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	if sshKeyFile != "" {
		m, err := authMethodFromKey(sshKeyFile)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{m}
	}
	return sshConfig, nil
}

func authMethodFromKey(keyfile string) (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}
