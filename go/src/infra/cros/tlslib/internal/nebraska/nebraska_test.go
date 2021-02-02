// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package nebraska

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"testing"

	"go.chromium.org/chromiumos/config/go/api/test/tls"
)

type fakeEnv struct {
	startFakeNebraska func(*Server) (Process, error)
}

func (e fakeEnv) DownloadMetadata(ctx context.Context, gsPathPrefix string, payloads []*tls.FakeOmaha_Payload) (string, error) {
	return "", nil
}

func (e fakeEnv) StartNebraska(n *Server) (Process, error) {
	return e.startFakeNebraska(n)
}

var _ Environment = fakeEnv{}

type fakeProc struct{}

func (p fakeProc) Pid() int {
	return 0
}

func (p fakeProc) Stop() error {
	return nil
}

func (p fakeProc) Args() []string {
	return []string{"fake", "process"}
}

var _ Process = fakeProc{}

func TestNebraska(t *testing.T) {
	t.Parallel()
	const portWant = 12345
	port := []byte(strconv.Itoa(portWant))
	e := fakeEnv{
		startFakeNebraska: func(n *Server) (Process, error) {
			err := ioutil.WriteFile(path.Join(n.runtimeRoot, "port"), port, 0644)
			if err != nil {
				t.Fatalf("create fake port file: %s", err)
			}
			return &fakeProc{}, nil

		},
	}
	n, err := NewServer(context.Background(), e, "gs://", []*tls.FakeOmaha_Payload{}, "http://cache-server/update")
	if err != nil {
		t.Errorf("NewServer: failed to create Nebraska: %s", err)
	}
	if n.port != portWant {
		t.Errorf("NewServer: port: got %d, want %d", n.port, portWant)
	}

	if err = n.Close(); err != nil {
		t.Errorf("close Nebraska: %s", err)
	}
	if _, err := os.Stat(n.runtimeRoot); err == nil {
		t.Errorf("close Nebraska: runtime root was not removed")
	}
	if n.proc != nil {
		t.Errorf("close Nebraska: process was not terminated")
	}
}

func TestNebraskaTimeoutOnPort(t *testing.T) {
	t.Parallel()
	e := fakeEnv{
		startFakeNebraska: func(*Server) (Process, error) {
			return &fakeProc{}, nil
		},
	}
	_, err := NewServer(context.Background(), e, "gs://", []*tls.FakeOmaha_Payload{}, "http://cache-server/update")
	if err == nil {
		t.Fatalf("NewServer() succeeded without Nebraska port file, want error")
	}
}

func TestEnv_DownloadMetadata(t *testing.T) {
	t.Parallel()
	got := ""
	e := env{
		runCmd: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			got = fmt.Sprintf("%s %s", name, strings.Join(args, " "))
			return exec.Command("true")
		},
	}
	tests := []struct {
		name     string
		types    []*tls.FakeOmaha_Payload
		patterns []string
	}{
		{
			"full payload only",
			[]*tls.FakeOmaha_Payload{{Type: tls.FakeOmaha_Payload_FULL}},
			[]string{fullPayloadPattern},
		},
		{
			"delta payload only",
			[]*tls.FakeOmaha_Payload{{Type: tls.FakeOmaha_Payload_DELTA}},
			[]string{deltaPayloadPattern},
		},
		{
			"full and delta payload",
			[]*tls.FakeOmaha_Payload{{Type: tls.FakeOmaha_Payload_FULL}, {Type: tls.FakeOmaha_Payload_DELTA}},
			[]string{fullPayloadPattern, deltaPayloadPattern},
		},
	}
	gsPrefix := "gs://bucket/build/version"
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e.DownloadMetadata(context.Background(), gsPrefix, tc.types)
			w := []string{"gsutil cp"}
			for _, p := range tc.patterns {
				w = append(w, fmt.Sprintf("%s/%s", gsPrefix, p))
			}
			prefix := strings.Join(w, " ")
			if !strings.HasPrefix(got, prefix) {
				t.Errorf("DownloadMetadata(FULL) error: want prefix %q, got %q", prefix, got)
			}
		})
	}
}
