// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"context"

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
			r.Flags.BoolVar(&r.expand, "expand", false, `Expand the mapping, i.e. inherit values in all directories`)
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

func (r *exportRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)
	return r.done(ctx, r.run(ctx, args))
}

func (r *exportRun) run(ctx context.Context, args []string) error {
	switch {
	case len(args) != 0:
		return errors.Reason("unexpected positional arguments: %q", args).Err()
	case r.expand && r.reduce:
		return errors.Reason("-expand and -reduce are mutually exclusive").Err()
	}

	if err := r.ReadAll(r.expand); err != nil {
		return err
	}

	ret := &r.Mapping
	if r.reduce {
		ret = ret.Reduce()
	}

	return r.writeMapping(ret)
}
