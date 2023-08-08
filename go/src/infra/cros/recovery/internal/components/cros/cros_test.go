// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"testing"
	"time"
)

type uptimeValue struct {
	valid bool
	value string
}

var uptimes = map[string]uptimeValue{
	"83.38 52.68":          {true, "1m23.38s"},
	"83.38":                {false, ""},
	"83":                   {false, ""},
	"0":                    {false, ""},
	"683503.88 1003324.85": {true, "189h51m43.88s"},
	"683503 1003324":       {true, "189h51m43s"},
	"0 0":                  {true, "0s"},
	"1 1":                  {true, "1s"},
	"1 1.0":                {true, "1s"},
	"1.0 1.0":              {true, "1s"},
	"1. 1.":                {true, "1s"},
	"0.0 0.0":              {true, "0s"},
	"0 0.0":                {true, "0s"},
	"0. 0.":                {true, "0s"},
	"10.0 0.0 1.3 5.5":     {false, ""},
	"0 0 0":                {false, ""},
	"  83.38 52.68":        {true, "1m23.38s"},
	"83.38 52.68  ":        {true, "1m23.38s"},
	"  83.38 52.68  ":      {true, "1m23.38s"},
}

func TestProcessUptime(t *testing.T) {
	for k, v := range uptimes {
		dur, err := ProcessUptime(k)
		if err == nil {
			if v.valid {
				if d, err := time.ParseDuration(v.value); err == nil {
					if d != *dur {
						t.Errorf("uptime value %s, expected %q, actual %q", k, d, dur)
					}
				}
				// if err is not nil, this is an erroneous test case
				// and we skip it
			} else {
				t.Errorf("uptime value %s, this is not valid uptime, this should have errored-out", k)
			}
		} else {
			if v.valid {
				t.Errorf("uptime value %s is valid, but error occured in parsing", k)
			}
		}
	}
}
