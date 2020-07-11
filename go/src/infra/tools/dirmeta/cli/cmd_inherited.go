// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"

	"infra/tools/dirmeta"
)

func cmdInherited() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `inherited ROOT TARGET`,
		ShortDesc: "Read metadata inherited by the target dir",
		LongDesc: text.Doc(`
			Read metadata inherited by the target dir.

			ROOT is the path to the root directory with metadata files.
			Print the metadata inherited by TARGET.
		`),
		CommandRun: func() subcommands.CommandRun {
			r := &inheritedRun{}
			r.RegisterOutputFlag()
			return r
		},
	}
}

type inheritedRun struct {
	baseCommandRun
}

func (r *inheritedRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	if len(args) != 2 {
		return r.done(ctx, errors.Reason("expected exactly two positional arguments, got %q", args).Err())
	}
	md, err := dirmeta.ReadInherited(args[0], args[1])
	if err != nil {
		return r.done(ctx, err)
	}

	data, err := protojson.Marshal(md)
	if err != nil {
		return r.done(ctx, err)
	}
	return r.done(ctx, r.writeTextOutput(data))
}
