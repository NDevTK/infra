// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
)

// runBBCmd runs a `bb` subcommand.
func (m tryRunBase) runBBCmd(ctx context.Context, subcommand string, args ...string) (stdout, stderr string, err error) {
	return m.RunCmd(ctx, "bb", prependString("add", args)...)
}

// BBAdd runs a `bb add` command, and prints stdout to the user.
func (m tryRunBase) BBAdd(ctx context.Context, args ...string) error {
	stdout, stderr, err := m.runBBCmd(ctx, "add", args...)
	if err != nil {
		fmt.Println(stderr)
		return err
	}
	fmt.Println(stdout)
	return nil
}
