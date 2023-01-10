// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package buildbucket

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// runCmd executes a shell command.
func (c *Client) runCmd(ctx context.Context, name string, args ...string) (stdout, stderr string, err error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	err = c.cmdRunner.RunCommand(ctx, &stdoutBuf, &stderrBuf, "", name, args...)
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		return stdout, stderr, errors.Annotate(err, fmt.Sprintf("running `%s %s`", name, strings.Join(args, " "))).Err()
	}
	return stdout, stderr, nil
}

// SeparateBucketFromBuilder takes a full builder name (like chromeos/release/release-main-orchestrator),
// and separates it into a bucket (chromeos/release) and a builder (release-main-orchestrator).
func SeparateBucketFromBuilder(fullBuilderName string) (bucket string, builder string, err error) {
	parts := strings.Split(fullBuilderName, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("builder %s has %d slash-delimited parts; expect 3", fullBuilderName, len(parts))
	}
	bucket = strings.Join(parts[:2], "/")
	builder = parts[2]
	return bucket, builder, nil
}
