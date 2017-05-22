// Copyright 2017 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"

	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"
	"github.com/luci/luci-go/common/system/environ"
	"github.com/luci/luci-go/common/system/exitcode"
	"github.com/luci/luci-go/common/system/filesystem"
)

// SystemProbe can Locate a Target executable by probing the local system PATH.
type SystemProbe struct {
	// Target is the name of the target (as seen by exec.LookPath) that we are
	// searching for.
	Target string

	// RelativePathOverride is a series of forward-slash-delimited paths to
	// directories relative to the Git wrapper executable that will be checked
	// prior to checking PATH. This allows bundles (e.g., CIPD) that include both
	// the Git wrapper and a Git implementation, to force the Git wrapper to use
	// the bundled Git.
	RelativePathOverride []string

	// testRunCommand is a testing stub that, if not nil, will be used
	// to run the wrapper check command instead of actually running it.
	testRunCommand func(cmd *exec.Cmd) (int, error)
}

// Locate attempts to locate the system's Target by traversing the available
// PATH.
//
// self is the path of the currently-running executable. It may be empty or
// invalid if the current executable could not be identified, or if it is no
// longer available at that location.
//
// cached is the cached path, passed from wrapper to wrapper through the a
// State struct in the environment. This may be empty, if there was no cached
// path or if the cached path was invalid.
//
// env is the environment to operate with, and will not be modified during
// execution.
func (p *SystemProbe) Locate(c context.Context, self, cached string, env environ.Env) (string, error) {
	// Stat "self" to ensure that we exist. We will use this later to assert that
	// our system target is not the same file as self.
	//
	// This may fail if we have been deleted since running. If so, we will skip
	// the SameFile check.
	var selfDir string
	var selfStat os.FileInfo
	if self != "" {
		selfDir = filepath.Dir(self)

		var err error
		if selfStat, err = os.Stat(self); err != nil {
			logging.Debugf(c, "Failed to stat self [%s]: %s", self, err)
		}
	}

	// If we have a cached path, check that it exists and is executable and use it
	// if it is.
	if cached != "" {
		switch cachedStat, err := os.Stat(cached); {
		case err == nil:
			// Use the cached path. First, pass it through a sanity check to ensure
			// that it is not self.
			if selfStat == nil || !os.SameFile(selfStat, cachedStat) {
				logging.Debugf(c, "Using cached Git: %s", cached)
				return cached, nil
			}
			logging.Debugf(c, "Cached value [%s] is this wrapper [%s]; ignoring.", cached, self)

		case os.IsNotExist(err):
			// Our cached path doesn't exist, so we will have to look for a new one.

		case err != nil:
			// We couldn't check our cached path, so we will have to look for a new
			// one. This is an unexpected error, though, so emit it.
			logging.Debugf(c, "Failed to stat cached [%s]: %s", cached, err)
		}
	}

	// Get stats on our parent directory. This may fail; if so, we'll skip the
	// SameFile check.
	var selfDirStat os.FileInfo
	if selfDir != "" {
		var err error
		if selfDirStat, err = os.Stat(selfDir); err != nil {
			logging.Debugf(c, "Failed to stat self directory [%s]: %s", selfDir, err)
		}
	}

	// We determine if it is a wrapper by executing it with a State that has
	// "checkWrapper" set to true. Since we will do this repeatedly, we will
	// generate the "check enabled" environment once and reuse it for each check.
	checkEnv := env.Clone()
	checkEnv.Set(gitWrapperCheckENV, "1")

	// Walk through PATH. Our goal is to find the first program named Target that
	// isn't self and doesn't identify as a wrapper.
	origPATH, _ := env.Get("PATH")
	pathParts := strings.Split(origPATH, string(os.PathListSeparator))

	// Build our list of directories to check for Git.
	checkDirs := make([]string, 0, len(pathParts)+len(p.RelativePathOverride))
	if selfDir != "" {
		for _, rpo := range p.RelativePathOverride {
			checkDirs = append(checkDirs, filepath.Join(selfDir, filepath.FromSlash(rpo)))
		}
	}
	checkDirs = append(checkDirs, pathParts...)

	// Iterate through each check directory and look for a Git candidate within
	// it.
	checked := make(map[string]struct{}, len(checkDirs))
	for _, dir := range checkDirs {
		if _, ok := checked[dir]; ok {
			continue
		}
		checked[dir] = struct{}{}

		path := p.checkDir(c, dir, selfStat, selfDirStat, checkEnv)
		if path != "" {
			return path, nil
		}
	}

	return "", errors.Reason("could not find target in system").
		D("target", p.Target).
		D("PATH", origPATH).
		Err()
}

