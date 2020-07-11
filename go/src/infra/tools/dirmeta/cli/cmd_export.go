// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/tools/dirmeta"
)

func cmdExport() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `export`,
		ShortDesc: "export metadata from a directory tree",
		LongDesc:  "Export metadata from a directory tree.",
		CommandRun: func() subcommands.CommandRun {
			r := &exportRun{}
			r.Flags.StringVar(&r.root, "root", ".", "Path to the root directory")
			r.RegisterOutputFlag()
			r.Flags.BoolVar(&r.expand, "expand", false, `Expand the mapping, i.e. inherit values`)
			r.Flags.BoolVar(&r.reduce, "reduce", false, `Reduce the mapping, i.e. remove all redundant information`)
			return r
		},
	}
}

type exportRun struct {
	baseCommandRun
	root           string
	expand, reduce bool
}

func (r *exportRun) parseInput(args []string) error {
	if len(args) != 0 {
		return errors.Reason("unexpected positional arguments: %q", args).Err()
	}

	if r.expand && r.reduce {
		return errors.Reason("-expand and -reduce are mutually excusive").Err()
	}

	return nil
}

func (r *exportRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	if err := r.parseInput(args); err != nil {
		return r.done(ctx, err)
	}

	return r.done(ctx, r.run(ctx))
}

func (r *exportRun) run(ctx context.Context) error {
	mapping, err := dirmeta.ReadMapping(r.root)
	if err != nil {
		return err
	}

	if r.expand {
		mapping = mapping.Expand()
	} else if r.reduce {
		mapping = mapping.Reduce()
	}

	data, err := protojson.Marshal(mapping.Proto())
	if err != nil {
		return err
	}

	return r.writeTextOutput(data)
}
