// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package whoami provides whoami subcommand.
package whoami

import (
	"fmt"
	"os"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"

	"infra/build/siso/auth/cred"
	"infra/build/siso/reapi"
)

func Cmd(authOpts cred.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "whoami",
		ShortDesc: "prints current auth status.",
		LongDesc:  "Prints current auth status.",
		CommandRun: func() subcommands.CommandRun {
			r := &whoamiRun{authOpts: authOpts}
			r.init()
			return r
		},
	}
}

type whoamiRun struct {
	subcommands.CommandRunBase
	authOpts  cred.Options
	projectID string
	reopt     *reapi.Option
}

func (r *whoamiRun) init() {
	r.Flags.StringVar(&r.projectID, "project", os.Getenv("SISO_PROJECT"), "cloud project ID. can set by $SISO_PROJECT")

	r.reopt = new(reapi.Option)
	envs := map[string]string{
		"SISO_REAPI_INSTANCE": os.Getenv("SISO_REAPI_INSTANCE"),
		"SISO_REAPI_ADDRESS":  os.Getenv("SISO_REAPI_ADDRESS"),
	}
	r.reopt.RegisterFlags(&r.Flags, envs)
}

func (r *whoamiRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)
	if len(args) != 0 {
		fmt.Fprintf(a.GetErr(), "%s: position arguments not expected\n", a.GetName())
		return 1
	}
	credential, err := cred.New(ctx, r.authOpts)
	if err != nil {
		fmt.Printf("auth error: %v\n", err)
		return 1
	}
	fmt.Printf("Logged in by %s\n", credential.Type)
	if credential.Email != "" {
		fmt.Printf(" as %s\n", credential.Email)
	}
	r.reopt.UpdateProjectID(r.projectID)
	if r.reopt.IsValid() {
		client, err := reapi.New(ctx, credential, *r.reopt)
		fmt.Printf("use reapi instance %s\n", r.reopt.Instance)
		if err != nil {
			fmt.Printf("access error: %v\n", err)
			return 1
		}
		defer client.Close()
	}
	return 0
}
