// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"

	"infra/tools/dirmeta"
)

func cmdExport() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `export`,
		ShortDesc: "export metadata from a directory tree",
		LongDesc: text.Doc(`
			Export metadata from a directory tree to stdout or to a file.

			The output format is JSON form of chrome.dir_metadata.Mapping protobuf
			message.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &exportRun{}
			r.RegisterOutputFlag()
			r.Flags.StringVar(&r.Root, "root", ".", "Path to the root directory")
			r.Flags.BoolVar(&r.expand, "expand", false, `Expand the mapping, i.e. inherit values`)
			r.Flags.BoolVar(&r.reduce, "reduce", false, `Reduce the mapping, i.e. remove all redundant information`)
			return r
		},
	}
}

type exportRun struct {
	baseCommandRun
	dirmeta.MappingReader
	expand, reduce bool
}

func (r *exportRun) parseInput(args []string) error {
	if len(args) != 0 {
		return errors.Reason("unexpected positional arguments: %q", args).Err()
	}

	if r.expand && r.reduce {
		return errors.Reason("-expand and -reduce are mutually exclusive").Err()
	}

	return nil
}

func (r *exportRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	if err := r.parseInput(args); err != nil {
		return r.done(ctx, err)
	}

	if err := r.ReadFull(); err != nil {
		return r.done(ctx, err)
	}

	ret := &r.Mapping
	if r.expand {
		ret = ret.Expand()
	} else if r.reduce {
		ret = ret.Reduce()
	}

	return r.done(ctx, r.writeMapping(ret))
}
