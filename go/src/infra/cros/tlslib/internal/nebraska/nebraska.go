// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

// Package nebraska implements a fake Omaha server based on "nebraska.py".
package nebraska

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

// Process represents an OS process.
// It is mainly for unit tests.
type Process interface {
	Args() []string
	Stop() error
}

// Environment is the runtime dependencies, e.g. networking, etc. of the
// implementation. The main goal of it is for unit test.
type Environment interface {
	// DownloadMetadata downloads update metadata for specified build and type
	// of payloads from GS to directory "dir" and returns the path of it.
	// It is the caller's responsibility to remove the temporary directory after
	// use.
	DownloadMetadata(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, dir string) (string, error)
	StartNebraska([]string) (Process, error)
}

// NewEnvironment returns a new instance of Environment that talks to GS and
// runs a real nebraska process.
func NewEnvironment() Environment {
	return &env{runCmd: exec.CommandContext}
}

// Server represents a running instance of 'nebraska.py'.
type Server struct {
	proc                  Process
	port                  int
	runtimeRoot           string
	metadataDir           string
	updatePayloadsAddress string
	env                   Environment
}

// NewServer starts a Nebraska process and returns a new instance of Server.
// gsPathPrefix is the GS path to the build of the update, e.g.
// "gs://chromeos-image-archive/banjo-release/R90-13809.0.0". The update
// metadata must exist there, so we can download them by appending the file name
// to this prefix.
// updatePayloadsAddress is the cache server URL from which we can download
// payloads, e.g. "http://<server>:<port>/download/banjo-release/R90-13809.0.0".
func NewServer(ctx context.Context, env Environment, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, updatePayloadsAddress string) (*Server, error) {
	n := &Server{env: env, updatePayloadsAddress: updatePayloadsAddress}
	if err := n.start(ctx, gsPathPrefix, payloads); err != nil {
		return nil, fmt.Errorf("new Nebraska: %s", err)
	}
	return n, nil
}

// Config is the Nebraska configurations.
type Config struct {
	CriticalUpdate         bool `json:"critical_update"`
	ReturnNoupdateStarting int  `json:"return_noupdate_starting"`
}

// UpdateConfig configures the started Nebraska.
func (n *Server) UpdateConfig(c Config) error {
	j, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("update Nebraska config: %s", err)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/update_config", n.port)
	rsp, err := http.Post(url, "application/json", bytes.NewReader(j))
	if err != nil {
		return fmt.Errorf("update Nebraska config: %s", err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		msg, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("update Nebraska config: %s", err)
		}
		return fmt.Errorf("update Nebraska config: %s", msg)
	}
	log.Printf("Nebrasak is configured with %q", j)
	return nil
}

// Script is the path of nebraska.py.
// Currently we download and build 'nebraska.py' into drone images at this path.
// TODO(guocb): Package 'nebraska.py' as part of TLS implementation.
const Script = "/opt/tls/fake_omaha/nebraska.py"

func (n *Server) cmdline() []string {
	return []string{
		Script,
		// We must use port number 0 in order to ask OS to assign a port.
		"--port", "0",
		"--update-metadata", n.metadataDir,
		"--update-payloads-address", n.updatePayloadsAddress,
		"--runtime-root", n.runtimeRoot,
		"--log-file", n.logfile(),
	}
}

func (n *Server) start(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) error {
	if n.proc != nil {
		panic(fmt.Sprintf("%s already started: %#v", n.proc, n.proc.Args()))
	}
	rootTmpDir, err := createRootTempDir()
	if err != nil {
		return fmt.Errorf("start Nebraska: %s", err)
	}
	n.metadataDir, err = n.env.DownloadMetadata(ctx, gsPathPrefix, payloads, rootTmpDir)
	if err != nil {
		return fmt.Errorf("start Nebraska: %s", err)
	}
	n.runtimeRoot, err = ioutil.TempDir(rootTmpDir, "nebraska_runtime_")
	if err != nil {
		return fmt.Errorf("start Nebraska: create runtime root: %s", err)
	}

	n.proc, err = n.env.StartNebraska(n.cmdline())
	if err != nil {
		return fmt.Errorf("start Nebraska: %s", err)
	}
	log.Printf("%s started", n.proc)
	if err := n.checkPort(ctx); err != nil {
		n.Close()
		return fmt.Errorf("start Nebraska: %s", err)
	}
	log.Printf("Nebraska is listening on %d", n.port)
	return nil
}

// createRootTempDir creates a directory as the root of other temp
// files/directories.
// This is a singleton directory shared with all instances of Nebraska, so we
// don't need to remove it as cleanup.
// It prevents us from creating too many entries in /tmp which causes issues
// with Swarming (see b/193141845 for details).
func createRootTempDir() (string, error) {
	dir := path.Join(os.TempDir(), "fake_omaha")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create root temp dir: %s", err)
	}
	return dir, nil
}

