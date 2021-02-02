// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package nebraska implements a fake Omaha server based on "nebraska.py".
package nebraska

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.chromium.org/chromiumos/config/go/api/test/tls"
	"golang.org/x/sys/unix"
)

// Environment is the runtime dependencies, e.g. networking, etc. of the
// implementation. The main goal of it is for unit test.
type Environment interface {
	DownloadMetadata(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) (string, error)
	StartNebraska([]string) (*exec.Cmd, error)
}

// NewEnvironment returns a new instance of Environment.
func NewEnvironment() Environment {
	return &env{runCmd: exec.CommandContext}
}

// Server represents a process of 'nebraska.py'.
type Server struct {
	cmd         *exec.Cmd
	port        int
	runtimeRoot string
	metadataDir string
	env         Environment
}

// NewServer starts a Nebraska process and returns a new instance of Server.
func NewServer(ctx context.Context, env Environment, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, payloadsServerURL string) (*Server, error) {
	n := &Server{env: env}
	if err := n.start(ctx, gsPathPrefix, payloads, payloadsServerURL); err != nil {
		return nil, err
	}
	return n, nil
}

// NebraskaScript is the path of nebraska.py.
// Currently We download and build 'nebraska.py' into drone images at this path.
// TODO(guocb): Package 'nebraska.py' as part of TLS implementation.
const NebraskaScript = "/opt/tls/fake_omaha/nebraska.py"

func (n *Server) start(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, payloadsServerURL string) error {
	if n.cmd != nil {
		return fmt.Errorf("Nebraska already started: (%d) %#v", n.cmd.Process.Pid, n.cmd.Args)
	}
	var err error
	if n.metadataDir, err = n.env.DownloadMetadata(ctx, gsPathPrefix, payloads); err != nil {
		return fmt.Errorf("download metadata: %s", err)
	}
	n.runtimeRoot, err = ioutil.TempDir("", "nebraska_runtimeroot_")
	if err != nil {
		return fmt.Errorf("create runtime root: %s", err)
	}
	cmdline := []string{
		NebraskaScript,
		// We must use port number 0 in order to ask OS to assign a port.
		"--port", "0",
		"--update-metadata", n.metadataDir,
		"--update-payloads-address", payloadsServerURL,
		"--runtime-root", n.runtimeRoot,
		"--log-file", n.logfile(),
	}
	log.Printf("Nebraska command line: %v\n", cmdline)

	if n.cmd, err = n.env.StartNebraska(cmdline); err != nil {
		return err
	}
	log.Printf("Nebraska started (pid: %d)", n.cmd.Process.Pid)
	if err = n.checkPort(); err != nil {
		n.Close()
		return fmt.Errorf("check Nebraska port: %s", err)
	}
	log.Printf("Nebraska is listening on %d", n.port)
	return nil
}

// Close terminates the nebraska server process and cleans up all temp
// dirs/files.
// This function is not concurrency safe.
func (n *Server) Close() error {
	if n.cmd == nil {
		return fmt.Errorf("Nebraska process has been terminated")
	}
	log.Printf("Closing Nebraska (pid: %d) %q", n.cmd.Process.Pid, n.cmd)
	errs := []string{}
	if err := unix.Kill(n.cmd.Process.Pid, syscall.SIGTERM); err != nil {
		errs = append(errs, fmt.Sprintf("terminate Nebraska process: %s", err))
	}
	if err := os.RemoveAll(n.metadataDir); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska metadata dir: %s", err))
	}
	if err := os.RemoveAll(n.runtimeRoot); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska runtime root: %s", err))
	}
	n.cmd = nil
	n.port = 0
	if len(errs) != 0 {
		return fmt.Errorf(strings.Join(errs, ", "))
	}
	return nil
}

// Port returns the port of the Nebraska.
func (n *Server) Port() int {
	return n.port
}

// checkPort checks the "port" file dropped by Nebraska in its runtime root
// directory and sets the "Server.port" accordingly.
func (n *Server) checkPort() error {
	const portFile = "port"
	filepath := path.Join(n.runtimeRoot, portFile)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sPort, err := readFileOrTimeout(ctx, filepath)
	if err != nil {
		return err
	}
	p, err := strconv.Atoi(sPort)
	if err != nil {
		return err
	}
	n.port = p
	return nil
}

func (n *Server) logfile() string {
	return path.Join(n.runtimeRoot, "nebraska.log")
}

// readFileOrTimeout reads a file to return its content, or timeout if the file
// is not ready before the deadline.
func readFileOrTimeout(ctx context.Context, filepath string) (string, error) {
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			if cnt, err := ioutil.ReadFile(filepath); err == nil {
				return string(cnt), nil
			}
			break
		case <-ctx.Done():
			return "", fmt.Errorf("read file %q: %s", filepath, ctx.Err())
		}
	}
}

type env struct {
	runCmd func(context.Context, string, ...string) *exec.Cmd
}

func (e env) DownloadMetadata(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) (string, error) {
	paths, err := metadataGsPaths(gsPathPrefix, payloads)
	if err != nil {
		return "", err
	}
	metadataDir, err := ioutil.TempDir("", "AU_metadata_")
	if err != nil {
		return "", fmt.Errorf("failed to create a temp dir for metadata: %s", err)
	}

	// Download Autoupdate metadata from Google Storage.
	// We cannot use CacheForDut TLW API since we download to Drones.
	cmd := []string{"gsutil", "cp", strings.Join(paths, " "), metadataDir}
	if err := e.runCmd(ctx, cmd[0], cmd[1:]...).Run(); err != nil {
		os.RemoveAll(metadataDir)
		return "", fmt.Errorf("download metadata: cmd: %s: %s", strings.Join(cmd, " "), err)
	}
	log.Printf("CreateFakeOmaha: metadata downloaded to %q", metadataDir)
	return metadataDir, nil
}

const (
	// Metadata files name pattern in GS wildcard chars. Please keep it sync
	// with https://chromium.googlesource.com/chromiumos/chromite/+/e55168c7e07cebc82dd6aa227c8e87201eb6766c/lib/xbuddy/build_artifact.py#586
	fullPayloadPattern  = "chromeos_*_full_dev*bin.json"
	deltaPayloadPattern = "chromeos_*_delta_dev*bin.json"
)

func metadataGsPaths(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) ([]string, error) {
	// We cannot use path.Join to format the path because it eliminate one of
	// two slashes of "gs://".
	// We cannot use "net/url" either because it escapes the wildcard chars.
	prefix := strings.TrimRight(gsPathPrefix, "/")
	r := []string{}
	for _, p := range payloads {
		switch t := p.GetType(); t {
		case tls.FakeOmaha_Payload_FULL:
			r = append(r, fmt.Sprintf("%s/%s", prefix, fullPayloadPattern))
		case tls.FakeOmaha_Payload_DELTA:
			r = append(r, fmt.Sprintf("%s/%s", prefix, deltaPayloadPattern))
		}
	}
	log.Printf("Metadata to download: %v\n", r)
	return r, nil
}

func (e env) StartNebraska(cmdline []string) (*exec.Cmd, error) {
	cmd := exec.Command("python3", cmdline...)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start Nebraska: %s", err)
	}
	go cmd.Wait()
	return cmd, nil
}
