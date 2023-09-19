// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dns

import (
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// ReadContents gets the content of a DNS file.
// If the DNS file does not exist, replace it with an empty container.
func ReadContents(executor executor.IExecCommander) (string, error) {
	// Defensively touch the file if it does not already exist.
	// See b/199796469 for details.
	args := []string{
		paths.DockerPath,
		"exec",
		"dns",
		"touch",
		paths.HostsFilePath,
	}
	if _, err := executor.Exec(exec.Command(args[0], args[1:]...)); err != nil {
		return "", errors.Annotate(err, "defensively touch dns file").Err()
	}

	args = []string{
		paths.DockerPath,
		"exec",
		"dns",
		"/bin/cat",
		paths.HostsFilePath,
	}

	out, err := executor.Exec(exec.Command(args[0], args[1:]...))
	if err != nil {
		return "", errors.Annotate(err, "get dns file content").Err()

	}

	return strings.TrimRight(string(out), "\n\t"), nil
}

// innerReadHostsToMap is a inner function read a dns file
// and parse the raw data to a map
func innerReadHostsToMap(
	executor executor.IExecCommander,
	useIPAsKey bool,
) (map[string]string, error) {
	res := map[string]string{}
	rawData, err := ReadContents(executor)

	if err != nil {
		return res, nil
	}

	list := strings.Split(rawData, "\n")

	for _, row := range list {
		r := strings.Split(row, "\t")
		// We only handle vaild data
		// e.g. <ip>\t<hostname>
		if len(r) == 2 {
			if useIPAsKey {
				res[r[0]] = r[1]
			} else {
				res[r[1]] = r[0]
			}
		}
	}

	return res, nil

}

// ReadHostsToIPMap read the hosts file to get the IP host mapping
func ReadHostsToIPMap(executor executor.IExecCommander) (map[string]string, error) {
	return innerReadHostsToMap(executor, true)
}

// ReadHostsToHostMap read the hosts file to get the host IP mapping
func ReadHostsToHostMap(executor executor.IExecCommander) (map[string]string, error) {
	return innerReadHostsToMap(executor, false)
}
