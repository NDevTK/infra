// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build linux
// +build linux

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/gerrit"
	"infra/cros/internal/gs"
)

const internalManifest = `<manifest>
	<project path="src/project/foo/bar1"
	remote="cros-internal"
	name="chromeos/project/foo/bar1"
	groups="partner-config" />
	<project path="src/project/foo/bar2"
	remote="cros-internal"
	name="chromeos/project/foo/bar2"
	groups="partner-config" />
</manifest>`

func checkFiles(t *testing.T, path string, expected map[string]string) {
	for filename, expectedContents := range expected {
		data, err := ioutil.ReadFile(filepath.Join(path, filename))
		assert.NilError(t, err)
		assert.StringsEqual(t, string(data), expectedContents)
	}
	// Make sure there are no extraneous files.
	files, err := ioutil.ReadDir(path)
	assert.NilError(t, err)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		_, ok := expected[file.Name()]
		if !ok {
			fmt.Printf("%v\n", file.Name())
		}
		assert.Assert(t, ok)
	}
}

func TestSetupProject(t *testing.T) {
	branch := "mybranch"
	expectedFiles := map[string]string{
		"foo_program.xml":                 "chromeos/program/foo",
		"bar_project.xml":                 "chromeos/project/foo/bar",
		"baz_chipset.xml":                 "chromeos/overlays/chipset-baz-private",
		"chromeos-other-project_repo.xml": "chromeos/other/project",
	}

	expectedDownloads := map[string]map[string]map[string]string{}
	for _, projectName := range expectedFiles {
		expectedDownloads[projectName] = map[string]map[string]string{
			branch: {
				"local_manifest.xml": projectName,
			},
		}
	}
	gc := &gerrit.FakeAPIClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
	}

	dir, err := ioutil.TempDir("", "setup_project")
	defer os.RemoveAll(dir)
	assert.NilError(t, err)
	localManifestDir := filepath.Join(dir, ".repo/local_manifests/")
	assert.NilError(t, os.MkdirAll(localManifestDir, os.ModePerm))

	b := setupProject{
		chromeosCheckoutPath: dir,
		program:              "foo",
		localManifestBranch:  branch,
		project:              "bar",
		chipset:              "baz",
		otherRepos:           []string{"chromeos/other/project"},
	}
	ctx := context.Background()
	assert.NilError(t, b.setupProject(ctx, nil, gc))
	checkFiles(t, localManifestDir, expectedFiles)
}

func TestSetupProject_onlyOtherRepo(t *testing.T) {
	branch := "mybranch"
	expectedFiles := map[string]string{
		"chromeos-other-project_repo.xml": "chromeos/other/project",
	}

	expectedDownloads := map[string]map[string]map[string]string{}
	for _, projectName := range expectedFiles {
		expectedDownloads[projectName] = map[string]map[string]string{
			branch: {
				"local_manifest.xml": projectName,
			},
		}
	}
	gc := &gerrit.FakeAPIClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
	}

	dir, err := ioutil.TempDir("", "setup_project")
	defer os.RemoveAll(dir)
	assert.NilError(t, err)
	localManifestDir := filepath.Join(dir, ".repo/local_manifests/")
	assert.NilError(t, os.MkdirAll(localManifestDir, os.ModePerm))

	b := setupProject{
		chromeosCheckoutPath: dir,
		localManifestBranch:  branch,
		otherRepos:           []string{"chromeos/other/project"},
	}
	ctx := context.Background()
	assert.NilError(t, b.setupProject(ctx, nil, gc))
	checkFiles(t, localManifestDir, expectedFiles)
}

func TestSetupProject_allProjects(t *testing.T) {
	branch := "mybranch"
	expectedFiles := map[string]string{
		"foo_program.xml":  "chromeos/program/foo",
		"bar1_project.xml": "chromeos/project/foo/bar1",
		"bar2_project.xml": "chromeos/project/foo/bar2",
		"baz_chipset.xml":  "chromeos/overlays/chipset-baz-private",
	}

	expectedDownloads := map[string]map[string]map[string]string{}
	for _, projectName := range expectedFiles {
		expectedDownloads[projectName] = map[string]map[string]string{
			branch: {
				"local_manifest.xml": projectName,
			},
		}
	}
	expectedDownloads["chromeos/manifest-internal"] = map[string]map[string]string{
		branch: {
			"internal_full.xml": internalManifest,
		},
	}

	gc := &gerrit.FakeAPIClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
		ExpectedProjects: []string{
			"chromeos/project/foo/bar1",
			"chromeos/project/foo/bar2",
		},
	}

	dir, err := ioutil.TempDir("", "setup_project")
	defer os.RemoveAll(dir)
	assert.NilError(t, err)
	localManifestDir := filepath.Join(dir, ".repo/local_manifests/")
	assert.NilError(t, os.MkdirAll(localManifestDir, os.ModePerm))

	b := setupProject{
		chromeosCheckoutPath: dir,
		program:              "foo",
		localManifestBranch:  branch,
		allProjects:          true,
		project:              "bar",
		chipset:              "baz",
	}
	ctx := context.Background()
	assert.NilError(t, b.setupProject(ctx, nil, gc))
	checkFiles(t, localManifestDir, expectedFiles)
}

