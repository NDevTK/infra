// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"testing"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func TestIsAndroid(t *testing.T) {
	d := &api.LegacyHW{}
	if IsAndroid(d) {
		t.Fatalf("should not be android")
	}
	d.Board = "foo"
	if IsAndroid(d) {
		t.Fatalf("should not be android")
	}
	d.Board = "pixel91"
	if !IsAndroid(d) || IsDevBoard(d) || IsCros(d) {
		t.Fatalf("Device should *only* be android")
	}
	d.Board = "pixel9ProFoldUltratiny"
	if !IsAndroid(d) || IsDevBoard(d) || IsCros(d) {
		t.Fatalf("Device should *only* be android")
	}
}

func TestIsDevBoard(t *testing.T) {
	d := &api.LegacyHW{}
	if IsDevBoard(d) {
		t.Fatalf("should not be DevBoard")
	}
	d.Board = "foo"
	if IsDevBoard(d) {
		t.Fatalf("should not be DevBoard")
	}
	d.Board = "some-devboard"
	if IsAndroid(d) || !IsDevBoard(d) || IsCros(d) {
		t.Fatalf("Device should *only* be DevBoard")
	}
	d.Board = "SOME-DEVBOARD"
	if IsAndroid(d) || !IsDevBoard(d) || IsCros(d) {
		t.Fatalf("Device should *only* be android")
	}
}

func TestIsCros(t *testing.T) {
	d := &api.LegacyHW{}
	if !IsCros(d) {
		t.Fatalf("Empty should default to Cros")
	}
	d.Board = "foo"
	if !IsCros(d) {
		t.Fatalf("Default should be cros")
	}
	d.Board = "brya"
	if IsAndroid(d) || IsDevBoard(d) || !IsCros(d) {
		t.Fatalf("Device should *only* be DevBoard")
	}
	d.Board = "brya-kernelnext"
	if IsAndroid(d) || IsDevBoard(d) || !IsCros(d) {
		t.Fatalf("Device should *only* be android")
	}
}
