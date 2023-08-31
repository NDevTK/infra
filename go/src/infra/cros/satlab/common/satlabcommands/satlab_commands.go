// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package satlabcommands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
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

// SubCommand provides the methods that will execute a command
type SubCommand struct {
	// the command executor that we can provide the different
	// executor for testing or real environment
	ExecCommander executor.IExecCommander
}

// NewSubCommand a constructor that create a SubCommand
func NewSubCommand() *SubCommand {
	return &SubCommand{
		ExecCommander: &executor.ExecCommander{},
	}
}

// GetHostIdentifier gets the host identifier value.
//
// Note that this command always returns the identifier in lowercase.
func (s *SubCommand) GetDockerHostBoxIdentifier() (string, error) {
	fmt.Fprintf(os.Stderr, "Get host identifier: run %s\n", paths.GetHostIdentifierScript)
	out, err := s.ExecCommander.Exec(exec.Command(paths.GetHostIdentifierScript))
	// Immediately normalize the satlab prefix to lowercase. It will save a lot of
	// trouble later.
	return strings.ToLower(TrimOutput(out)), errors.Annotate(err, "get host identifier").Err()
}

// parseOutput parse the raw data "<value>\\n"
// removing the `\"` and `\\n`
func parseOutput(s string) string {
	s = strings.TrimSpace(s)
	res := s
	if strings.HasPrefix(res, "\"") {
		res = res[1:]
	}
	if strings.HasSuffix(res, "\"") {
		res = res[:len(res)-1]
	}
	if strings.HasSuffix(res, "\\n") {
		res = res[:len(res)-3]
	}
	return res
}

// GetOsVersion gets the OS GetOsVersion
func (s *SubCommand) GetOsVersion() (*OSVersion, error) {
	fmt.Fprintf(os.Stderr, "Get host identifier: run %s\n", paths.GetHostIdentifierScript)
	out, err := s.ExecCommander.Exec(exec.Command(paths.GetOSVersionScript))
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

// GetSatlabVersion gets the Satlab version from docker container `compose` label.
func (s *SubCommand) GetSatlabVersion() (string, error) {
	out, err := s.ExecCommander.Exec(
		exec.Command(
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