func TestSetupProject_buildspecs(t *testing.T) {
	buildspec := "90/13811.0.0.xml"

	gc := &gerrit.FakeAPIClient{
		T: t,
		ExpectedProjects: []string{
			"chromeos/project/foo/bar1",
			"chromeos/project/foo/bar2",
		},
		ExpectedDownloads: map[string]map[string]map[string]string{
			"chromeos/manifest-internal": {
				"main": {
					"internal_full.xml": internalManifest,
				},
			},
		},
	}

	gsSuffix := "/buildspecs/" + buildspec
	expectedDownloads := map[string][]byte{
		"gs://chromeos-foo" + gsSuffix:                          []byte("chromeos/program/foo"),
		"gs://chromeos-foo-bar1" + gsSuffix:                     []byte("chromeos/project/foo/bar1"),
		"gs://chromeos-foo-bar2" + gsSuffix:                     []byte("chromeos/project/foo/bar2"),
		"gs://chromeos-other-project" + gsSuffix:                []byte("chromeos/other/project"),
		"gs://chromeos-overlays-chipset-baz-private" + gsSuffix: []byte("chromeos/overlays/chipset-baz-private"),
	}

	f := &gs.FakeClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
	}

	dir, err := ioutil.TempDir("", "setup_project")
	defer os.RemoveAll(dir)
	assert.NilError(t, err)
	localManifestDir := filepath.Join(dir, ".repo/local_manifests/")
	assert.NilError(t, os.MkdirAll(localManifestDir, os.ModePerm))

	b := setupProject{
		chromeosCheckoutPath: dir,
		program:              "foo",
		localManifestBranch:  "main", // Default value.
		allProjects:          true,
		project:              "bar",
		chipset:              "baz",
		buildspec:            buildspec,
		otherRepos:           []string{"chromeos/other/project"},
	}
	ctx := context.Background()
	assert.NilError(t, b.setupProject(ctx, f, gc))
	expectedFiles := map[string]string{
		"foo_program.xml":                 "chromeos/program/foo",
		"bar1_project.xml":                "chromeos/project/foo/bar1",
		"bar2_project.xml":                "chromeos/project/foo/bar2",
		"chromeos-other-project_repo.xml": "chromeos/other/project",
		"baz_chipset.xml":                 "chromeos/overlays/chipset-baz-private",
	}
	checkFiles(t, localManifestDir, expectedFiles)
}

func TestSetupProject_buildspecs_missingProgram(t *testing.T) {
	buildspec := "90/13811.0.0.xml"

	gsSuffix := "/buildspecs/" + buildspec
	expectedDownloads := map[string][]byte{
		"gs://chromeos-foo" + gsSuffix:     nil,
		"gs://chromeos-foo-bar" + gsSuffix: []byte("chromeos/project/foo/bar"),
	}

	f := &gs.FakeClient{
		T:                 t,
		ExpectedDownloads: expectedDownloads,
	}

	dir, err := ioutil.TempDir("", "setup_project")
	defer os.RemoveAll(dir)
	assert.NilError(t, err)
	localManifestDir := filepath.Join(dir, ".repo/local_manifests/")
	assert.NilError(t, os.MkdirAll(localManifestDir, os.ModePerm))

	b := setupProject{
		chromeosCheckoutPath: dir,
		program:              "foo",
		project:              "bar",
		buildspec:            buildspec,
	}
	ctx := context.Background()
	assert.NilError(t, b.setupProject(ctx, f, nil))
	expectedFiles := map[string]string{
		"bar_project.xml": "chromeos/project/foo/bar",
	}
	checkFiles(t, localManifestDir, expectedFiles)
}
