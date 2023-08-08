// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import "time"

// RateCounter counts the number of events recently happened.
// All the methods, including Add() and GetRate(), requires that a sequence of
// "current" time values given to them over multiple invocations is
// monotonically increasing. Otherwise, the results will be inaccurate.
// This condition is satisfied if the caller always pass time.Now().
type RateCounter struct {
	window  time.Duration
	history []time.Time
}

// NewRateCounter creates a new initialized RateCounter with the given time
// window size.
func NewRateCounter(window time.Duration) *RateCounter {
	return &RateCounter{window: window}
}

// Add records a new history of the time when a new event occurred.
func (r *RateCounter) Add(current time.Time) {
	r.history = append(r.history, current)
	r.removeStaleHistory(current)
}

// GetRate returns the number of events happened within the time window until
// the current time.
func (r *RateCounter) GetRate(current time.Time) int {
	r.removeStaleHistory(current)
	return len(r.history)
}

func (r *RateCounter) removeStaleHistory(current time.Time) {
	for len(r.history) > 0 && current.Sub(r.history[0]) > r.window {
		r.history[0] = time.Time{}
		r.history = r.history[1:]
	}
}
