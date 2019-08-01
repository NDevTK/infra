// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"context"
	"infra/cmd/cros/ipcpubsub/pubsublib"
	"testing"
	"time"
)

type dummyMessage struct {
	attrs map[string]string
	body  []byte
	id    string
}

func (m *dummyMessage) Attributes() map[string]string {
	return m.attrs
}

func (m *dummyMessage) Body() []byte {
	return m.body
}

func (m *dummyMessage) ID() string {
	return m.id
}

func (m *dummyMessage) Ack() {
}

func initializeChannels() (chan pubsublib.Message, chan error, chan interface{}) {
	m := make(chan pubsublib.Message)
	e := make(chan error)
	d := make(chan interface{})
	return m, e, d
}

func inputMessages(c chan<- pubsublib.Message, ms []pubsublib.Message) {
	go func() {
		for _, m := range ms {
			c <- m
		}
		time.Sleep(time.Second)
		close(c)
	}()
}

func TestSubscribeOneMessage(t *testing.T) {
	msg := dummyMessage{
		attrs: nil,
		body:  []byte("test message"),
		id:    "1",
	}
	m, e, d := initializeChannels()
	inputMessages(m, []pubsublib.Message{&msg})
	bodies, err := Subscribe(context.Background(), m, e, d, 1, nil)
	if err != nil {
		t.Fatalf("Error: %v ", err)
	}
	if len(bodies) != 1 {
		t.Errorf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
}

func TestReceiveMessageSentAfterSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test with >1s wait")
	}
	msg := dummyMessage{
		attrs: nil,
		body:  []byte("test message"),
		id:    "1",
	}

	m, e, d := initializeChannels()
	go func() {
		time.Sleep(time.Second)
		inputMessages(m, []pubsublib.Message{&msg})
	}()
	bodies, err := Subscribe(context.Background(), m, e, d, 1, nil)

	if err != nil {
		t.Fatalf("Error: %v ", err)
	}
	if len(bodies) != 1 {
		t.Errorf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
}

func TestIgnoreDuplicateMessages(t *testing.T) {
	msg1 := dummyMessage{
		attrs: nil,
		body:  []byte("foo"),
		id:    "1",
	}
	msg2 := dummyMessage{
		attrs: nil,
		body:  []byte("bar"),
		id:    "2",
	}
	msg3 := dummyMessage{
		attrs: nil,
		body:  []byte("quux"),
		id:    "3",
	}
	published := []pubsublib.Message{&msg1, &msg2, &msg1, &msg1, &msg1, &msg1, &msg2, &msg1, &msg1, &msg1, &msg3}

	m, e, d := initializeChannels()
	inputMessages(m, published)
	bodies, err := Subscribe(context.Background(), m, e, d, 3, nil)

	if err != nil {
		t.Fatalf("Error: %v ", err)
	}
	if len(bodies) != 3 {
		t.Errorf("Wrong number of messages read: expected 3, got %v", len(bodies))
	}
	expected := map[string]int{
		"foo":  1,
		"bar":  1,
		"quux": 1,
	}
	received := map[string]int{}
	for _, m := range bodies {
		t.Logf("saw message with bytes %v, string %v", m, string(m))
		received[string(m)]++
	}
	for k, v := range expected {
		if received[k] != v {
			t.Errorf("Expected to see 1 message with body %v, saw %v", k, received[k])
		}
	}
}

func TestAcceptMessagesWithExtraAttrs(t *testing.T) {
	msg1 := dummyMessage{
		attrs: map[string]string{
			"foo": "bar",
		},
		body: []byte("test message"),
		id:   "1",
	}
	published := []pubsublib.Message{&msg1}

	m, e, d := initializeChannels()
	inputMessages(m, published)
	bodies, err := Subscribe(context.Background(), m, e, d, 1, nil)

	if err != nil {
		t.Fatalf("Error: %v ", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
	if string(bodies[0]) != string(msg1.body) {
		t.Fatalf("Wrong messages accepted. Expected 'test message', got '%v'", string(bodies[0]))
	}
}

func TestRejectMessagesWithoutAttrs(t *testing.T) {
	foobarFilter := map[string]string{
		"foo": "bar",
	}
	msg1 := dummyMessage{
		attrs: nil,
		body:  []byte("manny"),
		id:    "1",
	}
	msg2 := dummyMessage{
		attrs: map[string]string{
			"spam": "eggs",
		},
		body: []byte("moe"),
		id:   "2",
	}
	msg3 := dummyMessage{
		attrs: map[string]string{
			"foo": "bar",
		},
		body: []byte("jack"),
		id:   "3",
	}
	published := []pubsublib.Message{&msg1, &msg1, &msg2, &msg2, &msg3}

	m, e, d := initializeChannels()
	inputMessages(m, published)
	bodies, err := Subscribe(context.Background(), m, e, d, 1, foobarFilter)

	if err != nil {
		t.Fatalf("Error: %v ", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
	body := string(bodies[0])
	if body != "jack" {
		t.Fatalf("Accepted a message which should have been rejected by the filter.")
	}
}
