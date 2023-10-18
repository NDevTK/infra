// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SourceResolver can resolve the latest valid version from source definition.
// It isolates the side effects with derivations, which makes the builds
// deterministic.
// Resolver may include executing non-hermetic host binaries and scripts.
type SourceResolver interface {
	ResolveGitSource(git *GitSource) (info GitSourceInfo, err error)
	ResolveScriptSource(hostCipdPlatform, dir string, script *ScriptSource) (info ScriptSourceInfo, err error)
}

type GitSourceInfo struct {
	// Regulated semantic versioning tag e.g. 1.2.3
	// This may not be the corresponding git tag.
	Tag string

	// Git commit for the tag.
	Commit string
}

type ScriptSourceInfo struct {
	Version string

	URL  []string
	Name []string
	Ext  string
}

//go:embed resolve_git.py
var resolveGitScript string

type DefaultSourceResolver struct {
	VPythonSpecPath string
}

func (r *DefaultSourceResolver) ResolveGitSource(git *GitSource) (info GitSourceInfo, err error) {
	cmd := r.command("resolve_git.py", resolveGitScript)

	in, err := json.Marshal(git)
	if err != nil {
		return
	}
	cmd.Args = append(cmd.Args, string(in))

	out, err := output(cmd)
	if err != nil {
		return
	}
	if err = json.Unmarshal([]byte(out), &info); err != nil {
		return
	}

	return
}

func (r *DefaultSourceResolver) ResolveScriptSource(hostCipdPlatform, dir string, script *ScriptSource) (info ScriptSourceInfo, err error) {
	scriptName := script.GetName()[0]
	f, err := os.Open(filepath.Join(dir, scriptName))
	if err != nil {
		return
	}
	defer f.Close()
	s, err := io.ReadAll(f)
	sourceScript := string(s)

	// Get version
	cmd := r.command(scriptName, sourceScript)
	cmd.Args = append(cmd.Args, script.GetName()[1:]...)
	cmd.Args = append(cmd.Args, "latest")
	cmd.Env = append(cmd.Env, fmt.Sprintf("_3PP_PLATFORM=%s", hostCipdPlatform))

	out, err := output(cmd)
	if err != nil {
		return
	}
	version := strings.TrimSpace(string(out))

	if script.GetUseFetchCheckoutWorkflow() {
		return
	}

	if script.GetUseFetchCheckoutWorkflow() {
		// TODO(fancl): running checkout inside derivation
		panic("not implemented")
	}

	// Get download urls
	cmd = r.command(scriptName, sourceScript)
	cmd.Args = append(cmd.Args, script.GetName()[1:]...)
	cmd.Args = append(cmd.Args, "get_url")
	cmd.Env = append(cmd.Env, fmt.Sprintf("_3PP_VERSION=%s", version))
	cmd.Env = append(cmd.Env, fmt.Sprintf("_3PP_PLATFORM=%s", hostCipdPlatform))

	out, err = output(cmd)
	if err != nil {
		return
	}
	if err = json.Unmarshal(out, &info); err != nil {
		return
	}
	info.Version = version

	return
}

func output(cmd *exec.Cmd) (out []byte, err error) {
	out, err = cmd.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, string(e.Stderr))
		}
	}
	return
}

func (r *DefaultSourceResolver) command(name, script string) *exec.Cmd {
	var cmd *exec.Cmd
	switch filepath.Ext(name) {
	case ".py":
		cmd = exec.Command("vpython3", "-vpython-spec", r.VPythonSpecPath, "-")
	case ".sh":
		cmd = exec.Command("bash", "-s", "-")
	default:
		panic("unknown script: " + name)
	}
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(script)
	return cmd
}
