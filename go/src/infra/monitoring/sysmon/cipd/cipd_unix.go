// Copyright (c) 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package cipd

var versionDirs = []string{
	"/opt/cq-canary",
	"/opt/cq-stable",
	"/opt/infra-python",
	"/opt/infra-tools", // luci-auth cipd version file is here
	"/opt/infra-tools/.versions",
}
