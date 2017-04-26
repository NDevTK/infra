// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/context"

	"github.com/maruel/subcommands"

	"github.com/luci/luci-go/client/authcli"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/flag/stringmapflag"
	"github.com/luci/luci-go/common/isolatedclient"
	"github.com/luci/luci-go/common/logging"
)

const defaultIsolateServer = "https://isolateserver.appspot.com"

func isolateCmd(authOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "isolate [-O project_id=/path/to/local/repo]*",
		ShortDesc: "Isolates a bundle of recipes from the current working directory.",
		LongDesc: `Takes recipes from the current repo (based on cwd), along with
		any supplied overrides, and pushes them to the isolate service.`,

		CommandRun: func() subcommands.CommandRun {
			ret := &cmdIsolate{}
			ret.logCfg.Level = logging.Info

			ret.logCfg.AddFlags(&ret.Flags)
			ret.authFlags.Register(&ret.Flags, authOpts)
			ret.isolatedFlags.Init(&ret.Flags)

			ret.Flags.Var(&ret.overrides, "O",
				"override a repo dependency. Must be in the form of project_id=/path/to/local/repo. May be specified multiple times.")

			ret.Flags.BoolVar(&ret.editMode, "em", false, "alias for `edit-mode`")
			ret.Flags.BoolVar(&ret.editMode, "edit-mode", false,
				("enables `edit mode`. This causes the command to read a JobDescription on stdin, add the isolate hash and print the " +
					"new JobDescription to stdout."))
			return ret
		},
	}
}

type cmdIsolate struct {
	subcommands.CommandRunBase

	logCfg        logging.Config
	authFlags     authcli.Flags
	isolatedFlags isolatedclient.Flags

	overrides stringmapflag.Value

	editMode bool
}

func (c *cmdIsolate) validateFlags(ctx context.Context, args []string) (authOpts auth.Options, err error) {
	if len(args) > 0 {
		err = errors.Reason("unexpected positional arguments: %(args)q").D("args", args).Err()
		return
	}

	for k, v := range c.overrides {
		if k == "" {
			err = errors.New("override has empty project_id")
			return
		}
		if v == "" {
			err = errors.Reason("override %(key)q has empty repo path").D("key", k).Err()
			return
		}
		v, err = filepath.Abs(v)
		if err != nil {
			err = errors.Annotate(err).Reason("override %(key)q").D("key", k).Err()
			return
		}
		c.overrides[k] = v

		var fi os.FileInfo
		switch fi, err = os.Stat(v); {
		case err != nil:
			err = errors.Annotate(err).Reason("override %(key)q").D("key", k).Err()
			return
		case !fi.IsDir():
			err = errors.Reason("override %(key)q: not a directory").D("key", k).Err()
			return
		}
	}

	if c.isolatedFlags.ServerURL == "" {
		c.isolatedFlags.ServerURL = defaultIsolateServer
	}
	if err = c.isolatedFlags.Parse(); err != nil {
		err = errors.Annotate(err).Reason("bad isolate flags").Err()
		return
	}

	return c.authFlags.Options()
}

func (c *cmdIsolate) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := c.logCfg.Set(cli.GetContext(a, c, env))
	authOpts, err := c.validateFlags(ctx, args)
	if err != nil {
		logging.Errorf(ctx, "bad arguments: %s", err)
		fmt.Fprintln(os.Stderr)
		subcommands.CmdHelp.CommandRun().Run(a, []string{"isolate"}, env)
		return 1
	}

	logging.Infof(ctx, "bundling recipes")
	bundlePath, err := bundle(ctx, c.overrides)
	if err != nil {
		logging.Errorf(ctx, "fatal error during bundle: %s", err)
		return 1
	}
	defer os.RemoveAll(bundlePath)
	logging.Infof(ctx, "bundling recipes: done")

	logging.Infof(ctx, "isolating recipes")
	hash, err := isolate(ctx, bundlePath, c.isolatedFlags, authOpts)
	if err != nil {
		logging.Errorf(ctx, "fatal error: %s", err)
		return 1
	}
	logging.Infof(ctx, "isolating recipes: done")

	if c.editMode {
		err := editMode(func(jd *JobDefinition) (*JobDefinition, error) {
			ret := &(*jd)
			ret.RecipeIsolatedHash = hash
			return ret, nil
		})
		if err != nil {
			logging.WithError(err).Errorf(ctx, "fatal")
			return 1
		}
	} else {
		logging.Infof(ctx, "isolated: %q", hash)
		logging.Infof(ctx, "URL: %s/browse?namespace=%s&hash=%s",
			c.isolatedFlags.ServerURL, c.isolatedFlags.Namespace, hash)
	}

	return 0
}
