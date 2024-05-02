// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"fmt"
	"testing"
)

var testSchedukeParamsData = []struct {
	taskStateIDs []int64
	users        []string
	deviceNames  []string
	wantParam    string
}{
	{
		taskStateIDs: []int64{4, 9, 2, 6, 0},
		users:        []string{"a", "b", "c"},
		deviceNames:  []string{"d", "f", "g"},
		wantParam:    "ids=4,9,2,6,0&users=a,b,c&device_names=d,f,g",
	},
	{
		taskStateIDs: []int64{},
		users:        []string{"a", "b", "e"},
		deviceNames:  []string{"d", "f", "g"},
		wantParam:    "users=a,b,e&device_names=d,f,g",
	},
	{
		taskStateIDs: []int64{4, 9, 2, 6, 0},
		users:        []string{},
		deviceNames:  []string{"e", "f", "g"},
		wantParam:    "ids=4,9,2,6,0&device_names=e,f,g",
	},
	{
		taskStateIDs: []int64{4, 9, 2, 6, 0},
		users:        []string{"a", "b", "c"},
		deviceNames:  nil,
		wantParam:    "ids=4,9,2,6,0&users=a,b,c",
	},
	{
		taskStateIDs: []int64{4, 9, 2, 6, 0},
		users:        nil,
		deviceNames:  nil,
		wantParam:    "ids=4,9,2,6,0",
	},
	{
		taskStateIDs: nil,
		users:        []string{"a", "b", "c"},
		deviceNames:  nil,
		wantParam:    "users=a,b,c",
	},
	{
		taskStateIDs: nil,
		users:        nil,
		deviceNames:  []string{"d", "f", "g"},
		wantParam:    "device_names=d,f,g",
	},
	{
		taskStateIDs: nil,
		users:        nil,
		deviceNames:  nil,
		wantParam:    "",
	},
}

func TestSchedukeParams(t *testing.T) {
	t.Parallel()
	for _, tt := range testSchedukeParamsData {
		tt := tt
		t.Run(fmt.Sprintf("(%v/%v/%v)", tt.taskStateIDs, tt.users, tt.deviceNames), func(t *testing.T) {
			t.Parallel()
			gotParam := schedukeParams(tt.taskStateIDs, tt.users, tt.deviceNames)
			if gotParam != tt.wantParam {
				t.Errorf("got %s, want %s", gotParam, tt.wantParam)
			}
		})
	}
}