// Close terminates the nebraska server process and cleans up all temp
// dirs/files.
// This function is not concurrency safe.
func (n *Server) Close() error {
	if n.proc == nil {
		return fmt.Errorf("close Nebraska: process has been terminated")
	}
	log.Printf("Closing %s: %q", n.proc, n.cmdline())
	errs := []string{}
	if err := n.proc.Stop(); err != nil {
		errs = append(errs, fmt.Sprintf("stop process: %s", err))
	}
	if err := os.RemoveAll(n.metadataDir); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska metadata dir: %s", err))
	}
	nLog, err := ioutil.ReadFile(n.logfile())
	if err != nil {
		log.Printf("Cannot read %s log: %s", n.proc, err)
	} else {
		log.Printf("%s log: %s", n.proc, nLog)
	}
	if err := os.RemoveAll(n.runtimeRoot); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska runtime root: %s", err))
	}
	n.proc, n.port = nil, 0
	if len(errs) != 0 {
		return fmt.Errorf("close Nebraska: %s", strings.Join(errs, ", "))
	}
	return nil
}

// Port returns the port of the Nebraska.
func (n *Server) Port() int {
	return n.port
}

// checkPort checks the "port" file dropped by Nebraska in its runtime root
// directory and sets the "Server.port" accordingly.
func (n *Server) checkPort(ctx context.Context) error {
	const portFile = "port"
	filepath := path.Join(n.runtimeRoot, portFile)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	sPort, err := readFileOrTimeout(ctx, filepath)
	if err != nil {
		return fmt.Errorf("check port: %s", err)
	}
	p, err := strconv.Atoi(sPort)
	if err != nil {
		return fmt.Errorf("check port: %s", err)
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
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if cnt, err := ioutil.ReadFile(filepath); err == nil {
				return string(cnt), nil
			}
		case <-ctx.Done():
			return "", fmt.Errorf("read file %q: %s", filepath, ctx.Err())
		}
	}
}

type env struct {
	runCmd func(context.Context, string, ...string) *exec.Cmd
}

func (e env) DownloadMetadata(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, dir string) (string, error) {
	paths := metadataGSPaths(gsPathPrefix, payloads)
	log.Printf("New Nebraska: metadata to download: %#v", paths)
	metadataDir, err := ioutil.TempDir(dir, "AU_metadata_")
	if err != nil {
		return "", fmt.Errorf("download metadata: %s", err)
	}

	// Download Autoupdate metadata from Google Storage.
	// We cannot use CacheForDut TLW API since we download to Drones.
	cmd := []string{"gsutil", "cp", strings.Join(paths, " "), metadataDir}
	if err := e.runCmd(ctx, cmd[0], cmd[1:]...).Run(); err != nil {
		os.RemoveAll(metadataDir)
		return "", fmt.Errorf("download metadata: cmd: %s: %s", strings.Join(cmd, " "), err)
	}
	log.Printf("Start Nebraska: metadata downloaded to %q", metadataDir)
	return metadataDir, nil
}

func (e env) StartNebraska(cmdline []string) (Process, error) {
	log.Printf("Nebraska command line: %v", cmdline)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	var buf bytes.Buffer
	cmd.Stderr = &buf
	cmd.Stdout = &buf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start Nebraska: %s", err)
	}
	p := &proc{cmd: cmd, terminated: make(chan struct{})}
	go func() {
		p.cmd.Wait()
		log.Printf("%s output: %q", p, &buf)
		close(p.terminated)
	}()
	return p, nil
}

type proc struct {
	cmd        *exec.Cmd
	terminated chan struct{}
}

func (p proc) Stop() error {
	pid := p.cmd.Process.Pid
	if err := unix.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop %s: %s", p, err)
	}
	select {
	case <-p.terminated:
		log.Printf("%s was exited", p)
	case <-time.After(2 * time.Second):
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("kill %s: %s", p, err)
		}
		log.Printf("%s was killed", p)
	}
	return nil
}

func (p proc) Args() []string {
	return p.cmd.Args
}

func (p proc) String() string {
	return fmt.Sprintf("Nebraska (pid: %d)", p.cmd.Process.Pid)
}

const (
	// Metadata files name pattern in GS wildcard chars. Please keep it sync
	// with https://chromium.googlesource.com/chromiumos/chromite/+/e55168c7e07cebc82dd6aa227c8e87201eb6766c/lib/xbuddy/build_artifact.py#586
	fullPayloadPattern  = "chromeos_*_full_dev*bin.json"
	deltaPayloadPattern = "chromeos_*_delta_dev*bin.json"
)

func metadataGSPaths(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) []string {
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
		default:
			panic(fmt.Sprintf("CreateFakeOmaha: unrecognized payload type: %s", t))
		}
	}
	return r
}
