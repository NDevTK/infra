// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package log

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"
)

// WriteResourceLogFile writes bytes to a log file with the specified name.
// The log file will be created the log directory for the given resource
// under the rootLogDir.
//
// Returns the file path of the log file, relative to the logging root.
func WriteResourceLogFile(ctx context.Context, rootLogDir string, resourceName string, logFileName string, fileContents []byte) (relativeLogFilePath string, err error) {
	relativeLogFilePath, file, err := OpenResourceLogFile(rootLogDir, resourceName, logFileName)
	if err != nil {
		return "", errors.Annotate(err, "write resource log file: open file").Err()
	}
	defer func() {
		_ = file.Close()
	}()
	writtenBytes, err := file.Write(fileContents)
	if err != nil {
		return "", errors.Annotate(err, "write resource log file: write to file").Err()
	}
	Debugf(ctx, "Wrote %d bytes to resource log file %q", writtenBytes, relativeLogFilePath)
	return relativeLogFilePath, nil
}

// OpenResourceLogFile creates and opens a new file with the provided name in
// the log directory for the given resource under the rootLogDir. All dirs in
// the path are created if they do not exist.
//
// Returns the file path of the log file, relative to the logging root, and the
// open file.
func OpenResourceLogFile(rootLogDir string, resourceName string, logFileName string) (relativeLogFilePath string, logFile *os.File, err error) {
	logFilePath := filepath.Join(rootLogDir, sanitizeFilePathPart(resourceName), logFileName)
	if err := os.MkdirAll(path.Dir(logFilePath), 0755); err != nil {
		return "", nil, errors.Annotate(err, "open resource log file: failed to create basedir for %q", logFilePath).Err()
	}
	logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", nil, errors.Annotate(err, "open resource log file: cannot open file %q", logFilePath).Err()
	}
	relativeLogFilePath = strings.TrimPrefix(rootLogDir, logFileName)
	return relativeLogFilePath, logFile, nil
}

// BuildFilename builds a log filename with a minimal timestamp prefix, all
// the name parts in the middle delimited by "_" with non-word characters
// replaced with underscores, and the provided file extension.
//
// This not only communicates the time of the log to users, but keeps similar
// files in chronological order within the same directory when displayed sorted
// by name (alphanumerical order) by most programs.
//
// Example result:
// ({"some", "log_name"}, "log") => "20220523-122753_some_log_name.log"
func BuildFilename(nameParts []string, ext string) string {
	// Build timestamp prefix.
	timestamp := time.Now().Format("20060102-150405")
	// Join and sanitize name parts.
	name := sanitizeFilePathPart(strings.Join(nameParts, "_"))
	// Combine timestamp, name, and extension.
	if name != "" {
		name = "_" + name
	}
	return fmt.Sprintf("%s%s.%s", timestamp, name, ext)
}

// sanitizeFilePathPart returns the given string with all non-word characters
// and repeated underscores replaced to a single underscore, making it safe to
// use as a file path part.
func sanitizeFilePathPart(pathPart string) string {
	pathPart = regexp.MustCompile(`\W`).ReplaceAllString(pathPart, "_")
	pathPart = regexp.MustCompile(`_+`).ReplaceAllString(pathPart, "_")
	return pathPart
}
