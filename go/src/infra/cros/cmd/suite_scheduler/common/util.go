// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package common has utilities that are not context specific and can be used by
// all packages.
package common

import (
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	infrapb "go.chromium.org/chromiumos/infra/proto/go/testplans"
)

// IsTimedEvent returns if the given config is a timed event or a build event type.
func IsTimedEvent(config *infrapb.SchedulerConfig) bool {
	return config.LaunchCriteria.LaunchProfile == infrapb.SchedulerConfig_LaunchCriteria_DAILY ||
		config.LaunchCriteria.LaunchProfile == infrapb.SchedulerConfig_LaunchCriteria_FORTNIGHTLY ||
		config.LaunchCriteria.LaunchProfile == infrapb.SchedulerConfig_LaunchCriteria_WEEKLY
}

// ReadLocalFile reads a file at the given path into memory and returns it's contents.
func ReadLocalFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	return data, err
}

// FetchFileFromURL retrieves text from the given URL. It assumes the text received
// will be base64 encoded.
func FetchFileFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	fileText, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return []byte{}, err
	}

	return fileText, nil
}

// WriteToFile copies the given data into a file with read/write permissions. If
// the directory structure given does not exist at the time of calling then this
// function will create it.
func WriteToFile(path string, data []byte) error {
	// Ensure path exists
	finalDir := filepath.Dir(path)
	err := os.MkdirAll(finalDir, fs.FileMode(os.O_RDWR))
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0664)
}

// FetchAndWriteFile retrieves a text file from the specified URL and writes
// into into the given path. The function will automatically create the
// directory structure if it does not exist at the time of calling.
func FetchAndWriteFile(url, path string) error {
	data, err := FetchFileFromURL(url)
	if err != nil {
		return err
	}

	return WriteToFile(path, data)
}

// HasString checks to see if the given string array has the target string in
// it.
func HasString(target string, strings []string) bool {
	found := false
	for _, item := range strings {
		if target == item {
			found = true
			break
		}
	}

	return found
}