// checkDir checks "checkDir" for our Target executable. It ignores
// executables whose target is the same file or shares the same parent directory
// as "self".
func (p *SystemProbe) checkDir(c context.Context, dir string, self, selfDir os.FileInfo, checkEnv environ.Env) string {
	// If we have a self directory defined, ensure that "dir" isn't the same
	// directory. If it is, we will ignore this option, since we are looking for
	// something outside of the wrapper directory.
	if selfDir != nil {
		switch checkDirStat, err := os.Stat(dir); {
		case err == nil:
			// "dir" exists; if it is the same as "selfDir", we can ignore it.
			if os.SameFile(selfDir, checkDirStat) {
				logging.Debugf(c, "Candidate shares wrapper directory [%s]; skipping...", dir)
				return ""
			}

		case os.IsNotExist(err):
			logging.Debugf(c, "Candidate directory does not exist [%s]; skipping...", dir)
			return ""

		default:
			logging.Debugf(c, "Failed to stat candidate directory [%s]: %s", dir, err)
			return ""
		}
	}

	t, err := findInDir(p.Target, dir, checkEnv)
	if err != nil {
		return ""
	}

	// Make sure this file isn't the same as "self", if available.
	if self != nil {
		switch st, err := os.Stat(t); {
		case err == nil:
			if os.SameFile(self, st) {
				return ""
			}

		case os.IsNotExist(err):
			// "t" no longer exists, so we can't use it.
			return ""

		default:
			logging.Debugf(c, "Failed to stat candidate path [%s]: %s", t, err)
			return ""
		}
	}

	if err := filesystem.AbsPath(&t); err != nil {
		logging.Debugf(c, "Failed to normalize candidate path [%s]: %s", t, err)
		return ""
	}

	// Try running the candidate command and confirm that it is not a wrapper.
	switch isWrapper, err := p.checkForWrapper(c, t, checkEnv); {
	case err != nil:
		logging.Debugf(c, "Failed to check if [%s] is a wrapper: %s", t, err)
		return ""

	case isWrapper:
		logging.Debugf(c, "Candidate is a Git wrapper: %s", t)
		return ""
	}

	return t
}

// checkForWrapper executes the target path and determines if it is a wrapper.
//
// The environment that we run "path" with has the "checkWrapper" State
// flag set to true. This means that if "path" is a wrapper, it will exit
// immediately with a non-zero return code.
//
// We will run the "version" command, which should be very safe and return
// a "0". If, for whatever, reason, "path" fails returns a non-zero even if it
// isn't a wrapper, we dismiss it as unsuitable.
func (p *SystemProbe) checkForWrapper(c context.Context, path string, checkEnv environ.Env) (bool, error) {
	cmd := exec.CommandContext(c, path, "version")
	cmd.Env = checkEnv.Sorted()

	runCommand := p.testRunCommand
	if runCommand == nil {
		// (Production)
		runCommand = func(cmd *exec.Cmd) (int, error) {
			if err := cmd.Run(); err != nil {
				if rc, ok := exitcode.Get(err); ok {
					return rc, nil
				}

				logging.Warningf(c, "Failed to run check command [%s] with environment: %s", path, strings.Join(checkEnv.Sorted(), " "))
				return 0, errors.Annotate(err).Reason("failed to run check command").Err()
			}
			return 0, nil
		}
	}

	// Run the command. If it returns non-zero, then "path" is considered a
	// wrapper.
	rc, err := runCommand(cmd)
	if err != nil {
		return false, err
	}
	return (rc != 0), nil
}
