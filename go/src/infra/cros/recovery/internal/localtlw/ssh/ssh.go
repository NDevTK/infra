// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"

	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
	"infra/libs/sshpool"
)

const (
	defaultSSHUser = "root"
	DefaultPort    = 22
)

// SSHConfig provides default config for SSH.
func SSHConfig(sshKeyPaths []string) *ssh.ClientConfig {
	keySigners := getKeySigners(sshKeyPaths)
	return &ssh.ClientConfig{
		User:            defaultSSHUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(keySigners...)},
		// The timeout specified to established connection only.
		// That is not an execution timeout.
		Timeout: 2 * time.Second,
	}
}

// Run executes command on the target address by SSH.
func Run(ctx context.Context, pool *sshpool.Pool, addr string, cmd string) (result *tlw.RunResult) {
	result = &tlw.RunResult{
		Command:  cmd,
		ExitCode: -1,
	}
	if pool == nil {
		result.Stderr = "run SSH: pool is not initialized"
		return
	} else if addr == "" {
		result.Stderr = "run SSH: addr is empty"
		return
	} else if cmd == "" {
		result.Stderr = fmt.Sprintf("run SSH %q: cmd is empty", addr)
		return
	}
	sc, err := pool.GetContext(ctx, addr)
	if err != nil {
		result.Stderr = fmt.Sprintf("run SSH %q: fail to get client from pool; %s", addr, err)
		return
	}
	defer func() {
		pool.Put(addr, sc)
		log.Debugf(ctx, "Finished update SSH pool for %q!", addr)
	}()
	result = createSessionAndExecute(ctx, cmd, sc)
	log.Debugf(ctx, "Run SSH %q: Cmd: %q", addr, result.Command)
	log.Debugf(ctx, "Run SSH %q: ExitCode: %d", addr, result.ExitCode)
	log.Debugf(ctx, "Run SSH %q: Stdout: %s", addr, result.Stdout)
	log.Debugf(ctx, "Run SSH %q: Stderr: %s", addr, result.Stderr)
	return result
}

// createSessionAndExecute creates ssh session and perform execution by ssh.
//
// The function also aborted execution if context canceled.
func createSessionAndExecute(ctx context.Context, cmd string, client *ssh.Client) (result *tlw.RunResult) {
	result = &tlw.RunResult{
		Command:  cmd,
		ExitCode: -1,
	}
	session, err := client.NewSession()
	if err != nil {
		result.Stderr = fmt.Sprintf("internal run ssh: %v", err)
		return
	}
	defer func() {
		session.Close()
	}()
	var stdOut, stdErr bytes.Buffer
	session.Stdout = &stdOut
	session.Stderr = &stdErr
	exit := func(err error) *tlw.RunResult {
		result.Stdout = stdOut.String()
		result.Stderr = stdErr.String()
		switch t := err.(type) {
		case nil:
			result.ExitCode = 0
		case *ssh.ExitError:
			result.ExitCode = int32(t.ExitStatus())
		case *ssh.ExitMissingError:
			result.ExitCode = -2
			result.Stderr = t.Error()
		default:
			// Set error 1 as not expected exit.
			result.ExitCode = -3
			result.Stderr = t.Error()
		}
		return result
	}
	// Chain to run ssh in separate thread and wait for single response from it.
	// If context will be closed before it will abort the session.
	sw := make(chan bool, 1)
	var runErr error
	go func() {
		runErr = session.Run(cmd)
		sw <- true
	}()
	select {
	case <-sw:
		return exit(runErr)
	case <-ctx.Done():
		// At the end abort session.
		// Session will be closed in defer.
		session.Signal(ssh.SIGABRT)
		return exit(ctx.Err())
	}
}
