// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	schedukepb "go.chromium.org/chromiumos/config/go/test/scheduling"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

const (
	// Higher integer value means lower priority.
	noAccountPriority      int64 = 10
	deviceNameDimensionKey       = "dut_name"
	poolDimensionKey             = "label-pool"
	quotaAccountTagKey           = "qs_account"
	// SchedukeTaskRequestKey is the key all Scheduke tasks are launched with.
	// Scheduke supports batch task creation, but we send individually for now, so
	// we use this key.
	SchedukeTaskRequestKey int64 = 1
	suiteSchedulerTagKey         = "analytics_name"
)

var (
	asapQSAccounts = []string{
		"pcq",
	}
	quotaAccountPriorities = map[string]int64{
		"bisector":             2,
		"bvt-sync":             3,
		"cq":                   1,
		"cts":                  5,
		"deputy":               3,
		"lacros":               3,
		"lacros_fyi":           5,
		"leases":               1,
		"legacypool-bvt":       4,
		"legacypool-suites":    5,
		"p0_cq_unmanaged":      1,
		"pcq":                  1,
		"pfq":                  1,
		"postsubmit":           2,
		"pupr":                 2,
		"release_direct_sched": 2,
		"release_high_prio":    2,
		"release_low_prio":     3,
		"release_med_prio":     4,
		"release_p0":           2,
		"toolchain":            3,
		"unmanaged_p0":         2,
		"unmanaged_p1":         3,
		"unmanaged_p2":         4,
		"unmanaged_p3":         5,
		"unmanaged_p4":         10,
		"wificell":             3,
	}
)

// ScheduleBuildReqToSchedukeReq converts a Buildbucket ScheduleBuildRequest to
// a Scheudke request with the given event time.
func ScheduleBuildReqToSchedukeReq(bbReq *buildbucketpb.ScheduleBuildRequest) (*schedukepb.KeyedTaskRequestEvents, error) {
	bbReqBytes := []byte(protojson.Format(bbReq))
	compressedReqJSON, err := compressAndEncodeBBReq(bbReqBytes)
	if err != nil {
		return nil, fmt.Errorf("error compressing and encoding ScheduleBuildRequest %v: %w", bbReq, err)
	}
	cftReq, ok := bbReq.GetProperties().GetFields()["cft_test_request"]
	if !ok {
		return nil, fmt.Errorf("no cft test request found on ScheduleBuildRequest %v", bbReq)
	}
	deadlineStruct, ok := cftReq.GetStructValue().GetFields()["deadline"]
	if !ok {
		return nil, fmt.Errorf("no deadline found on ScheduleBuildRequest %v", bbReq)
	}
	parentBBIDStr := cftReq.GetStructValue().GetFields()["parentBuildId"].GetStringValue()
	var parentBBID int64
	// Fail softly if parentBuildId field is not set on the request, as Scheduke
	// only uses this for metadata/logging.
	if parentBBIDStr != "" {
		parentBBID, err = strconv.ParseInt(parentBBIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid parent BBID found on ScheduleBuildRequest %v", bbReq)
		}
	}
	deadline, err := timeFromTimestampPBString(deadlineStruct.GetStringValue())
	if err != nil {
		return nil, fmt.Errorf("error parsing deadline for ScheduleBuildRequest %v: %w", bbReq, err)
	}
	tags := bbReq.GetTags()
	qsAccount := qsAccount(tags)
	periodic := periodic(tags)
	asap := asap(qsAccount, periodic)
	dims, deviceName, pool := dimensionsDeviceNameAndPool(bbReq.GetDimensions())

	schedukeTask := &schedukepb.TaskRequestEvent{
		EventTime:                time.Now().UnixMicro(),
		Deadline:                 deadline.UnixMicro(),
		Periodic:                 periodic,
		Priority:                 priority(tags),
		RequestedDimensions:      dims,
		RealExecutionMinutes:     0, // Unneeded outside of shadow mode.
		MaxExecutionMinutes:      30,
		QsAccount:                qsAccount,
		Pool:                     pool,
		Bbid:                     parentBBID,
		Asap:                     asap,
		ScheduleBuildRequestJson: compressedReqJSON,
		DeviceName:               deviceName,
	}

	return &schedukepb.KeyedTaskRequestEvents{
		Events: map[int64]*schedukepb.TaskRequestEvent{
			SchedukeTaskRequestKey: schedukeTask,
		},
	}, nil
}

// priority derives the approximate Scheduke priority from the given build's
// Quota Scheduler account, returning a high value (i.e. low priority) if no
// account was found.
func priority(tags []*buildbucketpb.StringPair) int64 {
	account := qsAccount(tags)
	priority, ok := quotaAccountPriorities[account]
	if !ok {
		priority = noAccountPriority
	}
	return priority
}

// qsAccount looks for the Quota Scheduler account on the given build's tags,
// returning an empty string if no account was found.
func qsAccount(tags []*buildbucketpb.StringPair) string {
	for _, t := range tags {
		if t.GetKey() == quotaAccountTagKey {
			return t.GetValue()
		}
	}
	return ""
}

// periodic checks if the given build is periodic by seeing if it has
// a specific tag only included on builds from Suite Scheduler.
func periodic(tags []*buildbucketpb.StringPair) bool {
	for _, t := range tags {
		if t.GetKey() == suiteSchedulerTagKey {
			return true
		}
	}
	return false
}

// timeFromTimestampPBString converts a timestamp PB string to time.Time.
func timeFromTimestampPBString(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// asap returns a bool indicating whether a build with the given Quota Scheduler
// account and periodicity should be scheduled with the "asap" flag.
func asap(qsAccount string, periodic bool) bool {
	if periodic {
		return false
	}
	for _, asapAcct := range asapQSAccounts {
		if qsAccount == asapAcct {
			return true
		}
	}
	return false
}

// dimensionsDeviceNameAndPool converts the given Buildbucket RequestedDimensions to
// Scheduke SwarmingDimensions, and returns the pool dimension separately.
func dimensionsDeviceNameAndPool(dims []*buildbucketpb.RequestedDimension) (schedukeDims *schedukepb.SwarmingDimensions, deviceName, pool string) {
	dimsMap := make(map[string]*schedukepb.DimValues)

	for _, dim := range dims {
		dimKey := dim.GetKey()
		theseVals := strings.Split(dim.GetValue(), "|")
		if dimsMap[dimKey] == nil {
			dimsMap[dimKey] = &schedukepb.DimValues{Values: nil}
		}
		dimsMap[dimKey].Values = append(dimsMap[dimKey].Values, theseVals...)
		if dimKey == poolDimensionKey {
			pool = dim.GetValue()
		}
		if dimKey == deviceNameDimensionKey {
			deviceName = dim.GetValue()
		}
	}
	schedukeDims = &schedukepb.SwarmingDimensions{DimsMap: dimsMap}
	return
}

// compressAndEncodeBBReq compresses the given bytes using zlib and encodes it
// through base64 codecs.
func compressAndEncodeBBReq(src []byte) (string, error) {
	var in bytes.Buffer
	w, err := zlib.NewWriterLevel(&in, zlib.DefaultCompression)
	if err != nil {
		return "", err
	}
	_, err = w.Write(src)
	w.Close()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(in.Bytes()), nil
}
