// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"reflect"
	"testing"
)

func Test_ChromeOSReleaseVersion_String(t *testing.T) {
	tests := []struct {
		name string
		v    ChromeOSReleaseVersion
		want string
	}{
		{
			"empty list",
			[]int{},
			"",
		},
		{
			"nil list",
			nil,
			"",
		},
		{
			"single",
			[]int{123},
			"123",
		},
		{
			"normal",
			[]int{123, 456, 789},
			"123.456.789",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsChromeOSReleaseVersionLessThan(t *testing.T) {
	type args struct {
		a ChromeOSReleaseVersion
		b ChromeOSReleaseVersion
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"same value - nil",
			args{
				nil, nil,
			},
			false,
		},
		{
			"same value - normal",
			args{
				[]int{123, 456, 789},
				[]int{123, 456, 789},
			},
			false,
		},
		{
			"first is less",
			args{
				[]int{122, 456, 789},
				[]int{123, 456, 789},
			},
			true,
		},
		{
			"second is less",
			args{
				[]int{123, 454, 789},
				[]int{123, 456, 789},
			},
			true,
		},
		{
			"second is greater",
			args{
				[]int{123, 457, 789},
				[]int{123, 456, 789},
			},
			false,
		},
		{
			"b has more parts",
			args{
				[]int{123, 456},
				[]int{123, 456, 789},
			},
			true,
		},
		{
			"b has less parts",
			args{
				[]int{123, 456},
				[]int{123},
			},
			false,
		},
		{
			"second is less, b has more parts",
			args{
				[]int{123, 454},
				[]int{123, 456, 789},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsChromeOSReleaseVersionLessThan(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsChromeOSReleaseVersionLessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ParseChromeOSReleaseVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    ChromeOSReleaseVersion
		wantErr bool
	}{
		{
			"empty string",
			args{
				"",
			},
			nil,
			true,
		},
		{
			"non-integer string",
			args{
				"abc",
			},
			nil,
			true,
		},
		{
			"non-integer string in one part",
			args{
				"123.abc.456",
			},
			nil,
			true,
		},
		{
			"single part",
			args{
				"123",
			},
			[]int{123},
			false,
		},
		{
			"3 parts (normal)",
			args{
				"123.456.789",
			},
			[]int{123, 456, 789},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChromeOSReleaseVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseChromeOSReleaseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseChromeOSReleaseVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReleaseVersionFromBuilderPath(t *testing.T) {
	type args struct {
		releaseBuilderPath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty string",
			args{
				"",
			},
			"",
			true,
		},
		{
			"no version in path",
			args{
				"board-release/R90-",
			},
			"",
			true,
		},
		{
			"bad version",
			args{
				"board-release/R90-abc",
			},
			"",
			true,
		},
		{
			"normal path",
			args{
				"board-release/R90-13816.47.0",
			},
			"13816.47.0",
			false,
		},
		{
			"just a release segment",
			args{
				"R90-13816.47.0",
			},
			"13816.47.0",
			false,
		},
		{
			"just a version",
			args{
				"13816.47.0",
			},
			"13816.47.0",
			false,
		},
		{
			"partial version",
			args{
				"13816",
			},
			"13816",
			false,
		},
		{
			"long version",
			args{
				"13816.123.456.7.8",
			},
			"13816.123.456.7.8",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReleaseVersionFromBuilderPath(tt.args.releaseBuilderPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReleaseVersionFromBuilderPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseReleaseVersionFromBuilderPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
