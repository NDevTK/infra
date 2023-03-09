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
	"strings"
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

// CheckIfFileExists checks if file exists at the provided path.
func CheckIfFileExists(filePath string) error {
	// File exists
	if _, err := os.Stat(filePath); err == nil {
		return nil
		// File does not exist
	} else if errors.Is(err, os.ErrNotExist) {
		return errors.Annotate(err, "Failed to find file at provided path %q : ", filePath).Err()
		// Unexpected error
	} else {
		return errors.Annotate(err, "Unexpected error while finding file at provided path %q : ", filePath).Err()
	}
}

// WriteToExistingFile writes provided contents to existing file.
func WriteToExistingFile(ctx context.Context, filePath string, contents string) error {
	if err := CheckIfFileExists(filePath); err != nil {
		return errors.Annotate(err, "Could not find file at: %s", filePath).Err()
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Annotate(err, "Error while opening file at: %s", filePath).Err()
	}
	defer file.Close()

	_, err = file.Write([]byte(contents))
	if err != nil {
		return errors.Annotate(err, "Error while writing to file at: %s", filePath).Err()
	}
	return nil
}

// GetFileContentsInMap returns a map of file contents where the contents on
// each line is separated by provided separator. Will return error if
// each line is not formatted like 'key<separator>value'. If writer provided,
// file contents will be written to writer simultaneously.
func GetFileContentsInMap(ctx context.Context, filePath string, contentSeparator string, writer io.Writer) (map[string]string, error) {
	if err := CheckIfFileExists(filePath); err != nil {
		return nil, errors.Annotate(err, "Could not find file at: %s", filePath).Err()
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Annotate(err, "Error while opening file at: %s", filePath).Err()
	}
	defer file.Close()

	fileContentsMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if writer != nil {
			writer.Write([]byte(fmt.Sprintln(scanner.Text())))
		}
		lineContents := strings.Split(scanner.Text(), contentSeparator)
		if len(lineContents) != 2 {
			return nil, fmt.Errorf("Line contents %q could not be separated by provided separator %q.", scanner.Text(), contentSeparator)
		}
		fileContentsMap[lineContents[0]] = lineContents[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Annotate(err, "Error while scanning file contents at %q: ", filePath).Err()
	}

	return fileContentsMap, nil
}

// GetCftServiceMetadataFromFile waits for the service metadata and returns
// metadata if found.
func GetCftServiceMetadataFromFile(ctx context.Context, metadataFilePath string, fileLog io.Writer) (map[string]string, error) {
	if metadataFilePath == "" {
		return nil, fmt.Errorf("Cannot get service metadata from empty file path.")
	}

	var err error

	// TODO (azrahman): use exponential backoff retry
	fileFound := false
	retryCount := 50 // This number is currently a bit high due to drone's lower than expected performance
	timeout := 5 * time.Second

	for !fileFound && retryCount > 0 {
		if err = CheckIfFileExists(metadataFilePath); err == nil {
			fileFound = true
		}
		retryCount = retryCount - 1
		time.Sleep(timeout)
	}

	logging.Infof(ctx, "filefound: %v, remainingretrycount: %v, timeout: %v", fileFound, retryCount, timeout)
	if !fileFound {
		return nil, errors.Annotate(err, "Error while retrieving service metadata: ").Err()
	}

	fileContentsMap, err := GetFileContentsInMap(ctx, metadataFilePath, CftServiceMetadataLineContentSeparator, fileLog)
	if err != nil {
		return nil, errors.Annotate(err, "Error while retrieving service metadata: ").Err()
	}

	return fileContentsMap, nil
}

// GetCftLocalServerAddress waits for the service metadata file and retrieves
// server address for localhost.
func GetCftLocalServerAddress(ctx context.Context, metadataFilePath string, fileLog io.Writer) (string, error) {
	if metadataFilePath == "" {
		return "", fmt.Errorf("Cannot get server address from empty file path.")
	}

	serviceMatadata, err := GetCftServiceMetadataFromFile(ctx, metadataFilePath, fileLog)
	if err != nil {
		return "", err
	}
	port, ok := serviceMatadata[CftServiceMetadataServicePortKey]
	if !ok || port == "" {
		return "", fmt.Errorf("Service port was not found in service metadata.")
	}

	serverAddress := fmt.Sprintf("localhost:%s", port)

	logging.Infof(ctx, "Server address found: %s", serverAddress)

	return serverAddress, nil
}
