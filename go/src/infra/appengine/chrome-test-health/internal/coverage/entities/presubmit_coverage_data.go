// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package entities

import "time"

type Cov struct {
	Path         string `datastore:"path"`
	CoveredLines int64  `datastore:"covered_lines"`
	TotalLines   int64  `datastore:"total_lines"`
}

type PresubmitCoverageData struct {
	AbsolutePercentages        []Cov     `datastore:"absolute_percentages"`
	AbsolutePercentagesRts     []Cov     `datastore:"absolute_percentages_rts"`
	AbsolutePercentagesUnit    []Cov     `datastore:"absolute_percentages_unit"`
	AbsolutePercentagesUnitRts []Cov     `datastore:"absolute_percentages_unit_rts"`
	BasedOn                    int64     `datastore:"based_on"`
	Change                     int64     `datastore:"cl_patchset.change"`
	Patchset                   int64     `datastore:"cl_patchset.patchset"`
	Project                    string    `datastore:"cl_patchset.project"`
	ServerHost                 string    `datastore:"cl_patchset.server_host"`
	Data                       []byte    `datastore:"data"`
	DataRts                    []byte    `datastore:"data_rts"`
	DataUnit                   []byte    `datastore:"data_unit"`
	DataUnitRts                []byte    `datastore:"data_unit_rts"`
	IncrementalPercentages     []Cov     `datastore:"incremental_percentages"`
	IncrementalPercentagesUnit []Cov     `datastore:"incremental_percentages_unit"`
	InsertTimestamp            time.Time `datastore:"insert_timestamp"`
	TimesUpdated               int64     `datastore:"times_updated"`
	TimesUpdatedUnit           int64     `datastore:"times_updated_unit"`
	UpdateTimestamp            time.Time `datastore:"update_timestamp"`
}
