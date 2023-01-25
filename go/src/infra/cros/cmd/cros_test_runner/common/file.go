// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.chromium.org/chromiumos/config/go/build/api"
	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

// CreateTempDir creates temp dir in luci TEMPDIR location.
func CreateTempDir(ctx context.Context, tempDirPattern string) (tempDir string, err error) {
	luciTempDir := os.Getenv("TEMPDIR")
	logging.Infof(ctx, fmt.Sprintf("Luci temp Dir: %s", luciTempDir))
	tempDir, err = os.MkdirTemp(luciTempDir, tempDirPattern)
	if err != nil {
		logging.Infof(ctx, fmt.Sprintf("Temp dir %q creation at %q failed", tempDir, luciTempDir))
	}
	return tempDir, err
}

// CreateImagePath creates image path from container image info.
// Ex: us-docker.pkg.dev/cros-registry/test-services/cros-provision:<tag>
func CreateImagePath(i *buildapi.ContainerImageInfo) (string, error) {
	if i.GetName() == "" {
		return "", errors.Reason("create image path: name is empty").Err()
	}
	if i.GetRepository() == nil {
		return "", errors.Reason("create image path: no repository info").Err()
	}
	r := i.GetRepository()
	if r.GetHostname() == "" || r.GetProject() == "" {
		return "", errors.Reason("create image path: repository info is missing").Err()
	}
	if len(i.GetTags()) == 0 {
		return "", errors.Reason("create image path: no tags found").Err()
	}
	// TODO: update logic ow to choose tags.
	tag := i.GetTags()[0]
	path := fmt.Sprintf("%s/%s/%s:%s", r.GetHostname(), r.GetProject(), i.GetName(), tag)
	return path, nil
}

// GetContainerImageFromMap retrieves the container image from provided map.
func GetContainerImageFromMap(key string, imageMap map[string]*api.ContainerImageInfo) (string, error) {
	if key == "" {
		return "", fmt.Errorf("Provided key is empty!")
	}
	if len(imageMap) == 0 {
		return "", fmt.Errorf("Provided map is empty!")
	}

	containerImageInfo, ok := imageMap[key]
	if !ok {
		return "", fmt.Errorf("Could not find container info for key: %s", key)
	}
	imagePath, err := CreateImagePath(containerImageInfo)
	if err != nil {
		return "", errors.Annotate(err, "%s image path creation: ", key).Err()
	}

	return imagePath, nil
}

// CreateRegistryName creates the Registry name used for authing to docker.
func CreateRegistryName(i *buildapi.ContainerImageInfo) (string, error) {
	if i.GetRepository() == nil {
		return "", errors.Reason("create image path: no repository info").Err()
	}
	r := i.GetRepository()
	if r.GetHostname() == "" || r.GetProject() == "" {
		return "", errors.Reason("create image path: repository info is missing").Err()
	}
	return fmt.Sprintf("%s/%s", r.GetHostname(), r.GetProject()), nil
}

// AddContentsToLog adds contents of the file of fileName to log
func AddFileContentsToLog(
	ctx context.Context,
	fileName string,
	rootDir string,
	msgToAdd string,
	writer io.Writer) error {

	filePath, err := FindFile(ctx, fileName, rootDir)
	if err != nil {
		logging.Infof(ctx, "%s finding file '%s' at '%s' failed:%s", msgToAdd, fileName, rootDir, err)
		return err
	}
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		logging.Infof(ctx, "%s reading file '%s' at '%s' failed:%s", msgToAdd, fileName, filePath, err)
		return err
	}

	_, err = writer.Write(fileContents)
	if err != nil {
		logging.Infof(ctx, "%s writing contains '%s' at '%s' failed:%s", msgToAdd, fileName, rootDir, err)
	}
	logging.Infof(ctx, "%s file '%s' info at '%s':\n\n%s\n", msgToAdd, fileName, filePath, string(fileContents))
	return nil
}

// FindFile finds file path in rootDir of fileName
func FindFile(ctx context.Context, fileName string, rootDir string) (string, error) {
	filePath := ""
	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == fileName {
			filePath = path
		}
		return nil
	})

	if filePath != "" {
		return filePath, nil
	}

	return "", errors.Reason(fmt.Sprintf("file '%s' not found!", fileName)).Err()
}

// StreamLogAsync starts an async reading of log file.
func StreamLogAsync(ctx context.Context, rootDir string, writer io.Writer) (chan<- bool, *sync.WaitGroup, error) {
	// Create channel and waitgroup for proper communication
	var wg sync.WaitGroup
	wg.Add(1)
	taskDone := make(chan bool)

	// Find file in root dir
	fileName := "log.txt"
	filePath, err := FindFile(ctx, fileName, rootDir)
	if err != nil {
		logging.Infof(ctx, "Failed to find file '%s' at '%s' with error:%s", fileName, rootDir, err)
		wg.Done()
		return taskDone, &wg, nil
	}

	// Open the file for reading
	fi, err := os.OpenFile(filePath, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		logging.Infof(ctx, "Failed to open file %s: %s", filePath, err)
		wg.Done()
		return taskDone, &wg, nil
	}

	// Kick off async reading the file and writing contents to writer
	go WriteFromFile(ctx, fi, writer, taskDone, &wg, 3*time.Second)

	// return channels and waitgroup to caller for it to control
	// file reading and writing
	return taskDone, &wg, nil
}

// WriteFromFile writes contents from a file to a provided writer.
func WriteFromFile(
	ctx context.Context,
	fi *os.File,
	writer io.Writer,
	taskDone <-chan bool,
	wg *sync.WaitGroup,
	poll time.Duration) {

	defer fi.Close()
	reader := bufio.NewReader(fi)

	isTaskDone := false
	for !isTaskDone {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			select {
			// Delay reading as we do not want to overwhelm io resources
			case <-time.After(poll):
			// Foreground task is done. prepare to conclude file reading
			case <-taskDone:
				isTaskDone = true
			}
		} else {
			writer.Write(line)
		}
	}

	// Write remaining unwritten bytes if any
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}
		writer.Write(line)
	}
	wg.Done()
}
