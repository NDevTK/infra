// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	schedukepb "go.chromium.org/chromiumos/config/go/test/scheduling"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

var testSchedukePriorityData = []struct {
	build        *buildbucketpb.Build
	wantPriority int64
}{
	{
		build: &buildbucketpb.Build{
			Tags: []*buildbucketpb.StringPair{
				{Key: "qs_account", Value: "release_low_prio"},
			}},
		wantPriority: 3,
	},
	{
		build:        &buildbucketpb.Build{},
		wantPriority: 10,
	},
}

func TestSchedukePriority(t *testing.T) {
	t.Parallel()
	for _, tt := range testSchedukePriorityData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.build.GetTags()), func(t *testing.T) {
			t.Parallel()
			gotPriority := priority(tt.build.GetTags())
			if gotPriority != tt.wantPriority {
				t.Errorf("got %d, want %d", gotPriority, tt.wantPriority)
			}
		})
	}
}

var testQuotaAccountData = []struct {
	build       *buildbucketpb.Build
	wantAccount string
}{
	{
		build: &buildbucketpb.Build{
			Tags: []*buildbucketpb.StringPair{
				{Key: "foo", Value: "bar"},
				{Key: "qs_account", Value: "the account"},
			}},
		wantAccount: "the account",
	},
	{
		build: &buildbucketpb.Build{
			Tags: []*buildbucketpb.StringPair{
				// Misspelled key; shouldn't return a value.
				{Key: "qs-account-with-typo", Value: "the account"},
			}},
		wantAccount: "",
	},
	{
		build:       &buildbucketpb.Build{},
		wantAccount: "",
	},
}

func TestQuotaAccount(t *testing.T) {
	t.Parallel()
	for _, tt := range testQuotaAccountData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.build.GetTags()), func(t *testing.T) {
			t.Parallel()
			gotAccount := qsAccount(tt.build.GetTags())
			if gotAccount != tt.wantAccount {
				t.Errorf("got %v, want %v", gotAccount, tt.wantAccount)
			}
		})
	}
}

var testPeriodicData = []struct {
	build           *buildbucketpb.Build
	wantPeriodicity bool
}{
	{
		build:           &buildbucketpb.Build{},
		wantPeriodicity: false,
	},
	{
		build: &buildbucketpb.Build{
			Tags: []*buildbucketpb.StringPair{
				{Key: "foo", Value: "bar"},
			},
		},
		wantPeriodicity: false,
	},
	{
		build: &buildbucketpb.Build{
			Tags: []*buildbucketpb.StringPair{
				{Key: "foo", Value: "bar"},
				{Key: "analytics_name", Value: "baz"},
			},
		},
		wantPeriodicity: true,
	},
}

func TestPeriodic(t *testing.T) {
	t.Parallel()
	for _, tt := range testPeriodicData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.build.GetTags()), func(t *testing.T) {
			t.Parallel()
			gotPeriodicity := periodic(tt.build.GetTags())
			if gotPeriodicity != tt.wantPeriodicity {
				t.Errorf("got %v, want %v", gotPeriodicity, tt.wantPeriodicity)
			}
		})
	}
}

func TestTimeFromTimestampPBString(t *testing.T) {
	t.Parallel()
	timestampPBString := "2023-01-23T14:23:05.808538180Z"
	wantTime := time.Date(2023, 1, 23, 14, 23, 5, 808538180, time.UTC)
	gotTime, err := timeFromTimestampPBString(timestampPBString)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if gotTime != wantTime {
		t.Errorf("got %v, want %v", gotTime, wantTime)
	}
}

var testAsapData = []struct {
	qsAccount string
	periodic  bool
	wantAsap  bool
}{
	{
		qsAccount: "pcq",
		periodic:  false,
		wantAsap:  true,
	},
	{
		qsAccount: "pcq",
		periodic:  true,
		wantAsap:  false,
	},
	{
		qsAccount: "not pcq",
		periodic:  true,
		wantAsap:  false,
	},
	{
		qsAccount: "not pcq",
		periodic:  false,
		wantAsap:  false,
	},
}

func TestAsap(t *testing.T) {
	t.Parallel()
	for _, tt := range testAsapData {
		tt := tt
		t.Run(fmt.Sprintf("(%s/%v)", tt.qsAccount, tt.periodic), func(t *testing.T) {
			t.Parallel()
			gotAsap := asap(tt.qsAccount, tt.periodic)
			if gotAsap != tt.wantAsap {
				t.Errorf("got %v, want %v", gotAsap, tt.wantAsap)
			}
		})
	}
}

