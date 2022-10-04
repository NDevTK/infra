// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"strings"
)

//go:embed resolve_git.py
var resolveGitScript string

type tagInfo struct {
	// Regulated semantic versioning tag e.g. 1.2.3
	// This may not be the corresponding git tag.
	Tag string

	// Git commit for the tag.
	Commit string
}

// resolveGitTag require python3 and git in the PATH.
func resolveGitRef(git *GitSource) (tagInfo, error) {
	cmd := exec.Command("python3", "-c", resolveGitScript)
	cmd.Env = []string{}
	cmd.Stderr = os.Stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return tagInfo{}, err
	}
	out, err := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return tagInfo{}, err
	}

	if err := json.NewEncoder(in).Encode(git); err != nil {
		return tagInfo{}, err
	}
	in.Close()

	var info tagInfo
	if err := json.NewDecoder(out).Decode(&info); err != nil {
		return tagInfo{}, err
	}
	out.Close()

	if err := cmd.Wait(); err != nil {
		return tagInfo{}, err
	}

	return info, nil
}

func gitCachePath(url string) string {
	url = strings.TrimPrefix(url, "https://chromium.googlesource.com/external/")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	return path.Clean(url)
}
