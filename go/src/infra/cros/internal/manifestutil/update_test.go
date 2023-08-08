// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//go:build linux
// +build linux

package manifestutil

import (
	"io/ioutil"
	"testing"

	"infra/cros/internal/assert"
	"infra/cros/internal/osutils"
	"infra/cros/internal/repo"
)

func TestGetSetDelAttr(t *testing.T) {
	tag := `<default foo="123" bar="456" baz="789" />`

	assert.StringsEqual(t, getAttr(tag, "foo"), "123")
	assert.StringsEqual(t, getAttr(tag, "bar"), "456")
	assert.StringsEqual(t, getAttr(tag, "baz"), "789")

	assert.StringsEqual(t, setAttr(tag, "foo", "000"), `<default foo="000" bar="456" baz="789" />`)
	assert.StringsEqual(t, setAttr(tag, "bar", "000"), `<default foo="123" bar="000" baz="789" />`)
	assert.StringsEqual(t, setAttr(tag, "baz", "000"), `<default foo="123" bar="456" baz="000" />`)

	assert.StringsEqual(t, delAttr(tag, "foo"), `<default bar="456" baz="789" />`)
	assert.StringsEqual(t, delAttr(tag, "bar"), `<default foo="123" baz="789" />`)
	assert.StringsEqual(t, delAttr(tag, "baz"), `<default foo="123" bar="456" />`)
}

func TestUpdateManifestElements(t *testing.T) {
	input, err := ioutil.ReadFile("test_data/update/pre.xml")
	assert.NilError(t, err)

	referenceManifest := &repo.Manifest{
		Default: repo.Default{
			RemoteName: "chromeos1",
			Revision:   "456",
		},
		Remotes: []repo.Remote{
			{
				Name:  "chromium",
				Alias: "chromeos1",
				Fetch: "https://chromium.org/remote",
			},
		},
		Projects: []repo.Project{
			{
				Name:       "baz",
				Path:       "baz/",
				RemoteName: "chromium1",
			},
			{
				Name:       "buz1",
				Path:       "buz/",
				RemoteName: "google",
			},
		},
	}

	got, err := UpdateManifestElements(referenceManifest, input)
	assert.NilError(t, err)

	expected, err := ioutil.ReadFile("test_data/update/post.xml")
	assert.NilError(t, err)
	if string(got) != string(expected) {
		t.Fatalf("mismatch on UpdateManifestElements(...)\ngot:%v\n\nexpected:%v\n\n", string(got), string(expected))
	}
}

func TestUpdateManifestElementsInFile(t *testing.T) {
	tmpFile, cleanup, err := osutils.CreateTmpCopy("test_data/update/pre.xml")
	assert.NilError(t, err)
	defer cleanup()

	referenceManifest := &repo.Manifest{
		Default: repo.Default{
			RemoteName: "chromeos1",
			Revision:   "456",
		},
		Remotes: []repo.Remote{
			{
				Name:  "chromium",
				Alias: "chromeos1",
				Fetch: "https://chromium.org/remote",
			},
		},
		Projects: []repo.Project{
			{
				Name:       "baz",
				Path:       "baz/",
				RemoteName: "chromium1",
			},
			{
				Name:       "buz1",
				Path:       "buz/",
				RemoteName: "google",
			},
		},
	}

	changed, err := UpdateManifestElementsInFile(tmpFile, referenceManifest)
	assert.NilError(t, err)
	assert.Assert(t, changed)

	expected, err := ioutil.ReadFile("test_data/update/post.xml")
	assert.NilError(t, err)

	got, err := ioutil.ReadFile(tmpFile)
	assert.NilError(t, err)

	if string(got) != string(expected) {
		t.Fatalf("mismatch on UpdateManifestElementsInFile(...)\ngot:%v\n\nexpected:%v\n\n", string(got), string(expected))
	}
}

func TestUpdateManifestElements_extraneous(t *testing.T) {
	input, err := ioutil.ReadFile("test_data/update/pre.xml")
	assert.NilError(t, err)

	referenceManifest := &repo.Manifest{
		Remotes: []repo.Remote{
			{
				Name: "extraneous",
			},
		},
	}

	_, err = UpdateManifestElements(referenceManifest, input)
	assert.ErrorContains(t, err, "contained remote(s)")

	referenceManifest = &repo.Manifest{
		Projects: []repo.Project{
			{
				Path: "extraneous/",
			},
		},
	}

	_, err = UpdateManifestElements(referenceManifest, input)
	assert.ErrorContains(t, err, "contained project(s)")

	input, err = ioutil.ReadFile("test_data/update/no_default.xml")
	assert.NilError(t, err)
	referenceManifest = &repo.Manifest{
		Default: repo.Default{
			RemoteName: "foo",
			Revision:   "bar",
		},
	}

	_, err = UpdateManifestElements(referenceManifest, input)
	assert.ErrorContains(t, err, "contained default")
}

func TestUpdateManifestElementsStrict(t *testing.T) {
	input, err := ioutil.ReadFile("test_data/update/pre.xml")
	assert.NilError(t, err)

	referenceManifest := &repo.Manifest{
		Default: repo.Default{
			RemoteName: "chromeos1",
			Revision:   "456",
		},
		Remotes: []repo.Remote{
			{
				Name:  "chromium",
				Alias: "chromeos1",
				Fetch: "https://chromium.org/remote",
			},
		},
		Projects: []repo.Project{
			{
				Name:       "baz",
				Path:       "baz/",
				RemoteName: "chromium1",
			},
			{
				Name:       "buz1",
				Path:       "buz/",
				RemoteName: "google",
			},
		},
	}

	got, err := UpdateManifestElementsStrict(referenceManifest, input)
	assert.NilError(t, err)

	expected, err := ioutil.ReadFile("test_data/update/post_strict.xml")
	assert.NilError(t, err)
	if string(got) != string(expected) {
		t.Fatalf("mismatch on UpdateManifestElementsStrict(...)\ngot:%v\n\nexpected:%v\n\n", string(got), string(expected))
	}
}
