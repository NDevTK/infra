// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package provider

import (
	"google.golang.org/protobuf/types/known/emptypb"
)

type stringSet = map[string]*emptypb.Empty

// newStringSet creates a string Set.
// We cannot use luci/common/data/stringset/stringset.go here
// since the types mismatch.
func newStringSet(s []string) stringSet {
	m := make(stringSet, len(s))
	for _, k := range s {
		m[k] = &emptypb.Empty{}
	}
	return m
}
