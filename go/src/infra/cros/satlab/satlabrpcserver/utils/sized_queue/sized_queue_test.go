// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package sized_queue

import (
	"testing"
)

func TestPush(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	// Add some data
	q.Push(1)
	data := q.Data()
	if len(data) != 1 {
		t.Errorf("Not Expected")
	}
}

func TestPushRespectsCapacity(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	// Add some data
	q.Push(1)
	q.Push(2)
	q.Push(3)
	data := q.Data()
	if len(data) != 3 {
		t.Errorf("Not Expected")
	}
}

func TestPop(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	// Add some data
	q.Push(1)
	data, err := q.Pop()
	if err != nil {
		t.Errorf("Not Expected")
	}
	if data != 1 {
		t.Errorf("Not Expected")
	}
}

func TestPopEmpty(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	_, err := q.Pop()
	if err == nil {
		t.Errorf("Queue should be empty")
	}
	// Add some data
	q.Push(1)
	data, err := q.Pop()
	if data != 1 {
		t.Errorf("Not Expected")
	}
	_, err = q.Pop()
	if err == nil {
		t.Errorf("Queue should be empty")
	}
}

func TestClear(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	// Add some data
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Clear()
	if q.Size() != 0 {
		t.Errorf("Clear not expected")
	}
}

func TestBasic(t *testing.T) {
	t.Parallel()
	// Create a queue with capacity
	q := New[int](3)
	_, err := q.Pop()
	if err == nil {
		t.Errorf("Should be empty")
	}
	// Push Some data
	q.Push(1)
	q.Push(2)
	q.Push(3)
	// Assert
	data, err := q.Pop()
	if err != nil {
		t.Errorf("Should not be empty")
	}
	if data != 1 {
		t.Errorf("The Order is wrong")
	}
	// Push more data that exceed the capacity, it will drop the oldest value `2` automatically
	q.Push(4)
	q.Push(5)
	// Assert
	data, err = q.Pop()
	if err != nil {
		t.Errorf("Should not be empty")
	}
	if data != 3 {
		t.Errorf("The Order is wrong")
	}
	data, err = q.Pop()
	if err != nil {
		t.Errorf("Should not be empty")
	}
	if data != 4 {
		t.Errorf("The Order is wrong")
	}
	data, err = q.Pop()
	if err != nil {
		t.Errorf("Should not be empty")
	}
	if data != 5 {
		t.Errorf("The Order is wrong")
	}
	_, err = q.Pop()
	if err == nil {
		t.Errorf("Should be empty")
	}
	// Add some data
	q.Push(1)
	q.Push(2)
	q.Push(3)
	// Assert
	data, err = q.Pop()
	if err != nil {
		t.Errorf("Should not be empty")
	}
	if data != 1 {
		t.Errorf("The Order is wrong")
	}
	// Clear
	q.Clear()
	_, err = q.Pop()
	if err == nil {
		t.Errorf("Should be empty")
	}
}
