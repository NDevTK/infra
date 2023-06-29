// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"context"
	"infra/libs/cipkg_new/core"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"go.chromium.org/luci/common/system/environ"
)

// ActionCIPDExportTransformer is the default transformer for cipd action.
func ActionCIPDExportTransformer(a *core.ActionCIPDExport, deps []PackageDependency) (*core.Derivation, error) {
	return ReexecDerivation(a, deps, true)
}

// ActionCIPDExportExecutor is the default executor for cipd action.
func ActionCIPDExportExecutor(ctx context.Context, a *core.ActionCIPDExport, dstFS afero.Fs) error {
	env := environ.FromCtx(ctx)
	env.Update(environ.New(a.Env))
	cmd := CIPDCommand(ctx, "export", "--root", env.Get("out"), "--ensure-file", "-")
	cmd.Env = env.Sorted()
	cmd.Stdin = strings.NewReader(a.EnsureFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Create a exec.Cmd for cipd which lookup and expands 'cipd' to it's path.
// exec.Command already did that and store the path in Cmd.Path, but doesn't
// work properly for .bat script.
func CIPDCommand(ctx context.Context, arg ...string) *exec.Cmd {
	cipd := lookup("cipd")

	// Use cmd to execute batch file on windows.
	if filepath.Ext(cipd) == ".bat" {
		return exec.CommandContext(ctx, lookup("cmd.exe"), append([]string{"/C", cipd}, arg...)...)
	}

	return exec.CommandContext(ctx, cipd, arg...)
}

func lookup(bin string) string {
	if path, err := exec.LookPath(bin); err == nil {
		return path
	}
	return bin
}
