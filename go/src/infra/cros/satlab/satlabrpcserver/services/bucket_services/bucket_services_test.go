// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package bucket_services

import (
	"context"
	"io"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iterator"
)

type fakeData struct {
	s        []byte
	i        int64
	prevRune int
}

func (f *fakeData) Read(b []byte) (n int, err error) {
	if f.i >= int64(len(f.s)) {
		return 0, io.EOF
	}
	f.prevRune = -1
	n = copy(b, f.s[f.i:])
	f.i += int64(n)
	return
}

func (f *fakeData) Close() error {
	return nil
}

type fakeObjectIter struct {
	data []string
	i    int
}

func (f *fakeObjectIter) Next() (*storage.ObjectAttrs, error) {
	for {
		if f.i == len(f.data) {
			return nil, iterator.Done
		}
		d := f.data[f.i]
		f.i += 1
		return &storage.ObjectAttrs{Name: d}, nil
	}
}

func createFakeObject() (io.ReadCloser, error) {
	return &fakeData{
		s:        []byte(`{"suite": [{"name": "audio"}], "test": [{"autotest": {"name": "t1", "test_args": "args"}}]}`),
		i:        0,
		prevRune: -1,
	}, nil
}

func Test_ListTestPlans(t *testing.T) {
	ctx := context.Background()

	expected := []string{
		"testplan1.json",
		"testplan2.json",
	}

	// Create a Mock `IBucketService`
	var mockBucketService = new(MockBucketServices)
	mockBucketService.
		On("QueryObjects", ctx, mock.Anything).
		Return(&fakeObjectIter{
			data: []string{
				"testplans/testplan1.json",
				"testplans/testplan2.json",
			},
			i: 0,
		}, nil)

	resp, err := innerListTestplans(ctx, mockBucketService)

	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
	}

	if diff := cmp.Diff(resp, expected); diff != "" {
		t.Errorf("unexpected diff: %v\n", diff)
	}
}
