// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"

	"cloud.google.com/go/bigquery"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

type savedValue struct {
	insertID string
	row      map[string]bigquery.Value
}

func value(insertID, jsonVal string) savedValue {
	v := savedValue{insertID: insertID}
	So(json.Unmarshal([]byte(jsonVal), &v.row), ShouldBeNil)
	return v
}

func doReadInput(data string, jsonList bool) ([]savedValue, error) {
	savers, err := readInput(strings.NewReader(data), "seed", jsonList)
	if err != nil {
		return nil, err
	}
	out := make([]savedValue, len(savers))
	for i, saver := range savers {
		if out[i].row, out[i].insertID, err = saver.Save(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func TestReadInput(t *testing.T) {
	t.Parallel()

	Convey("Empty", t, func() {
		vals, err := doReadInput("", false)
		So(err, ShouldBeNil)
		So(vals, ShouldHaveLength, 0)
	})

	Convey("Whitespace only", t, func() {
		vals, err := doReadInput("\n  \n\n  \n  ", false)
		So(err, ShouldBeNil)
		So(vals, ShouldHaveLength, 0)
	})

	Convey("One line", t, func() {
		vals, err := doReadInput(`{"k": "v"}`, false)
		So(err, ShouldBeNil)
		So(vals, ShouldResemble, []savedValue{
			value("seed:0", `{"k": "v"}`),
		})
	})

	Convey("A bunch of lines (with spaces)", t, func() {
		vals, err := doReadInput(`
			{"k": "v1"}

			{"k": "v2"}
			{"k": "v3"}

		`, false)
		So(err, ShouldBeNil)
		So(vals, ShouldResemble, []savedValue{
			value("seed:0", `{"k": "v1"}`),
			value("seed:1", `{"k": "v2"}`),
			value("seed:2", `{"k": "v3"}`),
		})
	})

	Convey("Broken line", t, func() {
		_, err := doReadInput(`
			{"k": "v1"}

			{"k": "v2
			{"k": "v2"}
		`, false)
		So(err, ShouldErrLike, `bad input line 4: bad JSON - unexpected end of JSON input`)
	})

	Convey("JSON List", t, func() {
		out, err := doReadInput(`[
			{"k": "v1"},
			{"k": "v2"},
			{"k": "v2"}
		]`, true)
		So(err, ShouldBeNil)
		So(out, ShouldResemble, []savedValue{
			value("seed:0", `{"k": "v1"}`),
			value("seed:1", `{"k": "v2"}`),
			value("seed:2", `{"k": "v2"}`),
		})
	})

	Convey("Huge line", t, func() {
		// Note: this breaks bufio.Scanner with "token too long" error.
		huge := fmt.Sprintf(`{"k": %q}`, strings.Repeat("x", 100000))
		vals, err := doReadInput(huge, false)
		So(err, ShouldBeNil)
		So(vals, ShouldResemble, []savedValue{
			value("seed:0", huge),
		})
	})
}

type fakeInserter struct {
	calls            [][]*tableRow
	mu               sync.Mutex
	failingInsertIDs []string
}

func (i *fakeInserter) Put(_ context.Context, src interface{}) error {
	rows := src.([]*tableRow)
	i.mu.Lock()
	i.calls = append(i.calls, rows)
	i.mu.Unlock()

	var multiErr bigquery.PutMultiError
	for _, row := range rows {
		for _, id := range i.failingInsertIDs {
			if id == row.insertID {
				multiErr = append(multiErr, bigquery.RowInsertionError{InsertID: id})
			}
		}
	}
	if len(multiErr) > 0 {
		return multiErr
	}
	return nil
}

func TestDoInsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	opts := uploadOpts{batchSize: 3}

	Convey("One batch", t, func() {
		rows := []*tableRow{
			{insertID: "1"},
			{insertID: "2"},
			{insertID: "3"},
		}
		var ins fakeInserter
		err := doInsert(ctx, io.Discard, &opts, &ins, rows)
		So(err, ShouldBeNil)
		So(ins.calls, shouldResembleUnsorted, [][]*tableRow{rows})
	})

	Convey("Multiple batches", t, func() {
		rows := []*tableRow{
			{insertID: "1"},
			{insertID: "2"},
			{insertID: "3"},
			{insertID: "4"},
			{insertID: "5"},
			{insertID: "6"},
			{insertID: "7"},
		}
		var ins fakeInserter
		err := doInsert(ctx, io.Discard, &opts, &ins, rows)
		So(err, ShouldBeNil)
		So(ins.calls, shouldResembleUnsorted, [][]*tableRow{
			rows[0:3],
			rows[3:6],
			rows[6:7],
		})
	})

	Convey("Multiple batches with failures", t, func() {
		rows := []*tableRow{
			{insertID: "1"},
			{insertID: "2"},
			{insertID: "3"},
			{insertID: "4"},
			{insertID: "5"},
			{insertID: "6"},
			{insertID: "7"},
		}
		ins := fakeInserter{
			failingInsertIDs: []string{"1", "5"},
		}
		err := doInsert(ctx, io.Discard, &opts, &ins, rows)
		multiErr := err.(bigquery.PutMultiError)
		// Sort in case insertions happened out-of-order.
		sort.Slice(multiErr, func(i, j int) bool {
			return multiErr[i].InsertID < multiErr[j].InsertID
		})
		So(err, ShouldResemble, bigquery.PutMultiError{
			bigquery.RowInsertionError{InsertID: "1"},
			bigquery.RowInsertionError{InsertID: "5"},
		})
		So(ins.calls, shouldResembleUnsorted, [][]*tableRow{
			rows[0:3],
			rows[3:6],
			rows[6:7],
		})
	})
}

// shouldResembleUnsorted is like ShouldResemble, but operates on [][]*tableRow
// and ignores the ordering of the `actual` slice, since the ordering doesn't
// matter as long as all rows are uploaded. The `expected` slice must already be
// sorted by insert ID.
func shouldResembleUnsorted(actual interface{}, expected ...interface{}) string {
	actualRows := actual.([][]*tableRow)
	sort.Slice(actualRows, func(i, j int) bool {
		return actualRows[i][0].insertID < actualRows[j][0].insertID
	})
	return ShouldResemble(actualRows, expected...)
}
