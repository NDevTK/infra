// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package configparser implements logic to handle SuiteScheduler configuration files.
package configparser

import (
	"encoding/base64"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// TODO(b/305286743): Update the links to use the .cfg files once StarDocker
	// begins copy them into the SuSch repo.
	SuiteSchedulerCfgURL = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/suite_scheduler.ini?format=text"
	LabCfgURL            = "https://chromium.googlesource.com/chromiumos/infra/suite_scheduler/+/refs/heads/main/generated_configs/lab_config.ini?format=text"
)

// fetchFileText retrieves text from the given URL. It assumes the text received
// will be base64 encoded.
func fetchFileText(url string) ([]byte, error) {
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

// writeToFile copies the given data into a file with read/write permissions. If
// the directory structure given does not exist at the time of calling then this
// function will create it.
func writeToFile(path string, data []byte) error {
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
	data, err := fetchFileText(url)
	if err != nil {
		return err
	}

	return writeToFile(path, data)
}
