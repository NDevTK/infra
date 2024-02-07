// Copyright (c) 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package puppet

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

var (
	cachedCommonAppdataPath string
)

func lastRunFile() (string, error) {
	appdata, err := commonAppdataPath()
	if err != nil {
		return "", err
	}
	return appdata + `\PuppetLabs\puppet\var\state\last_run_summary.yaml`, nil
}

func puppetCertPath() (string, error) {
	appdata, err := commonAppdataPath()
	if err != nil {
		return "", err
	}
	return appdata + `\PuppetLabs\puppet\etc\puppet\ssl\certs`, nil
}

func puppetConfFile() (string, error) {
	appdata, err := commonAppdataPath()
	if err != nil {
		return "", err
	}

	filePath := appdata + `\PuppetLabs\puppet\etc\puppet.conf`
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	return "", fmt.Errorf("no conf exists at derived path %s", filePath)
}

func commonAppdataPath() (string, error) {
	if cachedCommonAppdataPath != "" {
		return cachedCommonAppdataPath, nil
	}

	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\Shell Folders`,
		registry.READ)
	if err != nil {
		return "", err
	}

	path, _, err := key.GetStringValue("Common AppData")
	if err == nil {
		cachedCommonAppdataPath = path
	}
	return path, err
}

func exitStatusFiles() []string {
	return []string{
		`C:\chrome-infra-logs\puppet_exit_status.txt`,
		`E:\chrome-infra-logs\puppet_exit_status.txt`,
	}
}