var testDimensionsAndPoolData = []struct {
	bbDims           []*buildbucketpb.RequestedDimension
	wantSchedukeDims *schedukepb.SwarmingDimensions
	wantPool         string
	wantDev          bool
}{
	{
		bbDims: []*buildbucketpb.RequestedDimension{},
		wantSchedukeDims: &schedukepb.SwarmingDimensions{
			DimsMap: map[string]*schedukepb.DimValues{},
		},
		wantPool: "",
		wantDev:  false,
	},
	{
		bbDims: []*buildbucketpb.RequestedDimension{
			{
				Key:   "foo",
				Value: "val",
			},
			{
				Key:   "bar",
				Value: "val1|val2",
			},
		},
		wantSchedukeDims: &schedukepb.SwarmingDimensions{
			DimsMap: map[string]*schedukepb.DimValues{
				"foo": {Values: []string{"val"}},
				"bar": {Values: []string{"val1", "val2"}},
			},
		},
		wantPool: "",
		wantDev:  false,
	},
	{
		bbDims: []*buildbucketpb.RequestedDimension{
			{
				Key:   "foo",
				Value: "val",
			},
			{
				Key:   "bar",
				Value: "val1|val2",
			},
			{
				Key:   "bar",
				Value: "val3",
			},
			{
				Key:   "bar",
				Value: "val4|val5",
			},
			{
				Key:   "label-pool",
				Value: "pool1|pool2",
			},
		},
		wantSchedukeDims: &schedukepb.SwarmingDimensions{
			DimsMap: map[string]*schedukepb.DimValues{
				"foo":        {Values: []string{"val"}},
				"bar":        {Values: []string{"val1", "val2", "val3", "val4", "val5"}},
				"label-pool": {Values: []string{"pool1", "pool2"}},
			},
		},
		wantPool: "pool1|pool2",
		wantDev:  false,
	},
	{
		bbDims: []*buildbucketpb.RequestedDimension{
			{
				Key:   "foo",
				Value: "val",
			},
			{
				Key:   "bar",
				Value: "val1|val2",
			},
			{
				Key:   "label-pool",
				Value: "baz pool",
			},
		},
		wantSchedukeDims: &schedukepb.SwarmingDimensions{
			DimsMap: map[string]*schedukepb.DimValues{
				"foo":        {Values: []string{"val"}},
				"bar":        {Values: []string{"val1", "val2"}},
				"label-pool": {Values: []string{"baz pool"}},
			},
		},
		wantPool: "baz pool",
		wantDev:  false,
	},
	{
		bbDims: []*buildbucketpb.RequestedDimension{
			{
				Key:   "foo",
				Value: "val",
			},
			{
				Key:   "bar",
				Value: "val1|val2",
			},
			{
				Key:   "label-pool",
				Value: "schedukeTest",
			},
		},
		wantSchedukeDims: &schedukepb.SwarmingDimensions{
			DimsMap: map[string]*schedukepb.DimValues{
				"foo":        {Values: []string{"val"}},
				"bar":        {Values: []string{"val1", "val2"}},
				"label-pool": {Values: []string{"schedukeTest"}},
			},
		},
		wantPool: "schedukeTest",
		wantDev:  true,
	},
}

func TestDimensionsAndPool(t *testing.T) {
	t.Parallel()
	cmpOpts := cmpopts.IgnoreUnexported(
		schedukepb.DimValues{},
		schedukepb.SwarmingDimensions{})
	for _, tt := range testDimensionsAndPoolData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.bbDims), func(t *testing.T) {
			t.Parallel()
			gotSchedukeDims, gotPool, gotDev := dimensionsAndPool(tt.bbDims)
			if diff := cmp.Diff(gotSchedukeDims, tt.wantSchedukeDims, cmpOpts); diff != "" {
				t.Errorf("unexpected diff (%s)", diff)
			}
			if gotPool != tt.wantPool {
				t.Errorf("got %v, want %v", gotPool, tt.wantPool)
			}
			if gotDev != tt.wantDev {
				t.Errorf("got %v, want %v", gotDev, tt.wantDev)
			}
		})
	}
}

var testCompressAndEncodeBBReqData = []struct {
	src                           []byte
	wantCompressedEncodedBuildStr string
}{
	{
		src: []byte(`{
	RequestId: "foo",
	Priority:  20
}`),
		wantCompressedEncodedBuildStr: "eJyq5uIMSi0sTS0u8UyxUlBKy89X0uHiDCjKzC/KLKm0UlAwMuCqBQQAAP//2HYLCw==",
	},
	{
		src: []byte(`{
	RequestId: "bar",
	Priority:  20
}`),
		wantCompressedEncodedBuildStr: "eJyq5uIMSi0sTS0u8UyxUlBKSixS0uHiDCjKzC/KLKm0UlAwMuCqBQQAAP//1zQK/A==",
	},
}

func TestCompressAndEncodeBBReq(t *testing.T) {
	t.Parallel()
	for _, tt := range testCompressAndEncodeBBReqData {
		tt := tt
		t.Run(fmt.Sprintf("(%s)", tt.src), func(t *testing.T) {
			t.Parallel()
			gotBuildStr, err := compressAndEncodeBBReq(tt.src)
			if err != nil {
				t.Fatalf("error found: %s", err)
			}
			if diff := cmp.Diff(gotBuildStr, tt.wantCompressedEncodedBuildStr); diff != "" {
				t.Errorf("unexpected diff (%s)", diff)
			}
		})
	}
}
