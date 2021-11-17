// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package osutils

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func Abs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	return abs
}

// Verify that the given relative path exists.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Verify that a directory exists at the given relative path.
func DirExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// Look for a relative path, ascending through parent directories.
//
// Args:
//  pathToFind: The relative path to look for.
//  startPath: The path to start the search from.  If |startPath| is a
//    directory, it will be included in the directories that are searched.
//  endPath: The path to stop searching.
//  testFunc: The function to use to verify the relative path.
func FindInPathParents(
	pathToFind string,
	startPath string,
	endPath string,
	testFunc func(string) bool) string {

	// Default parameter values.
	if endPath == "" {
		endPath = "/"
	}
	if testFunc == nil {
		testFunc = PathExists
	}
	currentPath := startPath
	for {
		// Test to see if path exists in this directory
		targetPath := filepath.Join(Abs(currentPath), pathToFind)
		if testFunc(targetPath) {
			return Abs(targetPath)
		}
		rel, _ := filepath.Rel(endPath, currentPath)
		if rel == "." || rel == "" {
			// Reached endPath.
			return ""
		}

		// Go up one directory.
		currentPath += "/.."
	}
}

// CreateTmpCopy creates a temporary copy of the given file.
// It returns the file path, a clean up function, and a potential error.
func CreateTmpCopy(src string) (string, func(), error) {
	in, err := os.Open(src)
	if err != nil {
		return "", nil, err
	}
	defer in.Close()

	out, err := ioutil.TempFile("", "tmp_copy")
	if err != nil {
		return "", nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		os.Remove(out.Name())
		return "", nil, err
	}

	return out.Name(), func() {
		os.Remove(out.Name())
	}, nil
}

// ResolveHomeRelPath resolves any references to the invoker's home directory
// in the given path.
func ResolveHomeRelPath(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		path = filepath.Join(dir, path[2:])
	}
	return path, nil
}

// RecursiveChmod changes the mode of each file and directory rooted at root,
// including root.
func RecursiveChmod(root string, mode fs.FileMode) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		return os.Chmod(path, mode)
	})
}
