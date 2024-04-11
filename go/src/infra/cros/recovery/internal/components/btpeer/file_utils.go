// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package btpeer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
)

// AddLineToFile appends a single line option to a file if it does not already exist.
func AddLineToFile(ctx context.Context, runner components.Runner, filePath string, option string) error {
	// Check if line already exists.
	existsCmd := fmt.Sprintf("cat %q | grep -Fx %q", filePath, option)
	if out, err := runner(ctx, 30*time.Second, existsCmd); err == nil && strings.TrimSpace(out) != "" {
		// Nothing to do, already exists.
		return nil
	}

	optionCmd := fmt.Sprintf("'echo %q >> %q'", option, filePath)
	if _, err := runner(ctx, 30*time.Second, "bash", "-c", optionCmd); err != nil {
		return errors.Annotate(err, "add config option: failed to write %q to file %q", option, filePath).Err()
	}
	return nil
}

// RemoveLineFromFile removes any matching lines from a config file.
func RemoveLineFromFile(ctx context.Context, runner components.Runner, filePath string, option string) error {
	// Check if line already exists.
	existsCmd := fmt.Sprintf("cat %q | grep -Fx %q", filePath, option)
	if out, err := runner(ctx, 30*time.Second, existsCmd); err != nil || strings.TrimSpace(out) == "" {
		// Line not found, nothing to do.
		return nil
	}

	optionCmd := fmt.Sprintf("sed -i '/^%s$/d' %s", option, filePath)
	if _, err := runner(ctx, 30*time.Second, optionCmd); err != nil {
		return errors.Annotate(err, "remove config option: failed to remove line %q to file %q", option, filePath).Err()
	}
	return nil
}

// BuildInitrd builds a new /boot/initrd.img image.
func BuildInitrd(ctx context.Context, runner components.Runner) error {
	const buildCmd = "update-initramfs -v -c -k $(uname -r)"
	if _, err := runner(ctx, 2*time.Minute, buildCmd); err != nil {
		return errors.Annotate(err, "build initrd: failed to re-create initrd.img").Err()
	}

	const copyCmd = `cp "/boot/initrd.img-$(uname -r)" /boot/initrd.img`
	if _, err := runner(ctx, time.Minute, copyCmd); err != nil {
		return errors.Annotate(err, "build initrd: failed to copy initrd.img").Err()
	}
	return nil
}
