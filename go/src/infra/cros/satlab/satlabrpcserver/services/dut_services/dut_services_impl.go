// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package dut_services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"

	"infra/cros/satlab/satlabrpcserver/utils"
	"infra/cros/satlab/satlabrpcserver/utils/connector"
	"infra/cros/satlab/satlabrpcserver/utils/constants"
)

type ListFirmwareCommandResponse struct {
	FwId     string                         `json:"fwid"`
	Model    string                         `json:"model"`
	FwUpdate map[string]*ListFirmwareResult `json:"fw_update"`
}

type ListFirmwareResult struct {
	Host        *Host                  `json:"host"`
	Ec          map[string]interface{} `json:"ec"`
	SignatureId string                 `json:"signature_id"`
}

type Host struct {
	Versions *HostVersions `json:"versions"`
}

type HostVersions struct {
	RO string `json:"ro"`
	RW string `json:"rw"`
}

// DUTServicesImpl implement details of IDUTServices
type DUTServicesImpl struct {
	// config store the ssh configuration because we don't need
	// to create the config everytime.
	config ssh.ClientConfig
	// add this for testing
	port string
	// define a interface for how to connect to the host via ssh
	clientConnector connector.ISSHClientConnector
}

func New() (*DUTServicesImpl, error) {
	// TODO we should read from file, but we don't know the path now.
	signer, err := utils.ReadSSHKey(constants.SSHKeyPath)
	if err != nil {
		return nil, err
	}
	config := ssh.ClientConfig{
		User: constants.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         constants.SSHConnectionTimeout,
	}
	sshConnector := connector.New(constants.SSHMaxRetry, constants.SSHConnectionTimeout)

	return &DUTServicesImpl{
		config:          config,
		port:            constants.SSHPort,
		clientConnector: sshConnector,
	}, nil
}

// RunCommandOnIP send the command to the DUT device and then get the result back
//
// ip which device ip want to execute the command.
// cmd which command want to be executed.
// TODO: consider one thing if the command was executed failed should be an error?
func (d *DUTServicesImpl) RunCommandOnIP(ctx context.Context, IP string, cmd string) (*utils.SSHResult, error) {
	client, err := d.clientConnector.Connect(ctx, IP+":"+d.port, &d.config)
	if err != nil {
		log.Printf("Can't create a ssh client %v", err)
		return nil, err
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {
			log.Printf("Can't close a ssh client, %v", err)
		}
	}(client)

	session, err := client.NewSession()
	if err != nil {
		log.Printf("Can't create a ssh session, %v", err)
		return nil, err
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		// BUG: https://github.com/golang/go/issues/38115
		if err != nil && err != io.EOF {
			log.Printf("Can't close a ssh session, %v", err)
		}
	}(session)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		var out bytes.Buffer
		var outErr bytes.Buffer
		session.Stdout = &out
		session.Stderr = &outErr
		result := &utils.SSHResult{IP: IP}

		err = session.Run(cmd)
		if err != nil {
			result.Error = errors.New(outErr.String())
			return result, nil
		}

		result.Value = out.String()
		return result, nil
	}
}

// RunCommandOnIPs send the command to DUT devices and then get the result back
//
// ips the list of ip which want to execute the command.
// cmd which command want to be executed.
func (d *DUTServicesImpl) RunCommandOnIPs(ctx context.Context, IPs []string, cmd string) ([]*utils.SSHResult, error) {
	ch := make(chan *utils.SSHResult)

	var wg sync.WaitGroup

	for _, IP := range IPs {
		wg.Add(1)
		go func(IP string) {
			defer wg.Done()
			out, err := d.RunCommandOnIP(ctx, IP, cmd)
			// SSH connection error, we can't do anything here.
			// log the error message.
			if err != nil {
				log.Printf("Run command on IP: %s failed because the connection problem: %v", IP, err)
				return
			}
			ch <- out
		}(IP)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var res []*utils.SSHResult
	for data := range ch {
		res = append(res, data)
	}

	return res, nil
}
