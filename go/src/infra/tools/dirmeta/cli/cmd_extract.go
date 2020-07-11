// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cli

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/data/text"
	"go.chromium.org/luci/common/errors"

	"infra/tools/dirmeta"
)

func cmdExtract() *subcommands.Command {
	return &subcommands.Command{
		UsageLine: `extract`,
		ShortDesc: "extract metadata from a directory tree",
		LongDesc:  "Extract metadata from a directory tree.",
		CommandRun: func() subcommands.CommandRun {
			r := &extractRun{}
			r.RegisterOutputFlag()
			r.Flags.StringVar(&r.root, "root", ".", "Path to the root directory")
			r.Flags.StringVar(&r.formatFlag, "format", "proto-json", text.Doc(`
				Format of the output. Valid values:
				"proto-json" (JSON form of the chrome.dir_meta.Mapping protobuf message),
				"chrome-legacy" (format used in https://storage.googleapis.com/chromium-owners/component_map_subdirs.json)
			`))
			r.Flags.BoolVar(&r.expand, "expand", false, `Expand the mapping, i.e. inherit values`)
			r.Flags.BoolVar(&r.reduce, "reduce", false, `Reduce the mapping, i.e. remove all redundant information`)
			return r
		},
	}
}

type extractRun struct {
	baseCommandRun
	root           string
	formatFlag     string
	format         outputFormat
	expand, reduce bool
}

func (r *extractRun) parseInput(args []string) error {
	if len(args) != 0 {
		return errors.Reason("unexpected positional arguments: %q", args).Err()
	}

	var err error
	if r.format, err = parseOutputFormat(r.formatFlag); err != nil {
		return errors.Annotate(err, "invalid -format").Err()
	}

	if r.expand && r.reduce {
		return errors.Reason("-expand and -reduce are mutually excusive").Err()
	}

	return nil
}

func (r *extractRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, r, env)

	if err := r.parseInput(args); err != nil {
		return r.done(ctx, err)
	}

	return r.done(ctx, r.run(ctx))
}

func (r *extractRun) run(ctx context.Context) error {
	mapping, err := dirmeta.ReadMapping(r.root)
	if err != nil {
		return err
	}

	if r.expand {
		mapping = mapping.Expand()
	} else if r.reduce {
		mapping = mapping.Reduce()
	}

	data, err := r.marshal(mapping)
	if err != nil {
		return err
	}

	return r.writeTextOutput(data)
}

// marshal marshals m using the format specified in r.format.
func (r *extractRun) marshal(m *dirmeta.Mapping) ([]byte, error) {
	switch r.format {
	case protoJSON:
		return protojson.Marshal(m.Proto())

	case chromeLegacy:
		panic("not implemented")

	default:
		panic("impossible")
	}
}
