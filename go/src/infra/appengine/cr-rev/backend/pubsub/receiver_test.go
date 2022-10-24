// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pubsub

import (
	"context"
	"sync"
	"testing"

	"cloud.google.com/go/pubsub"
	. "github.com/smartystreets/goconvey/convey"
)

type psmObserver struct {
	acked  int
	nacked int
	mu     sync.Mutex
}

func (psmo *psmObserver) observe(ackOrNack bool) {
	psmo.mu.Lock()
	defer psmo.mu.Unlock()
	if ackOrNack {
		psmo.acked++
	} else {
		psmo.nacked++
	}
}

type mockPubsubReceiver struct {
	messages []*pubsub.Message
}

func (m *mockPubsubReceiver) Receive(ctx context.Context, f func(ctx context.Context, m *pubsub.Message)) error {
	for _, message := range m.messages {
		f(ctx, message)
	}
	return nil
}

type mockProcessMessage struct {
	calls int
}

func (m *mockProcessMessage) processPubsubMessage(ctx context.Context,
	event *SourceRepoEvent) error {
	m.calls++
	return nil
}

func TestPubsubSubscribe(t *testing.T) {
	t.Skip("Unsafe memory hacks in mockPubsubReceiver.Receive broke when PubSub library changed its internal structs")

	Convey("no messages", t, func() {
		psmo := &psmObserver{}
		ctx := WithObserver(context.Background(), psmo.observe)
		mReceiver := &mockPubsubReceiver{
			messages: make([]*pubsub.Message, 0),
		}
		mProcess := &mockProcessMessage{}

		err := Subscribe(ctx, mReceiver, mProcess.processPubsubMessage)
		So(err, ShouldBeNil)
		So(psmo.acked, ShouldEqual, 0)
		So(psmo.nacked, ShouldEqual, 0)
	})

	Convey("invalid message", t, func() {
		psmo := &psmObserver{}
		ctx := WithObserver(context.Background(), psmo.observe)
		mReceiver := &mockPubsubReceiver{
			messages: []*pubsub.Message{
				{
					Data: []byte("foo"),
				},
			},
		}
		mProcess := &mockProcessMessage{}

		err := Subscribe(ctx, mReceiver, mProcess.processPubsubMessage)
		So(err, ShouldBeNil)
		So(psmo.acked, ShouldEqual, 0)
		So(psmo.nacked, ShouldEqual, 1)
	})

	Convey("valid message", t, func() {
		psmo := &psmObserver{}
		ctx := WithObserver(context.Background(), psmo.observe)
		mReceiver := &mockPubsubReceiver{
			messages: []*pubsub.Message{
				{
					Data: []byte(`
{
  "name": "projects/chromium-gerrit/repos/chromium/src",
  "url": "http://foo/",
  "eventTime": "2020-08-01T00:01:02.333333Z",
  "refUpdateEvent": {
    "refUpdates": {
      "refs/heads/master": {
        "refName": "refs/heads/master",
        "updateType": "UPDATE_FAST_FORWARD",
        "oldId": "b82e8bfe83fadac69a6cad56c06ec45b85c86e49",
        "newId": "ef279f3d5c617ebae8573a664775381fe0225e63"
      }
    }
  }
}`),
				},
			},
		}
		mProcess := &mockProcessMessage{}

		err := Subscribe(ctx, mReceiver, mProcess.processPubsubMessage)
		So(err, ShouldBeNil)
		So(psmo.acked, ShouldEqual, 1)
		So(psmo.nacked, ShouldEqual, 0)
		So(mProcess.calls, ShouldEqual, 1)
	})
}
