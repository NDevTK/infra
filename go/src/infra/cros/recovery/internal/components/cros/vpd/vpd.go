// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package vpd provide ability to read and update VPD values.
package vpd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
)

// Read read vpd value of VPD.
func Read(ctx context.Context, ha components.HostAccess, timeout time.Duration, key string) (string, error) {
	return readValue(ctx, ha, "", timeout, key)
}

// ReadRO read vpd value of VPD from RO_VPD partition.
func ReadRO(ctx context.Context, ha components.HostAccess, timeout time.Duration, key string) (string, error) {
	return readValue(ctx, ha, "RO_VPD", timeout, key)
}

func readValue(ctx context.Context, ha components.HostAccess, partition string, timeout time.Duration, key string) (string, error) {
	errorMessage := "read vpd"
	cmd := "vpd"
	if partition != "" {
		cmd = fmt.Sprintf("%s -i %s", cmd, partition)
		errorMessage = fmt.Sprintf("%s of %q partition", errorMessage, partition)
	}
	errorMessage = fmt.Sprintf("%s for %q", errorMessage, key)
	cmd = fmt.Sprintf("%s -g %s", cmd, key)
	res, err := ha.Run(ctx, timeout, cmd)
	if err != nil {
		return "", errors.Annotate(err, errorMessage).Err()
	}
	return strings.TrimSpace(res.GetStdout()), nil
}

// Set sets vpd value of VPD by key.
func Set(ctx context.Context, ha components.HostAccess, timeout time.Duration, key, value string) error {
	cmd := fmt.Sprintf("vpd -s %s=%s", key, value)
	_, err := ha.Run(ctx, timeout, cmd)
	return errors.Annotate(err, "set vpd for %q:%q", key, value).Err()
}
