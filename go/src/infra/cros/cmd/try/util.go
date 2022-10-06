// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"fmt"
	"strings"
)

// prependString returns an array with an element at the beginning.
func prependString(newElem string, arr []string) []string {
	return append([]string{newElem}, arr...)
}

// separateBucketFromBuilder takes a full builder name (like chromeos/release/release-main-orchestrator),
// and separates it into a bucket (chromeos/release) and a builder (release-main-orchestrator).
func separateBucketFromBuilder(fullBuilderName string) (bucket string, builder string, err error) {
	parts := strings.Split(fullBuilderName, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("builder %s has %d slash-delimited parts; expect 3", fullBuilderName, len(parts))
	}
	bucket = strings.Join(parts[:2], "/")
	builder = parts[2]
	return bucket, builder, nil
}

// interfaceSlicetoStr converts a slice of interface{}s to a slice of strings.
func interfaceSliceToStr(s []interface{}) []string {
	ret := make([]string, len(s))
	for i := range s {
		ret[i] = s[i].(string)
	}
	return ret
}
