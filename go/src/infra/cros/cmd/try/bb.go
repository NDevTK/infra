// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"strings"
)

// runBBCmd runs a `bb` subcommand.
func (t *tryRunBase) runBBCmd(ctx context.Context, subcommand string, args ...string) (stdout, stderr string, err error) {
	return t.RunCmd(ctx, "bb", prependString(subcommand, args)...)
}

// BBAdd runs a `bb add` command, and prints stdout to the user.
func (t *tryRunBase) BBAdd(ctx context.Context, args ...string) error {
	stdout, stderr, err := t.runBBCmd(ctx, "add", args...)
	if err != nil {
		fmt.Println(stderr)
		return err
	}
	fmt.Println(stdout)
	return nil
}

// getBuilders runs the `bb builders` command to get all builders in the given bucket.
// The bucket param should not include the project prefix (normally "chromeos/").
func (t *tryRunBase) BBBuilders(ctx context.Context, bucket string) ([]string, error) {
	stdout, stderr, err := t.runBBCmd(ctx, "builders", fmt.Sprintf("chromeos/%s", bucket))
	if err != nil {
		fmt.Println(stderr)
		return []string{}, err
	}
	return strings.Split(stdout, "\n"), nil
}
