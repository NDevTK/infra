// Copyright 2022 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// in the test
const (
	nsjailLogTestKey = "holds an *os.File"
	testNsjailLog    = "nsjailLog"
)

func init() {
	setupNsJailLog = func(ctx context.Context) (*os.File, error) {
		retFile, ok := ctx.Value(nsjailLogTestKey).(*os.File)
		if !ok {
			return nil, errors.New("no nsjaillog in test!")
		}
		return retFile, nil
	}
}

// TestHelperProcess isn't a real test
// Inspired by: https://github.com/golang/go/blob/master/src/os/exec/exec_test.go#L758
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}
	cmd, _ := args[0], args[1:]
	if !strings.HasSuffix(cmd, "nsjail") {
		fmt.Fprintf(os.Stderr, "non-nsjail command: %s\n", cmd)
	}
}
func fakeExecCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestRunInNsjail(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows because this will not be built or deployed for Windows.")
	}

	ctx := context.Background()
	f, err := os.CreateTemp("", testNsjailLog)
	if err != nil {
		log.Fatal(err)
	}

	ctx = context.WithValue(ctx, nsjailLogTestKey, f)

	Convey("basic command tries to run nsjail", t, func() {
		err := RunInNsjail(ctx, []string{"cat", "hello world"})
		// // override exec.Command
		execCommand = fakeExecCommand
		defer func() { execCommand = exec.CommandContext }()
		So(err.Error(), ShouldContainSubstring, "nsjail: no such file or directory")
	})

	defer os.Remove(testNsjailLog)
}
