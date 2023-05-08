// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dut_services

import (
	"context"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/go-cmp/cmp"

	cssh "golang.org/x/crypto/ssh"
	"infra/cros/satlab/satlabrpcserver/fake"
	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/connector"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

func TestRunCommandOnIpShouldWork(t *testing.T) {
	expectedResponse := "connect success"
	server, err := fake.NewFakeServer(func(session ssh.Session) {
		_, err := io.WriteString(session, expectedResponse)
		if err != nil {
			log.Printf("Can't write the response to ssh client")
			return
		}
	})

	if err != nil {
		t.Errorf("Can't create a fake ssh server")
		return
	}

	go func() {
		err := server.Serve()
		// issue: https://github.com/golang/go/issues/43722
		time.Sleep(time.Second * 5)
		if err != nil {
			// If we start server, and we get an error,
			// we don't need to test other things.
			t.Errorf("Can't listen the addr: %v", err)
			return
		}
	}()

	t.Cleanup(func() {
		if server != nil {
			err := server.Close()
			// We can't do anything here
			// when closing the fake ssh server error.
			// Instead, we can log the error message.
			if err != nil {
				log.Printf("Can't close the fake server")
				return
			}
		}
	})

	// Wait for ssh server bring up
	time.Sleep(time.Second * 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	config := cssh.ClientConfig{
		User: "fake_user",
		Auth: []cssh.AuthMethod{
			cssh.Password(fake.Password),
		},
		HostKeyCallback: cssh.InsecureIgnoreHostKey(),
		Timeout:         constants.SSHConnectionTimeout,
	}

	dutServices := DUTServicesImpl{
		config: config,
		// if server is running, it should listen to some tcp
		// so the pattern should be "xxx.xxx.xxx.xxx:xx".
		// we can split the string into two part
		port:            strings.Split(server.GetAddr(), ":")[1],
		clientConnector: connector.New(0, time.Second),
	}

	res, err := dutServices.RunCommandOnIP(ctx, "127.0.0.1", "echo")
	if err != nil {
		t.Errorf("Run command failed")
	}

	expected := &utils.SSHResult{IP: "127.0.0.1", Value: expectedResponse}
	if diff := cmp.Diff(expected, res); diff != "" {
		t.Errorf("Got diff response, Expected %v, Got %v", expected, res)
	}
}

func TestRunCommandOnIpsShouldWork(t *testing.T) {
	expectedResponse := "connect success"
	server, err := fake.NewFakeServer(func(session ssh.Session) {
		_, err := io.WriteString(session, expectedResponse)
		if err != nil {
			log.Printf("Can't write the response to ssh client")
			return
		}
	})

	if err != nil {
		t.Errorf("Can't create a fake ssh server")
		return
	}

	go func() {
		err := server.Serve()
		// issue: https://github.com/golang/go/issues/43722
		time.Sleep(time.Second * 5)
		if err != nil {
			// If we start server, and we get an error,
			// we don't need to test other things.
			t.Errorf("Can't listen the addr: %v", server.GetAddr())
			return
		}
	}()

	t.Cleanup(func() {
		err := server.Close()
		// We can't do anything here
		// when closing the fake ssh server error.
		// Instead, we can log the error message.
		if err != nil {
			log.Printf("Can't close the fake server")
			return
		}
	})

	// Wait for ssh server bring up
	time.Sleep(time.Second * 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	config := cssh.ClientConfig{
		User: "fake_user",
		Auth: []cssh.AuthMethod{
			cssh.Password(fake.Password),
		},
		HostKeyCallback: cssh.InsecureIgnoreHostKey(),
		Timeout:         constants.SSHConnectionTimeout,
	}

	dutServices := DUTServicesImpl{
		config: config,
		// if server is running, it should listen to some tcp
		// so the pattern should be "xxx.xxx.xxx.xxx:xx".
		// we can split the string into two part
		port:            strings.Split(server.GetAddr(), ":")[1],
		clientConnector: connector.New(0, time.Second),
	}

	res := dutServices.RunCommandOnIPs(ctx, []string{"127.0.0.1"}, "echo")

	expected := []*utils.SSHResult{{IP: "127.0.0.1", Value: expectedResponse}}
	if diff := cmp.Diff(expected, res); diff != "" {
		t.Errorf("Got diff response, Expected %v, Got %v", expected, res)
	}
}
