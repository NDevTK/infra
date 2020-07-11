// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/tools/dirmeta"
	dirmetapb "infra/tools/dirmeta/proto"
)

func cmdCompute() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `compute -root ROOT TARGET1 [TARGET2...]`,
		ShortDesc: "compute metadata for the given target directories",
		LongDesc:  "Compute metadata for the given target directories.",
		CommandRun: func() subcommands.CommandRun {
			r := &computeRun{}
			r.RegisterOutputFlag()
			r.Flags.StringVar(&r.Root, "root", ".", "Path to the root directory")
			return r
		},
	}
}

type computeRun struct {
	baseCommandRun
	dirmeta.MappingReader
}

func (r *computeRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	for _, dir := range args {
		if err := r.ReadTowards(dir); err != nil {
			return r.done(ctx, errors.Annotate(err, "failed to read metadata for %q", dir).Err())
		}
	}

	// Print metadata for only target dirs.
	ret := &dirmeta.Mapping{Dirs: map[string]*dirmetapb.Metadata{}}
	for _, dir := range args {
		key, err := r.DirKey(dir)
		if err != nil {
			panic(err) // Impossible: we have just used thes paths above.
		}
		ret.Dirs[key] = r.Compute(key)
	}
	return r.done(ctx, r.writeMapping(ret))
}
