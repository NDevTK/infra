// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"testing"
)

var testIDsParamData = []struct {
	bbIDs        []int64
	wantIDsParam string
}{
	{
		bbIDs:        []int64{4, 9, 2, 6, 0},
		wantIDsParam: "ids=4,9,2,6,0",
	},
	{
		bbIDs:        []int64{1},
		wantIDsParam: "ids=1",
	},
	{
		bbIDs:        []int64{},
		wantIDsParam: "ids=",
	},
	{
		bbIDs:        nil,
		wantIDsParam: "ids=",
	},
}

func TestIDsParam(t *testing.T) {
	t.Parallel()
	for _, tt := range testIDsParamData {
		tt := tt
		t.Run(fmt.Sprintf("(%v)", tt.bbIDs), func(t *testing.T) {
			t.Parallel()
			gotIDsParam := idsParam(tt.bbIDs)
			if gotIDsParam != tt.wantIDsParam {
				t.Errorf("got %s, want %s", gotIDsParam, tt.wantIDsParam)
			}
		})
	}
}
