// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const BuiltinBuilderPrefix = "builtin:"

// Execute the command with support for builtin builders.
// Although Execute() includes context as argument, command should be created
// using exec.CommandContext. The context argument of Execute() is only used
// for builtin builders because we can't retrieve context from exec.Cmd.
func Execute(ctx context.Context, cmd *exec.Cmd) error {
	if !strings.HasPrefix(cmd.Path, BuiltinBuilderPrefix) {
		return cmd.Run()
	}

	if strings.HasPrefix(cmd.Path, UDFBuilderPrefix) {
		return executeUserdefinedFunction(ctx, cmd)
	}

	switch cmd.Path {
	case FetchURLsBuilder:
		return fetchURLs(ctx, cmd)
	case CopyFilesBuilder:
		return copyFiles(ctx, cmd)
	case ImportBuilder:
		return importFromHost(ctx, cmd)
	case CIPDEnsureBuilder:
		return cipdEnsure(ctx, cmd)
	}
	return fmt.Errorf("unknown builtin builder: %s", cmd.Path)
}

func GetEnv(k string, envs []string) string {
	for _, env := range envs {
		if ss := strings.SplitN(env, "=", 2); ss[0] == k {
			return ss[1]
		}
	}
	return ""
}
