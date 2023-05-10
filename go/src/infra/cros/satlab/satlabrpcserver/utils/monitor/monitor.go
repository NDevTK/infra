// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package monitor

import (
	"context"
	"log"
	"sync"
	"time"
)

// Observable provides the ability that Monitor want to observe.
type Observable interface {
	// Observe the behavior, which Monitor run it immediately and schedule the
	// behavior in later. After reaching the interval, Monitor will run and schedule it again
	Observe()
}

// Monitor uses to run and schedule Some Tasks (Observable) that we
// want to keep observing it after some period.
type Monitor struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// New a monitor
func New() Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return Monitor{
		ctx:        ctx,
		cancelFunc: cancel,
		wg:         sync.WaitGroup{},
	}
}

// Register the Task (Observable). It will run the task immediately and then
// schedule it. After the interval, it will run and schedule it again.
func (m *Monitor) Register(obj Observable, interval time.Duration) {
	m.wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		for {
			obj.Observe()
			select {
			case <-ctx.Done():
				wg.Done()
			default:
				time.Sleep(interval)
			}
		}
	}(m.ctx, &m.wg)
}

// Stop the monitor.
func (m *Monitor) Stop() {
	log.Printf("Shutdown now")
	m.cancelFunc()
	m.wg.Wait()
}
