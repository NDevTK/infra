// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package nebraska

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/chromiumos/config/go/api/test/tls"
)

// Environment is the runtime dependencies, e.g. networking, etc. of the
// implementation. The main goal of it is for unit test.
type Environment struct {
	fileReader     func(string) ([]byte, error)
	httpDownloader func(string) (*http.Response, error)
	commander      func(string, ...string) *exec.Cmd
}

// NewEnvironment returns a new instance of Environment.
func NewEnvironment() *Environment {
	return &Environment{
		fileReader:     ioutil.ReadFile,
		httpDownloader: http.Get,
		commander:      exec.Command,
	}
}

// Server represents a process of 'nebraska.py'.
type Server struct {
	Port int

	cmd         *exec.Cmd
	runtimeRoot string
	metadataDir string
	env         *Environment
}

// NewServer returns a new instance of Server.
func NewServer(env *Environment) *Server {
	return &Server{env: env}
}

// Start starts a nebraska server.
// This function is not concurrency safe.
// This function can only be called once.
func (n *Server) Start(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload, payloadsURL string) error {
	if n.cmd != nil {
		return fmt.Errorf("Nebraska has been started: (%d) %#v", n.cmd.Process.Pid, n.cmd.Args)
	}
	if err := n.downloadMetadata(gsPathPrefix, payloads); err != nil {
		return fmt.Errorf("fetch metadata: %s", err)
	}
	var err error
	n.runtimeRoot, err = ioutil.TempDir("", "nebraska_runtimeroot_")
	if err != nil {
		return fmt.Errorf("create runtime root: %s", err)
	}
	// TODO (guocb): Figure out a better way to get 'nebraska.py'.
	if err := n.downloadScript(); err != nil {
		return fmt.Errorf("fetch Nebraksa: %s", err)
	}

	return n.launch(payloadsURL)
}

func (n *Server) launch(payloadsURL string) error {
	cmdline := []string{
		n.script(),
		// We must use port number 0 in order to ask OS to assign a port.
		"--port", "0",
		"--update-metadata", n.metadataDir,
		"--update-payloads-address", payloadsURL,
		"--runtime-root", n.runtimeRoot,
		"--log-file", n.logfile(),
	}
	log.Printf("Nebraska command line: %v\n", cmdline)
	cmd := n.env.commander("python3", cmdline...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start Nebraska: %s", err)
	}
	n.cmd = cmd
	go cmd.Wait()
	if err := n.waitPortFile(); err != nil {
		return fmt.Errorf("read Nebraska port: %s", err)
	}
	return nil
}

// Close terminates the nebraska server process and cleans up all temp
// dirs/files.
// This function is not concurrency safe.
// This function can only be called once.
func (n *Server) Close() error {
	if n.cmd == nil {
		return fmt.Errorf("Nebraska process has been terminated")
	}
	errs := []string{}
	if err := n.cmd.Process.Kill(); err != nil {
		errs = append(errs, fmt.Sprintf("kill Nebraska process: %s", err))
	}
	if err := os.RemoveAll(n.metadataDir); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska metadata dir: %s", err))
	}
	if err := os.RemoveAll(n.runtimeRoot); err != nil {
		errs = append(errs, fmt.Sprintf("remove Nebraska runtime root: %s", err))
	}
	n.cmd = nil
	n.Port = 0
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errs, "|"))
}

func (n *Server) waitPortFile() error {
	const portFile = "port"
	filepath := path.Join(n.runtimeRoot, portFile)
	ch := make(chan string, 1)
	go func() {
		for {
			if d, err := n.env.fileReader(filepath); err == nil {
				ch <- string(d)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	timeout := 2 * time.Second
	select {
	case <-time.After(timeout):
		return fmt.Errorf("Timeout(%s) when wait file %q", timeout, filepath)
	case cnt := <-ch:
		p, err := strconv.Atoi(cnt)
		if err == nil {
			n.Port = p
		}
		return err
	}
}

func (n *Server) downloadMetadata(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) error {
	md, err := getMetadataPaths(gsPathPrefix, payloads)
	if err != nil {
		return err
	}
	metadataDir, err := ioutil.TempDir("", "AU_metadata_")
	if err != nil {
		return fmt.Errorf("failed to create a temp dir for metadata: %s", err)
	}

	// Download Autoupdate metadata.
	// We cannot use CacheForDut API since we download to Drones.
	cmd := []string{"gsutil", "cp", strings.Join(md, " "), metadataDir}

	if err := n.env.commander(cmd[0], cmd[1:]...).Run(); err != nil {
		os.RemoveAll(metadataDir)
		return fmt.Errorf("downlaod metadata: cmd: %s: %s", strings.Join(cmd, " "), err)
	}
	log.Printf("Metadata downloaded to %q", metadataDir)
	n.metadataDir = metadataDir
	return nil
}

func (n *Server) logfile() string {
	return path.Join(n.runtimeRoot, "nebraska.log")
}

func (n *Server) script() string {
	return path.Join(n.runtimeRoot, "nebraska.py")
}

func (n *Server) downloadScript() error {
	const url = "https://chromium.googlesource.com/chromiumos/platform/dev-util/+/master/nebraska/nebraska.py?format=TEXT"
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch Nebraska: bad http status: %s", rsp.Status)
	}

	fd, err := os.Create(n.script())
	if err != nil {
		return fmt.Errorf("fetch Nebraska: create script: %s", err)
	}
	// The file downloaded is base64 encoded.
	decoder := base64.NewDecoder(base64.StdEncoding, rsp.Body)
	_, err = io.Copy(fd, decoder)
	if err != nil {
		return fmt.Errorf("fetch Nebraska: write to %q: %s", n.script(), err)
	}
	return nil
}

func getMetadataPaths(gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) ([]string, error) {
	const (
		// Metadata files name pattern in GS wildcard chars. Please keep it sync
		// with http://cs/chromeos_public/chromite/lib/xbuddy/build_artifact.py;rcl=e55168c7e07cebc82dd6aa227c8e87201eb6766c;l=586
		fullPayloadPattern  = "chromeos_*_full_dev*bin.json"
		deltaPayloadPattern = "chromeos_*_delta_dev*bin.json"
	)
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
