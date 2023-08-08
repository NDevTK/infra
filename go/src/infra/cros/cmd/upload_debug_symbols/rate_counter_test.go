// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"
	"time"
)

func TestRateCounter(t *testing.T) {
	t0 := time.Date(2022, 12, 22, 0, 0, 0, 0, time.UTC)
	r := NewRateCounter(10 * time.Second)

	verify := func(elapsed_seconds, expected int) {
		if n := r.GetRate(t0.Add(time.Duration(elapsed_seconds) * time.Second)); n != expected {
			t.Errorf("GetRate(at %ds)=%d, want %d", elapsed_seconds, n, expected)
		}
	}

	verify(1, 0)

	r.Add(t0.Add(time.Second))
	r.Add(t0.Add(2 * time.Second))
	r.Add(t0.Add(4 * time.Second))
	r.Add(t0.Add(8 * time.Second))
	verify(8, 4)

	r.Add(t0.Add(16 * time.Second))
	verify(16, 2)
	verify(19, 1)
	verify(27, 0)
}
