// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/maruel/subcommands"

	"infra/libs/skylab/inventory"
)

// Summarize implements the migrate subcommand.
var Summarize = &subcommands.Command{
	UsageLine: "summarize -root DATA_DIR HOSTNAME [HOSTNAME...]",
	ShortDesc: "summarize DUT migration status",
	LongDesc:  "Summarize DUT migration status.",
	CommandRun: func() subcommands.CommandRun {
		c := &summarizeRun{}
		c.Flags.StringVar(&c.rootDir, "root", "", "root `directory` of the inventory checkout")
		return c
	},
}

type summarizeRun struct {
	subcommands.CommandRunBase
	rootDir string
}

func (c *summarizeRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	return 0
}

func (c *summarizeRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if c.rootDir == "" {
		return errors.New("-root is required")
	}

	labs, err := loadAllLabsData(c.rootDir)
	if err != nil {
		return err
	}

	sCounts := summarize(labs.Skylab.Duts)
	aCounts := summarize(labs.AutotestProd.Duts)

	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "Number of models with non-zero DUTs per pool:\n")
	fmt.Fprintf(tw, "Pool\tAutotest(prod)\tSkylab\n")
	for k := range mergedKeySets(sCounts, aCounts) {
		fmt.Fprintf(tw, "%s\t%d\t%d\n", k, len(aCounts[k]), len(sCounts[k]))
	}
	return nil
}

type poolCounts map[string]modelCounts

type modelCounts map[string]int

func summarize(duts []*inventory.DeviceUnderTest) poolCounts {
	pCounts := make(poolCounts)
	for _, d := range duts {
		ls := d.GetCommon().GetLabels()
		cp := ls.GetCriticalPools()
		var p string
		if len(cp) > 0 {
			p = cp[0].String()
		} else {
			p = "[NON_CRITICAL_POOL]"
		}
		mc, ok := pCounts[p]
		if !ok {
			mc = make(modelCounts)
			pCounts[p] = mc
		}
		mc[ls.GetModel()]++
	}
	return pCounts
}

func mergedKeySets(a, b poolCounts) map[string]bool {
	m := make(map[string]bool)
	for k := range a {
		m[k] = true
	}
	for k := range b {
		m[k] = true
	}
	return m
}
