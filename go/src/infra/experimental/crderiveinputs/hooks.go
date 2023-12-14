// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"path"
	"strings"

	"go.chromium.org/luci/common/data/stringset"
)

type HookImpl interface {
	HandleHook(oracle *Oracle, cwd string, action *GclientHook) (handled bool, err error)
}

var knownPythonNames = stringset.NewFromSlice(
	"python",
	"python2",
	"python3",
	"vpython",
	"vpython2",
	"vpython3",
)

type DisableDepotToolsSelfupdate struct{}

func (DisableDepotToolsSelfupdate) forPath(oracle *Oracle, depot_tools_path string) {
	oracle.PinRawFile(path.Join(depot_tools_path, ".disable_auto_update"), "Disabled by crderiveinputs.", "crderiveinputs.DisableDepotToolsSelfupdate hook")
}

func (d DisableDepotToolsSelfupdate) HandleHook(oracle *Oracle, cwd string, action *GclientHook) (handled bool, err error) {
	if strings.HasSuffix(action.Action[len(action.Action)-2], "update_depot_tools_toggle.py") && action.Action[len(action.Action)-1] == "--disable" {
		LEAKY("update_depot_tools_toggle.py .disable_auto_update")

		depot_tools_path := path.Join(cwd, path.Dir(action.Action[len(action.Action)-2]))
		d.forPath(oracle, depot_tools_path)
		return true, nil
	}
	return false, nil
}

type ignoreHookNamed string

func (i ignoreHookNamed) HandleHook(oracle *Oracle, cwd string, action *GclientHook) (handled bool, err error) {
	if action.Name == string(i) {
		LEAKY("ignoring %q hook", i)
		handled = true
	}
	return
}
