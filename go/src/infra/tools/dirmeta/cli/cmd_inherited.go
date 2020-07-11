// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"context"
	"infra/tools/dirmeta"
	"path/filepath"
	"strings"

	"github.com/maruel/subcommands"
	"google.golang.org/protobuf/encoding/protojson"

	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"
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
	root string

	// target is the directory for which we need to read the inherited metadata.
	// It is relative to root and uses forward slashes as path separators.
	target string
	output string
}

func (r *inheritedRun) parseInput(args []string) error {
	if len(args) != 2 {
		return errors.Reason("expected exactly two positional arguments, got %q", args).Err()
	}
	r.root = filepath.Clean(args[0])

	r.target = filepath.Clean(args[1])
	var err error
	switch r.target, err = filepath.Rel(r.root, r.target); {
	case err != nil:
		return errors.Annotate(err, "invalid combination of ROOT and TARGET").Err()
	case strings.HasPrefix(r.target, ".."):
		return errors.Reason("TARGET must same as ROOT or it must be a sub-directory of ROOT").Err()
	}

	return nil
}
func (r *inheritedRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)
	if err := r.parseInput(args); err != nil {
		return r.done(ctx, err)
	}

	return r.done(ctx, r.run(ctx))
}

func (r *inheritedRun) run(ctx context.Context) error {
	mapping, err := dirmeta.ReadMapping(r.root)
	if err != nil {
		return err
	}

	data, err := protojson.Marshal(mapping.Inherited(r.target))
	if err != nil {
		return err
	}
	return r.writeTextOutput(data)
}
