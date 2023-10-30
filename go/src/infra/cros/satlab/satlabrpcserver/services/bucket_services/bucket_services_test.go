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

type mockBucketClient struct {
	mock.Mock
}

func (m *mockBucketClient) GetAttrs(ctx context.Context) (*storage.BucketAttrs, error) {
	args := m.Called(ctx)
	return args.Get(0).(*storage.BucketAttrs), args.Error(1)
}

// QueryObjects query objects from the bucket
func (m *mockBucketClient) QueryObjects(ctx context.Context, q *storage.Query) iObjectIterator {
	args := m.Called(ctx, q)
	return args.Get(0).(iObjectIterator)
}

// ReadObject read the object content by the given name
func (m *mockBucketClient) ReadObject(ctx context.Context, name string) (io.ReadCloser, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// Close do clean up
func (m *mockBucketClient) Close() error {
	return nil
}

func Test_ListTestPlans(t *testing.T) {
	ctx := context.Background()

	expected := []string{
		"testplan1.json",
		"testplan2.json",
	}

	// Create a Mock `IBucketService`
	var mockBucketClient = new(mockBucketClient)
	mockBucketClient.
		On("QueryObjects", ctx, mock.Anything).
		Return(&fakeObjectIter{
			data: []string{
				"testplans/testplan1.json",
				"testplans/testplan2.json",
			},
			i: 0,
		}, nil)

	b := BucketConnector{
		client: mockBucketClient,
	}

	resp, err := b.ListTestplans(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
	}

	if diff := cmp.Diff(resp, expected); diff != "" {
		t.Errorf("unexpected diff: %v\n", diff)
	}
}
