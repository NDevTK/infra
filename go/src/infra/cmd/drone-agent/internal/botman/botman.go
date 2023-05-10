// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package botman provides a bot manager that ensures that designated
// Swarming bots are running, restarting them if necessary.
package botman

import (
	"log"
	"sync"

	"infra/cmd/drone-agent/internal/bot"
)

// WorldHook defines the interface that a Botman uses to
// interact with the external world.
type WorldHook interface {
	// StartBot starts a bot process for the given ID.
	// This method should be safe to call concurrently.
	StartBot(id string) (bot.Bot, error)
	// ReleaseResources is called to release resources for a bot process
	// that has finished.  This method should be idempotent.
	ReleaseResources(id string)
}

// Botman manages running Swarming bots.  Callers tell Botman what
// bots to add, drain, or terminate using an ID, and Botman makes sure
// there are bots running or not running for those IDs.  IDs may refer
// to resources such as DUTs or some arbitrary index of bots to run.
type Botman struct {
	hook WorldHook
	wg   sync.WaitGroup

	// The following fields are covered by the mutex.
	m       sync.Mutex
	blocked bool
	bots    map[string]botSignals
}

// NewBotman creates a new Botman.
func NewBotman(h WorldHook) *Botman {
	c := &Botman{
		hook: h,
		bots: make(map[string]botSignals),
	}
	return c
}

// AddBot adds a bot to the Botman.
// The controller ensures that an instance Swarming bot is running for the given resource ID.
// If the bot was already added or if the controller is blocked, do nothing.
// This method is concurrency safe.
func (c *Botman) AddBot(id string) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.blocked {
		return
	}
	if _, ok := c.bots[id]; ok {
		// ID already has bot running.
		return
	}
	log.Printf("Starting new bot for ID %v", id)
	s := newBotSignals()
	c.bots[id] = s
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		runBotForID(c.hook, id, s)
		c.m.Lock()
		delete(c.bots, id)
		c.m.Unlock()
	}()
}

// runBotForID keeps a Swarming bot running for the ID.
// Signals to drain or terminate should be sent using botSignals.
// This function otherwise runs forever.
func runBotForID(h WorldHook, id string, s botSignals) {
	defer h.ReleaseResources(id)
	for {
		select {
		case <-s.drain:
			return
		case <-s.terminate:
			return
		default:
		}
		b, err := h.StartBot(id)
		if err != nil {
			log.Printf("Fail to start bot %s %s", id, err)
			continue
		}
		wait := make(chan struct{})
		go func() {
			_ = b.Wait()
			close(wait)
		}()
		var stop bool
	listenForSignals:
		for {
			select {
			case <-s.drain:
				_ = b.Drain()
				stop = true
			case <-s.terminate:
				if err = b.TerminateOrKill(); err != nil {
					log.Printf("Failed to terminate or kill bot: %s", err)
				}
				stop = true
			case <-wait:
				break listenForSignals
			}
		}
		if stop {
			return
		}
	}
}

// DrainBot removes an ID to no longer have bots running for it and
// drains its current bot.
// This method can be called repeatedly.
// If the controller does not have the ID, just call ReleaseResources on
// the controller's hook.
// This method is concurrency safe.
func (c *Botman) DrainBot(id string) {
	c.m.Lock()
	s, ok := c.bots[id]
	c.m.Unlock()
	if ok {
		log.Printf("Draining Bot with ID %v", id)
		s.sendDrain()
	} else {
		c.hook.ReleaseResources(id)
	}
}

// TerminateBot removes an ID to no longer have bots running for it
// and terminates its current bot.
// This method can be called repeatedly.
// If the controller does not have the ID, just call ReleaseResources on
// the controller's hook.
// This method is concurrency safe.
func (c *Botman) TerminateBot(id string) {
	c.m.Lock()
	s, ok := c.bots[id]
	c.m.Unlock()
	if ok {
		log.Printf("Terminating Bot with ID %v", id)
		s.sendTerminate()
	} else {
		c.hook.ReleaseResources(id)
	}
}

// DrainAll drains all Bots.
// You almost certainly want to call BlockBots first to make sure Bots
// don't get added right after calling this.
func (c *Botman) DrainAll() {
	c.m.Lock()
	for _, s := range c.bots {
		s.sendDrain()
	}
	c.m.Unlock()
}

// TerminateAll terminates all Bots.
// You almost certainly want to call BlockBots first to make sure Bots
// don't get added right after calling this.
func (c *Botman) TerminateAll() {
	c.m.Lock()
	for _, s := range c.bots {
		s.sendTerminate()
	}
	c.m.Unlock()
}

// BlockBots marks the controller to not accept new Bots.
// This method is safe to call concurrently.
func (c *Botman) BlockBots() {
	c.m.Lock()
	c.blocked = true
	c.m.Unlock()
}

// ActiveBots returns a slice of all Bots the controller is keeping alive.
// This includes Bots that are draining or terminated but not exited yet.
// This method is safe to call concurrently.
func (c *Botman) ActiveBots() []string {
	var ds []string
	c.m.Lock()
	for d := range c.bots {
		ds = append(ds, d)
	}
	c.m.Unlock()
	return ds
}

// Wait for all Swarming bots to finish.  It is the caller's
// responsibility to make sure all bots are terminated or drained,
// else this call will hang.
func (c *Botman) Wait() {
	c.wg.Wait()
}

type botSignals struct {
	drain     chan struct{}
	terminate chan struct{}
}

func newBotSignals() botSignals {
	return botSignals{
		drain:     make(chan struct{}, 1),
		terminate: make(chan struct{}, 1),
	}
}

func (s botSignals) sendDrain() {
	select {
	case s.drain <- struct{}{}:
	default:
	}
}

func (s botSignals) sendTerminate() {
	select {
	case s.terminate <- struct{}{}:
	default:
	}
}
