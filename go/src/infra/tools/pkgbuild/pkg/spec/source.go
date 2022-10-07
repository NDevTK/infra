// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
)

// SourceResolver can resolve the latest valid version from source definition.
// It isolates the side effects with derivations, which makes the builds
// deterministic.
// Resolver may include executing non-hermetic host binaries and scripts.
type SourceResolver interface {
	ResolveGitSource(git *GitSource) (tag, commit string, err error)
	ResolveScriptSource(script *ScriptSource) (version string, err error)
}

//go:embed resolve_git.py
var resolveGitScript string

type tagInfo struct {
	// Regulated semantic versioning tag e.g. 1.2.3
	// This may not be the corresponding git tag.
	Tag string

	// Git commit for the tag.
	Commit string
}

type DefaultSourceResolver struct{}

func (*DefaultSourceResolver) ResolveGitSource(git *GitSource) (tag, commit string, err error) {
	cmd := exec.Command("python3", "-I", "-c", resolveGitScript)
	cmd.Env = []string{}
	cmd.Stderr = os.Stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return "", "", err
	}
	out, err := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", "", err
	}

	if err := json.NewEncoder(in).Encode(git); err != nil {
		return "", "", err
	}
	in.Close()

	var info tagInfo
	if err := json.NewDecoder(out).Decode(&info); err != nil {
		return "", "", err
	}
	out.Close()

	if err := cmd.Wait(); err != nil {
		return "", "", err
	}

	return info.Tag, info.Commit, nil
}

func (*DefaultSourceResolver) ResolveScriptSource(script *ScriptSource) (version string, err error) {
	panic("not implemented")
}
