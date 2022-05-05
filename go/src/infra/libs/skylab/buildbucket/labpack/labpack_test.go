// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package labpack

import (
	"sort"
	"testing"
)

// TestAsMap tests structbuilder-compatibility.
//
// Make sure that we only have keys of a type that structbuilder understands.
//
// We will be more conservative than structbuilder and reject everything that isn't a bool or a string.
//
// Keep the function deterministic by sorting the keys before we check for
// values that have an unsupported type.
func TestAsMap(t *testing.T) {
	t.Parallel()
	zero := Params{}
	zeroMap := zero.AsMap()

	// Keep the function deterministic by sorting the keys before we check for
	// values that have an unsupported type.
	keys := make([]string, 0, len(zeroMap))
	for k := range zeroMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := zeroMap[k]
		switch v := v.(type) {
		case bool, string:
			// do nothing
		default:
			t.Errorf("key %q has value %v with unsupported type %T", k, v, v)
		}
	}
}
