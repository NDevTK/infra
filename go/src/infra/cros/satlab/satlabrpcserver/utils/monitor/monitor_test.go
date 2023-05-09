// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package monitor

import (
	"testing"
	"time"

	"infra/cros/satlab/satlabrpcserver/utils/sized_queue"
)

type S struct {
	data sized_queue.SizedQueue[int]
}

func NewMock(capacity int) S {
	return S{
		data: sized_queue.New[int](capacity),
	}
}

func (i *S) Observe() {
	i.data.Push(1)
}

func TestMonitorShouldWork(t *testing.T) {
	t.Parallel()
	// Create a mock object
	r := NewMock(20)
	// Create a monitor
	m := New()
	// Observe the obj every 2 sec
	m.Register(&r, time.Second*2)
	// Sleep 5 sec
	time.Sleep(time.Second * 5)
	// Assert
	if r.data.Size() != 3 {
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
	// Observe the obj every 2 sec
	m.Register(&r1, time.Second*2)
	m.Register(&r2, time.Second)
	// Sleep 5 sec
	time.Sleep(time.Second * 5)
	// Assert
	if r1.data.Size() != 3 {
		t.Errorf("Observe isn't expected")
	}
	if r2.data.Size() != 5 {
		t.Errorf("Observe isn't expected")
	}
}
