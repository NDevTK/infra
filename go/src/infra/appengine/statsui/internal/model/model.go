// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package contains internal model objects.
package model

// Period that defines what time period to look at for the dates given.
type Period string

const (
	// Day specifies a time period of a day.
	Day Period = "day"
	// Week specifies a time period of a week.
	Week Period = "week"
)

// DataSet is a map of dates to numbers.
type DataSet map[string]float32

// Metric is the data associated with the named metric for the given dates.
// If the data is grouped in some way, it will be in the Sections field instead.
type Metric struct {
	// Name of the metric.
	Name string
	// Map of dates to numbers of the metric.
	Data DataSet
	// Grouped data.
	Sections map[string]DataSet
}

// Section is a top-level grouping of metrics.
type Section struct {
	// Name of the section.
	Name string `json:"name"`
	// Metrics is the list of metrics for this section.
	Metrics []Metric `json:"metrics"`
}
