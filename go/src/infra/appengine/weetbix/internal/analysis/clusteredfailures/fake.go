// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package clusteredfailures

import "context"

// FakeClient represents a fake implementation of the clustered failures
// exporter, for testing.
type FakeClient struct {
	Insertions []*Entry
}

// NewFakeClient creates a new FakeClient for exporting clustered failures.
func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

// Insert inserts the given rows in BigQuery.
func (fc *FakeClient) Insert(ctx context.Context, rows []*Entry) error {
	fc.Insertions = append(fc.Insertions, rows...)
	return nil
}
