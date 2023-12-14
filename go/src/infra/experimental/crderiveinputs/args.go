// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/common/flag/stringmapflag"
)

type Args struct {
	CrCommit string
	GClientVars
	CacheDirectory string
	FailFast       bool

	cipdExpander template.Expander
}

var hostOSToCipdOS = map[string]string{
	"cygwin": "windows",
	"win":    "windows",
	"win32":  "windows",

	"darwin": "mac",

	"linux2": "linux",
	"linux":  "linux",

	"aix6": "aix",
}

var hostCPUToCipdArch = map[string]string{
	"x86":     "386",
	"x64":     "amd64",
	"arm64":   "arm64",
	"arm":     "armv6l",
	"mips64":  "mips64",
	"mips":    "mips",
	"ppc":     "ppc",
	"s390":    "s390",
	"riscv64": "riscv64",
}

func (a *Args) CIPDOS() string {
	hostOS := a.GClientVars.HostOS
	ret, ok := hostOSToCipdOS[hostOS]
	if !ok {
		panic(fmt.Sprintf("Args.CIPDOS() - Do not know how to map hostOS(%q) value to cipd 'os'.", hostOS))
	}
	return ret
}

func (a *Args) CIPDArch() string {
	hostCPU := a.GClientVars.HostCPU
	ret, ok := hostCPUToCipdArch[hostCPU]
	if !ok {
		panic(fmt.Sprintf("Args.CIPDArch() - Do not know how to map hostCPU(%q) value to cipd 'arch'.", hostCPU))
	}
	return ret
}

func parseArgs(args []string, cwd string) (*Args, error) {
	ret := &Args{
		GClientVars: GClientVars{
			StrVars:  stringmapflag.Value{},
			BoolVars: boolmapflag{},
		},
	}

	fs := flag.NewFlagSet("crderiveinputs", flag.ContinueOnError)
	fs.StringVar(
		&ret.CrCommit, "cr-commit", "refs/heads/main",
		"The commit/branch/tag of Chromium to generate an input manifest for. If a commit is given, must be a full commit hash."+
			" NOTE: 'HEAD' is hard-coded to refs/heads/main (otherwise ls-remote takes ~45s).")

	fs.StringVar(
		&ret.CacheDirectory, "cache", filepath.Join(cwd, "crderiveinputs.cache"),
		"Path to a directory to use to cache git objects and other temporary things (like gclient). "+
			"Will be created if it doesn't exist.")

	fs.BoolVar(
		&ret.FailFast, "fail-fast", false, "Exit with non-zero after first TODO.",
	)

	ret.GClientVars.AddToFlagset(fs)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	logFailFast = ret.FailFast

	if ret.CrCommit == "HEAD" {
		ret.CrCommit = "refs/heads/main"
	}

	return ret, nil
}
