// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package monitor

import (
	"context"
	"testing"
	"time"
)

type Counter struct {
	count int
}

func NewMock(capacity int) Counter {
	return Counter{
		count: 0,
	}
}

func (c *Counter) Observe() {
	c.count += 1
}

func TestMonitorShouldWork(t *testing.T) {
	t.Parallel()

	// Create a mock object
	r := NewMock(20)

	// Create a monitor
	m := New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	m.ctx = ctx

	// Observe the obj every 2 sec
	m.Register(&r, time.Second*2)

	// Sleep 5 sec
	time.Sleep(time.Second * 3)

	// Assert
	if r.count != 2 {
		t.Errorf("Observe isn't expected")
	}
}

func TestMonitorObserveMultipleShouldWork(t *testing.T) {
	t.Parallel()

	// Create a mock object
	r1 := NewMock(20)
	r2 := NewMock(10)

	// Create a monitor
	m := New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	m.ctx = ctx

	// Observe the obj every 2 sec
	m.Register(&r1, time.Second*2)
	m.Register(&r2, time.Second*3)

	// Sleep 5 sec
	time.Sleep(time.Second * 5)

	// Assert
	if r1.count != 3 {
		t.Errorf("Observe isn't expected")
	}

	if r2.count != 2 {
		t.Errorf("Observe isn't expected")
	}
}
