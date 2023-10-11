// Copyright (c) 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package puppet

func lastRunFile() (string, error) {
	return "/var/lib/puppet_last_run_summary.yaml", nil
}

func puppetConfFiles() ([]string, error) {
	return []string{"/etc/puppetlabs/puppet/puppet.conf", "/etc/puppet/puppet.conf"}, nil
}

func exitStatusFiles() []string {
	return []string{"/var/log/puppet/puppet_exit_status.txt"}
}
