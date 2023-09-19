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
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/satlabrpcserver/fake"
	"infra/cros/satlab/satlabrpcserver/models"
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

	expected := &models.SSHResult{IP: "127.0.0.1", Value: expectedResponse}
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

	expected := []*models.SSHResult{{IP: "127.0.0.1", Value: expectedResponse}}
	if diff := cmp.Diff(expected, res); diff != "" {
		t.Errorf("Got diff response, Expected %v, Got %v", expected, res)
	}
}

func TestPingDUTsShouldSuccess(t *testing.T) {
	// We Set this test run in parallel
	t.Parallel()

	ctx := context.Background()

	// We fake the command executor
	dutServices := DUTServicesImpl{
		commandExecutor: &executor.FakeCommander{
			CmdOutput: "192.168.231.2",
		},
	}

	input := []string{"192.168.231.2", "192.168.231.3"}
	res, err := dutServices.pingDUTs(ctx, input)
	if err != nil {
		t.Errorf("Expected should succes, but got an error: %v\n", err)
	}

	expectedActiveIPs := []string{"192.168.231.2"}

	if diff := cmp.Diff(expectedActiveIPs, res); diff != "" {
		t.Errorf("Expected %v, got %v\n", expectedActiveIPs, res)
	}
}

func TestFetchLeasesShouldWork(t *testing.T) {
	// We Set this test run in parallel
	t.Parallel()

	// We fake the command executor
	dutServices := DUTServicesImpl{
		commandExecutor: &executor.FakeCommander{
			CmdOutput: `
1694651422 00:14:3d:14:c4:02 192.168.231.221 * 01:00:14:3d:14:c4:02
1694634664 e8:9f:80:83:3d:c8 192.168.231.213 * 01:e8:9f:80:83:3d:c8
1694301051 88:54:1f:0f:5f:dd 192.168.231.163 * 01:88:54:1f:0f:5f:dd
1694283411 e8:9f:80:83:74:fe 192.168.231.201 * 01:e8:9f:80:83:74:fe
      `,
		},
	}

	res, err := dutServices.fetchLeasesFile()
	if err != nil {
		t.Errorf("Expected should succes, but got an error: %v\n", err)
	}

	expectedActiveIPs := []string{"192.168.231.221", "192.168.231.213", "192.168.231.163", "192.168.231.201"}

	if diff := cmp.Diff(expectedActiveIPs, res); diff != "" {
		t.Errorf("Expected %v, got %v\n", expectedActiveIPs, res)
	}

}
