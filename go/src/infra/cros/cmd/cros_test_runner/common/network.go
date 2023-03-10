// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// GetHostIp returns the IP address that is accessible from outside the host
func GetHostIp() (string, error) {
	cmd := exec.Command("hostname", "-I")
	stdout, stderr, err := RunCommand(context.Background(), cmd, "hostname", nil, true)
	if err != nil {
		return "", errors.Annotate(err, "Unable to find localhost IP: "+stderr).Err()
	}
	if strings.TrimSpace(stdout) == "" {
		return "", errors.New("Unable to find localhost IP: hostname -I returns no results")
	}
	return strings.Fields(stdout)[0], nil
}
