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
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/collection"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/satlabrpcserver/models"
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
	// commandExecutor define a interface for executing a command
	commandExecutor executor.IExecCommander
}

func New() (IDUTServices, error) {
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
	sshConnector := connector.New(constants.SSHMaxRetry, constants.SSHRetryDelay)

	return &DUTServicesImpl{
		config:          config,
		port:            constants.SSHPort,
		clientConnector: sshConnector,
		commandExecutor: &executor.ExecCommander{},
	}, nil
}

// RunCommandOnIP send the command to the DUT device and then get the result back
//
// ip which device ip want to execute the command.
// cmd which command want to be executed.
// TODO: consider one thing if the command was executed failed should be an error?
func (d *DUTServicesImpl) RunCommandOnIP(ctx context.Context, IP string, cmd string) (*models.SSHResult, error) {
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
		result := &models.SSHResult{IP: IP}

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
func (d *DUTServicesImpl) RunCommandOnIPs(ctx context.Context, IPs []string, cmd string) []*models.SSHResult {
	ch := make(chan *models.SSHResult)

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
				ch <- &models.SSHResult{IP: IP, Error: err}
				return
			}
			ch <- out
		}(IP)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var res []*models.SSHResult
	for data := range ch {
		res = append(res, data)
	}

	return res
}

var subnetSearchRe = regexp.MustCompile(`(?P<IP>192\.168\.231\.[0-9][0-9]*[0-9]*).*`)

func (d *DUTServicesImpl) fetchLeasesFile() (map[string]string, error) {
	// List all IPs that we applied.
	out, err := d.commandExecutor.Exec(exec.Command(
		paths.DockerPath,
		"exec",
		"dhcp",
		"/bin/cat",
		paths.LeasesPath,
	))

	if err != nil {
		return nil, err
	}

	rawData := strings.Split(string(out), "\n")
	ipToMAC := map[string]string{}

	dnsmasqIPIndex := 2
	dnsmasqMACAddressIndex := 1
	for _, row := range rawData {
		fields := strings.Fields(row)
		// Handle valid data
		if len(fields) == 5 {
			IP := fields[dnsmasqIPIndex]
			mac := fields[dnsmasqMACAddressIndex]
			ipToMAC[IP] = mac
		}
	}

	return ipToMAC, nil
}

func (d *DUTServicesImpl) pingDUTs(ctx context.Context, potentialIPs []string) ([]string, error) {
	// Use fping to figure out which IPs are active
	args := []string{"-a", "-t200", "-B1.0", "-r2"}
	args = append(args, potentialIPs...)

	out, err := d.commandExecutor.Exec(exec.Command(paths.Fping, args...))

	if err != nil {
		xerr, ok := err.(*exec.ExitError)
		// For reference:
		// fping will return exit status 1 if some hosts were unreachable.
		// https://fping.org/fping.1.html
		if !ok || xerr.ExitCode() != 1 {
			return []string{}, err
		}
	}

	rawData := strings.Split(string(out), "\n")
	activeIPs := []string{}

	for _, row := range rawData {
		if subnetSearchRe.MatchString(row) {
			matches := subnetSearchRe.FindStringSubmatch(row)
			IPIndex := subnetSearchRe.SubexpIndex("IP")
			activeIPs = append(activeIPs, matches[IPIndex])
		}
	}

	return activeIPs, nil
}

// GetConnectedIPs get the connected IPs from `dnsmasq.lease`
// and then check the IPs are alive.
func (d *DUTServicesImpl) GetConnectedIPs(ctx context.Context) ([]Device, error) {
	// This will list all IPs from a leases file
	ipToMACMap, err := d.fetchLeasesFile()
	if err != nil {
		return []Device{}, err
	}

	// Try to ping the IPs and get the active IPs
	potentialIPs := []string{}
	for IP := range ipToMACMap {
		potentialIPs = append(potentialIPs, IP)
	}
	activeIPs, err := d.pingDUTs(ctx, potentialIPs)
	if err != nil {
		return []Device{}, err
	}
	inactiveIPs := collection.Subtract(potentialIPs, activeIPs, func(a, b string) bool {
		return a == b
	})

	// We need to send a command to make sure ssh connection is avaliable.
	// Some DUTs can be pingable, but they can't establish the ssh connection.
	res := d.RunCommandOnIPs(ctx, activeIPs, constants.CheckDUTIsConnectedCommand)

	result := []Device{}
	for _, r := range res {
		macAddress := ipToMACMap[r.IP]
		if r.Error != nil {
			result = append(result, Device{IP: r.IP, IsConnected: false, MACAddress: macAddress})
		} else {
			// we check the some DUTs which install the stable image but they can
			// open the ssh connection.
			hasTestImage := strings.Contains(strings.ToLower(r.Value), constants.ChromeosTestImageReleaseTrack)
			result = append(result, Device{IP: r.IP, IsConnected: hasTestImage, MACAddress: macAddress})
		}
	}

	for _, r := range inactiveIPs {
		macAddress := ipToMACMap[r]
		result = append(result, Device{IP: r, IsConnected: false, MACAddress: macAddress})
	}

	return result, nil
}
