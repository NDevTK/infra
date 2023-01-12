// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	bb "infra/cros/internal/buildbucket"
	"infra/cros/internal/cmd"

	"github.com/golang/protobuf/jsonpb"
	"github.com/maruel/subcommands"
	pb "go.chromium.org/chromiumos/infra/proto/go/chromiumos"
	"go.chromium.org/luci/common/cli"
)

var (
	unmarshaler = jsonpb.Unmarshaler{AllowUnknownFields: true}
)

type collectRun struct {
	subcommands.CommandRunBase
	stdoutLog *log.Logger
	stderrLog *log.Logger
	cmdRunner cmd.CommandRunner
	bbClient  *bb.Client

	inputJSON              string
	pollingIntervalSeconds int
	bbids                  list
}

type list []string

func (l *list) Set(value string) error {
	*l = strings.Split(strings.TrimSpace(value), ",")
	return nil
}

func (l *list) String() string {
	return strings.Join(*l, ",")
}

func cmdCollect() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "collect",
		ShortDesc: "Collect on the specified builds, retrying as configured.",
		CommandRun: func() subcommands.CommandRun {
			c := &collectRun{}
			c.cmdRunner = cmd.RealCommandRunner{}
			c.Flags.StringVar(&c.inputJSON, "input_json", "", "Path to JSON proto representing a CollectConfig")
			c.Flags.IntVar(&c.pollingIntervalSeconds, "polling_interval", 60, "Seconds to wait between polling builders")
			c.Flags.Var(&c.bbids, "bbids", "(comma-separated) initial set of BBIDs to watch.")
			return c
		}}
}

// validate validates release-specific args for the command.
func (c *collectRun) validate() error {
	if c.inputJSON == "" {
		return fmt.Errorf("--input_json is required")
	}

	if len(c.bbids) == 0 {
		return fmt.Errorf("Must specify at least one BBID.")
	}

	return nil
}

func (c *collectRun) readInput() (*pb.CollectConfig, error) {
	inputBytes, err := ioutil.ReadFile(c.inputJSON)
	if err != nil {
		return nil, fmt.Errorf("Failed reading input_json\n%v", err)
	}
	req := &pb.CollectConfig{}
	if err := unmarshaler.Unmarshal(bytes.NewReader(inputBytes), req); err != nil {
		return nil, fmt.Errorf("Couldn't decode %s as a CollectConfig\n%v", c.inputJSON, err)
	}
	return req, nil
}

func (c *collectRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	c.stdoutLog = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	c.stderrLog = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	c.bbClient = bb.NewClient(c.cmdRunner, c.stdoutLog, c.stderrLog)

	ctx := context.Background()
	if err := c.bbClient.EnsureLUCIToolsAuthed(ctx, "bb"); err != nil {
		c.LogErr(err.Error())
		// TODO(b/264680777): Factor return_codes.go out of try and use those.
		return 1
	}

	if err := c.validate(); err != nil {
		c.LogErr(err.Error())
		return 2
	}

	collectConfig, err := c.readInput()
	if err != nil {
		c.LogErr(err.Error())
		return 3
	}

	if err := c.Collect(ctx, collectConfig); err != nil {
		c.LogErr(err.Error())
		return 4
	}

	return 0
}

// LogOut logs to stdout.
func (t *collectRun) LogOut(format string, a ...interface{}) {
	if t.stdoutLog != nil {
		t.stdoutLog.Printf(format, a...)
	}
}

// LogErr logs to stderr.
func (t *collectRun) LogErr(format string, a ...interface{}) {
	if t.stderrLog != nil {
		t.stderrLog.Printf(format, a...)
	}
}

// GetApplication returns an instance of the application.
func GetApplication() *cli.Application {
	return &cli.Application{
		Name: "conductor",

		Context: func(ctx context.Context) context.Context {
			return ctx
		},

		Commands: []*subcommands.Command{
			subcommands.CmdHelp,
			cmdCollect(),
		},
	}
}

func main() {
	app := GetApplication()
	os.Exit(subcommands.Run(app, nil))
}
