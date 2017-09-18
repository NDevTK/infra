// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/common/system/exitcode"
	"go.chromium.org/luci/common/system/filesystem"

	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testDataDir          = "test_data"
	testMainRunScriptENV = "_VPYTHON_MAIN_TEST_RUN_SCRIPT"
)

var vpythonDebug = flag.Bool("vpython.debug", false, "Enable vpython debug otuput.")

func TestMain(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("could not get executable path: %s", err)
	}

	// Are we a spawned subprocess of TestMain?
	env := environ.System()
	if v := env.GetEmpty(testMainRunScriptENV); v != "" {
		os.Exit(testMainRunDelegate(self, v))
		return
	}

	testCases := loadTestCases(t, self)

	// Execute each test case in parallel.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, tc.run)
	}
}

type testDelegateParams struct {
	Args []string
}

type testDelegateCommand struct {
	*exec.Cmd

	output bytes.Buffer
	tc     *testCase
	params testDelegateParams
}

func (tdc *testDelegateCommand) prepare() {
	base := tdc.Env
	if len(base) == 0 {
		base = os.Environ()
	}
	env := environ.New(base)
	env.Set(testMainRunScriptENV, encodeEnvironmentParam(&tdc.params))
	tdc.Env = env.Sorted()
}

func (tdc *testDelegateCommand) Start() error {
	tdc.prepare()
	return tdc.Cmd.Start()
}

func (tdc *testDelegateCommand) Run(t *testing.T) error {
	tdc.prepare()
	err := tdc.Cmd.Run()
	t.Logf("Test output for %q:\n%s", tdc.tc.name, tdc.output.Bytes())
	return err
}

func (tdc *testDelegateCommand) Wait(t *testing.T) error {
	err := tdc.Cmd.Wait()
	t.Logf("Test output for %q:\n%s", tdc.tc.name, tdc.output.Bytes())
	return err
}

func (tdc *testDelegateCommand) CheckOutput(t *testing.T) bool {
	matches := bytes.Equal(tdc.output.Bytes(), tdc.tc.output)
	if !matches {
		t.Errorf("Outputs do not match. Expected:\n%s", tdc.tc.output)
	}
	return matches
}

type testCase struct {
	self   string
	name   string
	script string
	output []byte
}

func loadTestCases(t *testing.T, self string) []testCase {
	var testCases []testCase
	testCaseErrors := 0
	fis, err := ioutil.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("could not read test directory %q: %v", testDataDir, err)
	}
	for _, fi := range fis {
		ext := filepath.Ext(fi.Name())
		if ext != ".py" {
			continue
		}

		script := filepath.Join(testDataDir, fi.Name())
		base := script[:len(script)-len(ext)]
		outputPath := base + ".output"

		content, err := ioutil.ReadFile(outputPath)
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("could not load output for %q at %q: %v", script, outputPath, err)
			testCaseErrors++
			continue
		}

		testCases = append(testCases, testCase{
			self:   self,
			name:   filepath.Base(base),
			script: script,
			output: content,
		})
	}
	switch {
	case testCaseErrors > 0:
		t.Fatalf("errors encountered while loading test cases")
	case len(testCases) == 0:
		t.Fatalf("no test cases found")
	}

	return testCases
}

func (tc *testCase) String() string { return tc.name }

func (tc *testCase) getDelegateCommand(c context.Context, root string) *testDelegateCommand {
	args := []string{
		"-vpython-root", root,
	}
	if *vpythonDebug {
		args = append(args, "-log-level", "debug")
	}
	args = append(args, "-u", tc.script)

	tdc := testDelegateCommand{
		tc: tc,
		params: testDelegateParams{
			Args: args,
		},
	}

	tdc.Cmd = exec.CommandContext(c, tdc.tc.self, "-test.run", "^TestMain$")
	tdc.Stdout = &tdc.output
	tdc.Stderr = os.Stderr

	return &tdc
}

func (tc *testCase) run(t *testing.T) {
	t.Parallel()

	Convey(fmt.Sprintf(`Testing %q`, tc), t, func() {
		// Run our test suite with a tempdir for our vpython root.
		tdir := filesystem.TempDir{
			Dir:    filepath.Dir(tc.script),
			Prefix: "vpython_main_test",
			CleanupErrFunc: func(tdir string, err error) {
				t.Logf("(Non-fatal) could not remove tempdir %q: %v", tdir, err)
			},
		}
		err := tdir.With(func(td string) error {
			switch tc.name {
			case "test_signals":
				tc.runTestSignals(t, td)
			case "test_exit_code":
				tc.runCommon(t, td, 42)
			default:
				tc.runCommon(t, td, 0)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("could not create tempdir: %v", err)
		}
	})
}

func (tc *testCase) runCommon(t *testing.T, td string, exitCode int) {
	// Kill the process after 10 seconds.
	//
	// This can be increased if network tests are taking too long to fetch
	// resources.
	c, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	tdc := tc.getDelegateCommand(c, td)

	err := tdc.Run(t)
	if rc, ok := exitcode.Get(err); ok {
		So(rc, ShouldEqual, exitCode)
	} else {
		So(err, ShouldBeNil)
	}
	So(tdc.CheckOutput(t), ShouldBeTrue)
}

func (tc *testCase) runTestSignals(t *testing.T, td string) {
	// Kill the process after 10 seconds.
	//
	// This can be increased if network tests are taking too long to fetch
	// resources.
	c, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	tdc := tc.getDelegateCommand(c, td)

	startedC := make(chan struct{})
	handlerDoneC := make(chan struct{})

	// After the process starts, send a SIGTERM signal.
	signalR, signalW, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create signal pipe")
	}
	tdc.ExtraFiles = []*os.File{signalW}

	go func() {
		defer close(handlerDoneC)

		<-startedC
		signalW.Close()
		t.Logf("Waiting for writer pipe to close...")
		_, _ = ioutil.ReadAll(signalR)

		t.Log("Sending signal...")
		if err := tdc.Process.Signal(os.Interrupt); err != nil {
			t.Logf("failed to signal process")
		}
	}()

	err = tdc.Start()
	if err != nil {
		t.Fatalf("Failed to start subprocess: %s", err)
	}

	// Notify that the process has started.
	close(startedC)

	err = tdc.Wait(t)
	<-handlerDoneC
	So(err, ShouldBeNil)
	So(tdc.CheckOutput(t), ShouldBeTrue)
}

func testMainRunDelegate(self, v string) int {
	var p testDelegateParams
	if err := decodeEnvironmentParam(v, &p); err != nil {
		log.Fatalf("could not decode enviornment param %q: %s", v, err)
	}

	argv := make([]string, 1, len(p.Args)+1)
	argv[0] = self
	argv = append(argv, p.Args...)

	c := context.Background()
	return mainImpl(c, argv)
}

func encodeEnvironmentParam(i interface{}) string {
	d, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(d)
}

func decodeEnvironmentParam(v string, i interface{}) error {
	d, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, i)
}
