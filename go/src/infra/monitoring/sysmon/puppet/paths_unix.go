// Copyright (c) 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package puppet

import (
	"fmt"
	"os"
)

func lastRunFile() (string, error) {
	return "/var/lib/puppet_last_run_summary.yaml", nil
}

func puppetConfFile() (string, error) {
	confPaths := []string{"/etc/puppetlabs/puppet/puppet.conf", "/etc/puppet/puppet.conf"}
	for _, filePath := range confPaths {
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	}
	return "", fmt.Errorf("No puppet.conf found in either location: %s", confPaths)
}

func exitStatusFiles() []string {
	return []string{"/var/log/puppet/puppet_exit_status.txt"}
}
