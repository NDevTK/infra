// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package plugsupport

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"go.chromium.org/luci/common/errors"
)

// ProjectDir is an absolute path to a project directory.
type ProjectDir string

func (p ProjectDir) PluginDir() string {
	return filepath.Join(string(p), "_plugin")
}
func (p ProjectDir) ConfigDir() string {
	return filepath.Join(string(p), ".migration")
}
func (p ProjectDir) ConfigFile() string {
	return filepath.Join(p.ConfigDir(), "_plugin")
}
func (p ProjectDir) TrashDir() string {
	return filepath.Join(string(p), ".trash")
}
func (p ProjectDir) ReportPath() string {
	return filepath.Join(string(p), "scan.csv")
}

func (p ProjectDir) ProjectLog(projectID string) string {
	return filepath.Join(string(p), projectID+".scan.log")
}
func (p ProjectDir) ProjectRepoTemp(projectID string) string {
	return filepath.Join(p.TrashDir(), projectID)
}
func (p ProjectDir) ProjectRepo(projectID string) string {
	return filepath.Join(string(p), projectID)
}

func (p ProjectDir) MkTempDir() (string, error) {
	if err := os.Mkdir(p.TrashDir(), 0777); err != nil {
		return "", err
	}
	return ioutil.TempDir(p.TrashDir(), "")
}
func (p ProjectDir) CleanTrash() error {
	return os.RemoveAll(p.TrashDir())
}

func FindProjectRoot(abspath string) (ProjectDir, error) {
	curPath := ProjectDir(abspath)
	for {
		if st, err := os.Stat(curPath.ConfigFile()); err == nil {
			if st.Mode().IsRegular() {
				return curPath, nil
			}
		}
		newPath := ProjectDir(filepath.Dir(string(curPath)))
		if newPath == curPath {
			break
		}
		curPath = newPath
	}
	return "", errors.Reason("not in a migrator project: %q", abspath).Err()
}
