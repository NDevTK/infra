// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package pubsub wraps all the pubsub API interactions that will be required by SuiteScheduler.
package pubsub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"

	"infra/cros/cmd/suite_scheduler/common"
)

const (
	// MaxIdleTime is the maximum amount of time we will let the Pub/Sub receive
	// client sit idle.
	MaxIdleSeconds = 5

	loopDuration = 100 * time.Millisecond
)

// ReceiveClient defines the minimum requires that this project will need of a
// Pub/Sub API.
type ReceiveClient interface {
	initClient(projectID string) error
	initSubscription(subscriptionID string) error
	ingestMessage(ctx context.Context, msg *pubsub.Message)
	PullMessages() error
}

// ReceiveTimer defines an interface with for an auto-decrementing timer.
type ReceiveTimer interface {
	Start(parentCtxCancel context.CancelFunc)
	Refresh()
	Decrement(duration time.Duration)
	checkMillisecondsLeft() int64
}

// Timer implements the ReceiveTimer interface with thread-safe functionality.
type Timer struct {
	mutex            sync.Mutex
	maxSeconds       int
	millisecondsLeft int64
}

// checkMillisecondsLeft returns the amount of milliseconds left in a thread-safe manner.
func (t *Timer) checkMillisecondsLeft() int64 {
	t.mutex.Lock()
	retMilliseconds := t.millisecondsLeft
	t.mutex.Unlock()

	return retMilliseconds
}

// Start is a busy loop that will auto decrement the timer and call the provided
// cancel function when it has fully expired.
func (t *Timer) Start(parentCtxCancel context.CancelFunc) {
	lastTick := time.Now()

	common.Stdout.Printf("Starting the Pub/Sub timer with a max of %d seconds\n", t.maxSeconds)
	for {
		timeSince := time.Since(lastTick)
		t.Decrement(time.Duration(timeSince))
		lastTick = time.Now()

		if t.checkMillisecondsLeft() < 0 {
			// Cancel the parent context controlling this timer.
			parentCtxCancel()
			common.Stdout.Println("No time left, cancelling the timer context to end receiving from Pub/Sub")
			return
		}
		time.Sleep(loopDuration)
	}
}

// Decrement is a thread-safe function to reduce the amount of time left in the
// timer.
func (t *Timer) Decrement(duration time.Duration) {
	t.mutex.Lock()
	t.millisecondsLeft = t.millisecondsLeft - duration.Milliseconds()
	t.mutex.Unlock()
}

// Refresh sets the timer to the maximum amount of allotted time.
func (t *Timer) Refresh() {
	t.mutex.Lock()
	timerCeiling := time.Second * time.Duration(t.maxSeconds)
	t.millisecondsLeft = timerCeiling.Milliseconds()
	t.mutex.Unlock()
}

// InitTimer returns a waiting Timer set to the maximum amount of milliseconds
// provided.
func InitTimer(maxSeconds int) *Timer {
	t := &Timer{
		mutex:      sync.Mutex{},
		maxSeconds: maxSeconds,
	}

	// Set the time left using a thread-safe function.
	t.Refresh()

	return t
}

// ReceiveWithTimer implements the ReceiveClient interface with an
// auto-decrementing timer to cap idle time.
//
// NOTE: An idle timer is being implemented because the build reporting Pub/Sub
// feed is not a high QPS service so once we flush the channel, we do not expect
// more to arrive within the next hour(s). If any unexpectedly arrive after the
// receive is closed then they will be picked up in the next run.
type ReceiveWithTimer struct {
	ctx           context.Context
	cancel        context.CancelFunc
	client        *pubsub.Client
	subscription  *pubsub.Subscription
	handleMessage func(*pubsub.Message) error
	idleTimer     ReceiveTimer
}

// InitReceiveClientWithTimer returns a newly created Pub/Sub Client interface.
func InitReceiveClientWithTimer(ctx context.Context, projectID, subscriptionID string, handleMessage func(*pubsub.Message) error) (ReceiveClient, error) {
	psClient := &ReceiveWithTimer{
		handleMessage: handleMessage,
	}

	psClient.ctx, psClient.cancel = context.WithCancel(ctx)

	err := psClient.initClient(projectID)
	if err != nil {
		return nil, err
	}

	err = psClient.initSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}

	psClient.idleTimer = InitTimer(MaxIdleSeconds)

	return psClient, nil
}

// initClient creates the client interface for the current Pub/Sub Client.
func (r *ReceiveWithTimer) initClient(projectID string) error {
	common.Stdout.Printf("Initializing Pub/Sub client to %s GCP project\n", projectID)
	if r.client != nil {
		return fmt.Errorf("client is already initialized")
	}

	var err error
	r.client, err = pubsub.NewClient(r.ctx, projectID)
	if err != nil {
		return err
	}
	return nil
}

// initSubscription creates the client interface for the current Pub/Sub Client.
func (r *ReceiveWithTimer) initSubscription(subscriptionID string) error {
	common.Stdout.Printf("Initializing Pub/Sub subscription to %s \n", subscriptionID)
	if r.subscription != nil {
		return fmt.Errorf("subscription is already initialized")
	}

	rSettings := pubsub.ReceiveSettings{
		MaxOutstandingMessages: 1000,
	}

	r.subscription = r.client.Subscription(subscriptionID)
	r.subscription.ReceiveSettings = rSettings

	return nil
}

// ingestMessage places all messages into a channel buffer where they will wait
// to be processed.
func (r *ReceiveWithTimer) ingestMessage(ctx context.Context, msg *pubsub.Message) {
	r.idleTimer.Refresh()
	err := r.handleMessage(msg)
	if err != nil {
		common.Stdout.Println(err)
		msg.Nack()
		return
	}

	// TODO(b/315340446): remove blanket nack-ing once we are able to store the
	// messages in long term storage.
	msg.Nack()
}

// PullMessages does a streaming pull of all messages in the release pubsub
// feed.
func (r *ReceiveWithTimer) PullMessages() error {
	// Begin the timer. When it expires it'll cancel the Receive client's
	// context ending the blocking receive.
	go r.idleTimer.Start(r.cancel)

	// Blocking pull all messages in the feed.
	common.Stdout.Printf("Begin receiving from Pub/Sub Subscription %s\n", r.subscription.ID())
	err := r.subscription.Receive(r.ctx, r.ingestMessage)
	common.Stdout.Printf("Done receiving from Pub/Sub Subscription %s\n", r.subscription.ID())
	if err != nil {
		return err
	}

	return nil
}
