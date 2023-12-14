// Copyright (c) 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

func ResolveDepotTools(oracle *Oracle) error {
	if err := oracle.PinGit("depot_tools", "https://chromium.googlesource.com/chromium/tools/depot_tools", "refs/heads/main"); err != nil {
		return err
	}

	LEAKY("assuming cipd_manifest.txt in depot_tools")
	SBOM("assuming all packages in depot_tools/cipd_manifest.txt are needed for build")
	LEAKY("assuming depot_tools/.cipd_bin subdirectory for CIPD packages")

	err := oracle.PinCipdEnsureFile(
		"depot_tools/.cipd_bin",
		"depot_tools/cipd_manifest.txt",
	)
	if err != nil {
		return err
	}

	DisableDepotToolsSelfupdate{}.forPath(oracle, "depot_tools")

	return nil
}
