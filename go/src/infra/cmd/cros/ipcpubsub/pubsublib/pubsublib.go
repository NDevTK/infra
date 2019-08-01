// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// pubsublib specifies an interface around pubsub.Message for testing purposes, implements
//  a wrapper type which implements that interface, and defines a readToChannel function which
//  transmits pubsub messages to a channel where they can be subscribed to in a data-hiding way

package pubsublib

import (
	"context"

	"cloud.google.com/go/pubsub"
	"go.chromium.org/luci/common/errors"
)

// Message describes the behaviors of a pubsub message.
type Message interface {
	Attributes() map[string]string
	Body() []byte
	ID() string
	Ack()
}

// Receiver is an interface wrapping pubsub.Subscription's Receive method
type Receiver interface {
	Receive(context.Context, func(context.Context, *pubsub.Message)) error
}

var _ Receiver = &pubsub.Subscription{}

// realMessage wraps pubsub messages as a Message
type realMessage struct {
	message *pubsub.Message
}

var _ Message = &realMessage{}

// Attributes of the message
func (m *realMessage) Attributes() map[string]string {
	return m.message.Attributes
}

// Body of message, as unformatted bytes
func (m *realMessage) Body() []byte {
	return m.message.Data
}

//ID is a unique message identifier to identify messages being sent more than once
func (m *realMessage) ID() string {
	return m.message.ID
}

func (m *realMessage) Ack() {
	m.message.Ack()
}

// readToChannel pulls messages from a subscription sub until it gets an error or its context is closed
//  It returns the channels for messages and errors and the cancellation method, and is never a blocking call.
func readToChannel(ctx context.Context, sub Receiver) (<-chan Message, <-chan error, func()) {
	cctx, cancel := context.WithCancel(ctx)
	msgs := make(chan Message)
	errs := make(chan error, 1)

	handler := func(c context.Context, m *pubsub.Message) {
		select {
		case <-c.Done():
		case msgs <- &realMessage{message: m}:
		}
	}
	go func() {
		errs <- sub.Receive(cctx, handler)
		close(errs)
		cancel()
	}()
	go func() {
		<-ctx.Done()
		close(msgs)
	}()
	return msgs, errs, cancel
}

// subscribe pulls messageCount messages from the message stream msgs, returning each of them as unformatted bytes
func subscribe(ctx context.Context, msgs <-chan Message, messageCount int, filter map[string]string) ([][]byte, error) {
	storedMessages := map[string]Message{}

	for m := range msgs {
		if _, present := storedMessages[m.ID()]; present {
			m.Ack()
			continue
		}
		if !matchesFilter(filter, m) {
			m.Ack()
			continue
		}
		storedMessages[m.ID()] = m
		m.Ack()
		if len(storedMessages) >= messageCount {
			return extractBodiesFromMap(storedMessages), nil
		}
	}
	return nil, errors.Reason("subscribe ended without sufficient messages.").Err()
}

// PubsubSubscribe presents an end-to-end, single-caller interface to using Cloud Pub/Sub for IPC.
//  Its test coverage is all in the form of tests for the two principal components, since the body of PubsubSubscribe is minimal
func PubsubSubscribe(ctx context.Context, subscription *pubsub.Subscription, msgCount int, filter map[string]string) ([][]byte, error) {
	msgs, errs, cancel := readToChannel(ctx, subscription)
	defer cancel()
	received, err := subscribe(ctx, msgs, msgCount, filter)
	if err != nil {
		select {
		case e := <-errs:
			return nil, e
		default:
			return nil, err
		}
	}
	return received, nil
}

func matchesFilter(f map[string]string, m Message) bool {
	a := m.Attributes()
	if len(f) == 0 {
		return true
	}
	if len(a) == 0 {
		return false
	}
	for k, v := range f {
		if a[k] != v {
			return false
		}
	}
	return true
}

func extractBodiesFromMap(m map[string]Message) [][]byte {
	lst := make([][]byte, 0, len(m))
	for _, v := range m {
		lst = append(lst, v.Body())
	}
	return lst
}
