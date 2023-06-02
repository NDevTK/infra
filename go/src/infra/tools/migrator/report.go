// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package migrator

import (
	"fmt"
	"sort"
	"strings"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/config"
)

// ReportID is a (Checkout, Project, ConfigFile) tuple and identifies the object
// which generated the report.
type ReportID struct {
	Checkout   string
	Project    string
	ConfigFile string
}

// ConfigSet returns the luci-config "config.Set" for this report.
//
// e.g. "projects/${Project}"
func (r ReportID) ConfigSet() config.Set {
	cfg, err := config.ProjectSet(r.Project)
	if err != nil {
		panic(err)
	}
	return cfg
}

func (r ReportID) String() string {
	chunks := make([]string, 1, 3)
	chunks[0] = r.Checkout
	if r.Project != "" {
		chunks = append(chunks, r.Project)
	}
	if r.ConfigFile != "" {
		chunks = append(chunks, r.ConfigFile)
	}
	return strings.Join(chunks, "|")
}

// Report stores a single tagged problem (and metadata).
type Report struct {
	ReportID

	Tag     string
	Problem string

	// If true, indicates that this report can be fixed by ApplyFix.
	Actionable bool

	Metadata map[string]stringset.Set
}

// Clone returns a deep copy of this Report.
func (r *Report) Clone() *Report {
	ret := *r
	if len(ret.Metadata) > 0 {
		meta := make(map[string]stringset.Set, len(r.Metadata))
		for k, vals := range r.Metadata {
			meta[k] = vals.Dup()
		}
		ret.Metadata = meta
	}
	return &ret
}

// ToCSVRow returns a CSV row:
//
//	Checkout, Project, ConfigFile, Tag, Problem, Actionable, Metadata*
//
// Where Metadata* is one key:value entry per value in Metadata.
func (r *Report) ToCSVRow() []string {
	ret := []string{r.Checkout, r.Project, r.ConfigFile, r.Tag, r.Problem, fmt.Sprintf("%t", r.Actionable)}
	if len(r.Metadata) > 0 {
		keys := make([]string, len(r.Metadata))
		for key := range r.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, value := range r.Metadata[key].ToSortedSlice() {
				ret = append(ret, fmt.Sprintf("%s:%s", key, value))
			}
		}
	}
	return ret
}

// NewReportFromCSVRow creates a new Report from a CSVRow written with ToCSVRow.
func NewReportFromCSVRow(row []string) (ret *Report, err error) {
	shift := func() (string, bool) {
		if len(row) == 0 {
			return "", false
		}
		ret := row[0]
		row = row[1:]
		return ret, true
	}

	ret = &Report{}
	var ok bool
	if ret.Checkout, ok = shift(); !ok || ret.Checkout == "" {
		err = errors.New("Checkout field required")
		return
	}
	if ret.Project, ok = shift(); !ok || ret.Project == "" {
		err = errors.New("Project field required")
		return
	}
	if ret.ConfigFile, ok = shift(); !ok {
		err = errors.New("ConfigFile field required (may be empty)")
		return
	}
	if ret.Tag, ok = shift(); !ok || ret.Tag == "" {
		err = errors.New("Tag field required")
		return
	}
	if ret.Problem, ok = shift(); !ok {
		err = errors.New("Problem field required (may be empty)")
		return
	}

	actionable := ""
	if actionable, ok = shift(); !ok {
		err = errors.New("Actionable field required")
		return
	}
	ret.Actionable = actionable == "true"

	for i, mdata := range row {
		toks := strings.SplitN(mdata, ":", 2)
		if len(toks) != 2 {
			err = errors.Reason("Malformed metadata item %d, expected colon: %q",
				i, mdata).Err()
			return
		}
		MetadataOption(toks[0], toks[1])(ret)
	}

	return
}

// ReportOption allows attaching additional optional data to reports.
type ReportOption func(*Report)

// MetadataOption returns a ReportOption which allows attaching a string-string
// multimap of metadatadata to a Report.
func MetadataOption(key string, values ...string) ReportOption {
	return func(r *Report) {
		if r.Metadata == nil {
			r.Metadata = map[string]stringset.Set{}
		}
		set, ok := r.Metadata[key]
		if !ok {
			r.Metadata[key] = stringset.NewFromSlice(values...)
			return
		}
		set.AddAll(values)
	}
}

// NonActionable is-a ReportOption which indicates that this Report cannot be
// fixed by ApplyFix. If there are no Actionable Reports for a given project in
// FindProblems, the checkout and ApplyFix phase will be skipped.
func NonActionable(r *Report) {
	r.Actionable = false
}
