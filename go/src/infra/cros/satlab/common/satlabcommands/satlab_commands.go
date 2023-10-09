// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package satlabcommands

import (
	"context"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/common/utils/misc"
)

// Decision is a classification of a line in a file.
// Lines may be kept, modified, or deleted.
// Functions that process lines of text are split conceptually
// into a decision which classifies lines and a transformation
// which only applies to selected lines.
type Decision int

const (
	Unknown Decision = iota
	Keep
	Reject
	// Modify is used only by replacing things.
	Modify
)

// OSVersion contains the OS version information
// from the command `get_host_os_version`
type OSVersion struct {
	// Version the ChromeOS verison
	Version     string
	Track       string
	Description string
}

// GetHostIdentifier gets the host identifier value.
//
// Note that this command always returns the identifier in lowercase.
func GetDockerHostBoxIdentifier(ctx context.Context, executor executor.IExecCommander) (string, error) {
	out, err := executor.Exec(exec.CommandContext(ctx, paths.GetHostIdentifierScript))
	// Immediately normalize the satlab prefix to lowercase. It will save a lot of
	// trouble later.
	return strings.ToLower(misc.TrimOutput(out)), errors.Annotate(err, "get host identifier").Err()
}

// parseOutput parse the raw data "<value>\\n"
// removing the `\"` and `\\n`
func parseOutput(s string) string {
	return strings.Trim(s, "\n \"")
}

// GetSatlabVersion gets the Satlab version from docker container `compose` label.
func GetSatlabVersion(ctx context.Context, executor executor.IExecCommander) (string, error) {
	// We want to get the env variables from `docker inspect`
	// The output is like this
	// ```
	// KEY=VALUE
	// KEY=VALUE
	// ...
	// ```
	out, err := executor.Exec(
		exec.CommandContext(
			ctx,
			paths.DockerPath,
			"inspect",
			"--format='{{range .Config.Env}}{{println .}}{{end}}'",
			"compose",
		),
	)
	if err != nil {
		return "", nil
	}
	rawData := strings.Split(string(out), "\n")
	for _, row := range rawData {
		if strings.HasPrefix(row, "LABEL=") {
			r := strings.Split(row, "=")
			return r[1], nil
		}
	}
	return "", errors.New("can't find the version")
}

// GetOsVersion gets the OS GetOsVersion
func GetOsVersion(ctx context.Context, executor executor.IExecCommander) (*OSVersion, error) {
	out, err := executor.Exec(exec.CommandContext(ctx, paths.GetOSVersionScript))
	if err != nil {
		return nil, err
	}
	rawData := strings.Split(string(out), "\n")
	resp := OSVersion{}
	for _, r := range rawData {
		row := strings.Split(r, ":")
		if len(row) == 2 {
			if strings.ToLower(row[0]) == "version" {
				resp.Version = parseOutput(row[1])
			}
			if strings.ToLower(row[0]) == "track" {
				resp.Track = parseOutput(row[1])
			}
			if strings.ToLower(row[0]) == "description" {
				resp.Description = parseOutput(row[1])
			}
		}
	}
	return &resp, nil
}
