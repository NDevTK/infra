// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import "time"

// TODO: Explore spanner solution for storing
// new entities. See crbug/1510733
type CQSummaryCoverageData struct {
	Timestamp         time.Time `datastore:"timestamp"`
	Change            int64     `datastore:"change"`
	Patchset          int64     `datastore:"patchset"`
	IsUnitTest        bool      `datastore:"is_unit_test"`
	Path              string    `datastore:"path"`
	DataType          string    `datastore:"data_type"`
	FilesCovered      int64     `datastore:"files_covered"`
	TotalFilesChanged int64     `datastore:"total_files_changed"`
}
