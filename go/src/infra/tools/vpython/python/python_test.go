// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package python

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"golang.org/x/net/context"

	. "github.com/luci/luci-go/common/testing/assertions"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParsePythonCommandLine(t *testing.T) {
	t.Parallel()

	successes := []struct {
		args []string
		cmd  CommandLine
	}{
		{nil, CommandLine{}},

		{[]string{"-a", "-b", "-Q'foo.bar.baz'", "-Wbar"},
			CommandLine{
				Flags: []string{"-a", "-b", "-Q'foo.bar.baz'", "-Wbar"},
				Args:  []string{},
			},
		},

		{[]string{"path.py", "--", "foo", "bar"},
			CommandLine{
				Type:  TargetScript,
				Value: "path.py",
				Flags: []string{},
				Args:  []string{"--", "foo", "bar"},
			},
		},

		{[]string{"-a", "-Wfoo", "-", "--", "foo"},
			CommandLine{
				Type:  TargetScript,
				Value: "-",
				Flags: []string{"-a", "-Wfoo"},
				Args:  []string{"--", "foo"},
			},
		},

		{[]string{"-a", "-b", "-W", "foo", "-Wbar", "-c", "<script>", "--", "arg"},
			CommandLine{
				Type:  TargetCommand,
				Value: "<script>",
				Flags: []string{"-a", "-b", "-W", "foo", "-Wbar"},
				Args:  []string{"--", "arg"},
			},
		},

		{[]string{"-a", "-b", "-m'foo.bar.baz'", "arg"},
			CommandLine{
				Type:  TargetModule,
				Value: "'foo.bar.baz'",
				Flags: []string{"-a", "-b"},
				Args:  []string{"arg"},
			},
		},
	}

	failures := []struct {
		args []string
		err  string
	}{
		{[]string{"-a", "-b", "-Q"}, "truncated two-variable argument"},
		{[]string{"-c"}, "missing second value"},
		{[]string{"-\x80"}, "invalid rune in flag"},
	}

	Convey(`Testing Python command-line parsing`, t, func() {
		for _, tc := range successes {
			Convey(fmt.Sprintf(`Success cases: %v`, tc.args), func() {
				cmd, err := ParseCommandLine(tc.args)
				So(err, ShouldBeNil)
				So(cmd, ShouldResemble, tc.cmd)
			})
		}

		for _, tc := range failures {
			Convey(fmt.Sprintf(`Error cases: %v`, tc.args), func() {
				_, err := ParseCommandLine(tc.args)
				So(err, ShouldErrLike, tc.err)
			})
		}
	})
}

func TestInterpreter(t *testing.T) {
	t.Parallel()

	versionSuccesses := []struct {
		output string
		vers   Version
	}{
		{"Python 2.7.1\n", Version{2, 7, 1}},
		{"Python 3", Version{3, 0, 0}},
	}

	versionFailures := []struct {
		output string
		err    string
	}{
		{"", "unknown version output"},
		{"Python2.7.11\n", "unknown version output"},
		{"Python", "unknown version output"},
		{"Python 2.7.11 foo bar junk", "invalid number value"},
		{"Python 3.1.2.3.4", "failed to parse version"},
	}

	Convey(`A Python interpreter`, t, func() {
		c := context.Background()

		var (
			runnerOutput string
			runnerErr    error
			lastCmd      *exec.Cmd
		)
		i := Interpreter{
			runner: func(cmd *exec.Cmd, capture bool) (string, error) {
				// Sanitize exec.Cmd to exclude its private members for comparison.
				lastCmd = &exec.Cmd{
					Path:   cmd.Path,
					Args:   cmd.Args,
					Stdin:  cmd.Stdin,
					Stdout: cmd.Stdout,
					Stderr: cmd.Stderr,
					Env:    cmd.Env,
					Dir:    cmd.Dir,
				}
				return runnerOutput, runnerErr
			},
		}

		Convey(`Will error if no Python interpreter is supplied.`, func() {
			So(i.Run(c, "foo", "bar"), ShouldErrLike, "a Python interpreter must be supplied")
		})

		Convey(`With a Python interpreter installed`, func() {
			i.Python = "/path/to/python"

			Convey(`Testing Run`, func() {
				Convey(`Can run a command.`, func() {
					So(i.Run(c, "foo", "bar"), ShouldBeNil)
					So(lastCmd, ShouldResemble, &exec.Cmd{
						Path:   "/path/to/python",
						Args:   []string{"/path/to/python", "foo", "bar"},
						Stdout: os.Stdout,
						Stderr: os.Stderr,
					})
				})

				Convey(`Can run an isolated environment command.`, func() {
					i.Isolated = true

					So(i.Run(c, "foo", "bar"), ShouldBeNil)
					So(lastCmd, ShouldResemble, &exec.Cmd{
						Path:   "/path/to/python",
						Args:   []string{"/path/to/python", "-B", "-E", "-s", "foo", "bar"},
						Stdout: os.Stdout,
						Stderr: os.Stderr,
					})
				})

				Convey(`Can connect STDIN.`, func() {
					i.ConnectSTDIN = true

					So(i.Run(c, "foo", "bar"), ShouldBeNil)
					So(lastCmd, ShouldResemble, &exec.Cmd{
						Path:   "/path/to/python",
						Args:   []string{"/path/to/python", "foo", "bar"},
						Stdout: os.Stdout,
						Stderr: os.Stderr,
						Stdin:  os.Stdin,
					})
				})

				Convey(`Will forward a working directory and enviornment`, func() {
					i.WorkDir = "zugzug"
					i.Env = []string{"pants=on"}

					So(i.Run(c, "foo", "bar"), ShouldBeNil)
					So(lastCmd, ShouldResemble, &exec.Cmd{
						Path:   "/path/to/python",
						Args:   []string{"/path/to/python", "foo", "bar"},
						Stdout: os.Stdout,
						Stderr: os.Stderr,
						Env:    []string{"pants=on"},
						Dir:    "zugzug",
					})
				})
			})

			Convey(`Testing GetVersion`, func() {
				for _, tc := range versionSuccesses {
					Convey(fmt.Sprintf(`Can successfully parse %q => %s`, tc.output, tc.vers), func() {
						runnerOutput = tc.output
						vers, err := i.GetVersion(c)
						So(err, ShouldBeNil)
						So(vers, ShouldResemble, tc.vers)
					})
				}

				for _, tc := range versionFailures {
					Convey(fmt.Sprintf(`Will fail to parse %q (%s)`, tc.output, tc.err), func() {
						runnerOutput = tc.output
						_, err := i.GetVersion(c)
						So(err, ShouldErrLike, tc.err)
					})
				}
			})
		})
	})
}
