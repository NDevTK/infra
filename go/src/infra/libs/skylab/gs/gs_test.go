// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/googleapi"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/gcloud/gs"
	gcgs "go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/retry"
)

// Implements InnerClient interface, writing to provided local directory instead
// of Google Storage.
type fakeInnerClient struct{}

// fakeWriter implements gs.Writer interface.
type fakeWriter struct {
	os.File
}

func (w *fakeWriter) Count() int64 {
	// Not _really_ implemented.
	return 0
}

func (c *fakeInnerClient) NewWriter(p gcgs.Path) (gs.Writer, error) {
	// This assumes that the incoming path is a valid local path, as opposed to
	// a GS URL (starting with gs://).
	l := string(p)
	d := filepath.Dir(l)
	if err := os.MkdirAll(d, 0777); err != nil {
		return nil, errors.Reason("failed to create directory %s: %s", d, err).Err()
	}
	f, err := os.Create(l)
	if err != nil {
		return nil, err
	}
	return &fakeWriter{File: *f}, nil
}

type testFixture struct {
	// Temporary source directory to copy files from.
	src string
	// Temporary destination directory to copy files to.
	dst string

	// A DirWriter instance to test.
	w *DirWriter
}

// Creates a new test fixture, taking care of common boilerplate.
//
// Returns a function that must be deferred for cleaning up temporary
// directories.
func newTestFixture(t *testing.T) (*testFixture, func()) {
	t.Helper()

	tmp, err := ioutil.TempDir("", "phosphorus")
	if err != nil {
		t.Fatalf("Failed to create temporary directory")
	}

	closer := func() {
		if err := os.RemoveAll(tmp); err != nil {
			panic(fmt.Sprintf("Failed to delete temporary directory %s: %s", tmp, err))
		}
	}

	src := filepath.Join(tmp, "src")
	if err := os.Mkdir(src, 0777); err != nil {
		closer()
		t.Fatalf("Failed to create source directory: %s", err)
	}
	dst := filepath.Join(tmp, "dst")
	if err := os.Mkdir(dst, 0777); err != nil {
		closer()
		t.Fatalf("Failed to create destination directory: %s", err)
	}

	return &testFixture{
		src: src,
		dst: dst,
		w: &DirWriter{
			client:               &fakeInnerClient{},
			maxConcurrentUploads: 1,
			retryIterator:        retry.None(),
		},
	}, closer
}

func TestUploadSingleFile(t *testing.T) {
	f, closer := newTestFixture(t)
	defer closer()
	s, err := os.Create(filepath.Join(f.src, "regular.txt"))
	if err != nil {
		t.Fatalf("Failed to create source file: %s", err)
	}
	defer s.Close()
	if err := f.w.WriteDir(context.Background(), f.src, gcgs.Path(f.dst)); err != nil {
		t.Fatalf("Error writing directory: %s", err)
	}
	if _, err := os.Stat(filepath.Join(f.dst, "regular.txt")); os.IsNotExist(err) {
		t.Errorf("Regular file not copied. os.Stat() returned: %s", err)
	}
}

func TestConcurrentUploads(t *testing.T) {
	tf, closer := newTestFixture(t)
	defer closer()

	const numDirs = 10
	const numFilesPerDir = 100
	files := make([]string, numDirs*numFilesPerDir)
	for i := 0; i < numDirs; i++ {
		subdir := fmt.Sprintf("d%d", i)
		if err := os.Mkdir(filepath.Join(tf.src, subdir), 0755); err != nil {
			t.Fatalf("Failed to create source directory %s: %s", subdir, err)
		}
		for j := 0; j < numFilesPerDir; j++ {
			f := filepath.Join(subdir, fmt.Sprintf("f%d", j))
			files[i*numFilesPerDir+j] = f
			s, err := os.Create(filepath.Join(tf.src, f))
			if err != nil {
				t.Fatalf("Failed to create source file %s: %s", f, err)
			}
			s.Close()
		}
	}

	tf.w.maxConcurrentUploads = 5
	if err := tf.w.WriteDir(context.Background(), tf.src, gcgs.Path(tf.dst)); err != nil {
		t.Fatalf("Error writing directory: %s", err)
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(tf.dst, f)); os.IsNotExist(err) {
			t.Errorf("File %s not copied. os.Stat() returned: %s", f, err)
		}
	}
}

// TestExtractCloudErrorCode tests that we can extract a response code from an arbitrarily wrapped googleapi.Error.
func TestExtractCloudErrorCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		in       error
		cloudErr *googleapi.Error
		code     int
	}{
		{
			name:     "nil",
			in:       nil,
			cloudErr: nil,
			code:     0,
		},
		{
			name:     "non-cloud error",
			in:       fmt.Errorf("I am not a cloud error"),
			cloudErr: nil,
			code:     0,
		},
		{
			name: "403 cloud error",
			in: &googleapi.Error{
				Code: 403,
			},
			cloudErr: &googleapi.Error{
				Code: 403,
			},
			code: 403,
		},
		{
			name: "wrapped cloud error",
			in: fmt.Errorf("wrapping an error: %w", &googleapi.Error{
				Code: 403,
			}),
			cloudErr: &googleapi.Error{
				Code: 403,
			},
			code: 403,
		},
		{
			name: "LUCI wrapped cloud error",
			in: errors.Annotate(
				&googleapi.Error{
					Code: 403,
				},
				"something",
			).Err(),
			cloudErr: &googleapi.Error{
				Code: 403,
			},
			code: 403,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualErr, actualCode := extractCloudErrorCode(tt.in)

			if diff := cmp.Diff(tt.cloudErr, actualErr, cmp.AllowUnexported(googleapi.Error{})); diff != "" {
				t.Errorf("cloudErr differs (-want +got): %s", diff)
			}

			if diff := cmp.Diff(tt.code, actualCode); diff != "" {
				t.Errorf("error code differs (-want +got): %s", diff)
			}
		})
	}
}
