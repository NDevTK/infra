package cmd

import (
	"context"
	"infra/cmd/cros/ipcpubsub/pubsublib"
	"testing"
)

type dummyIDlessSubscription struct {
	filter       map[string]string
	stream       chan pubsublib.Message
	messagesSeen map[string]bool
}

func newDummySubscription() *dummyIDlessSubscription {
	return &dummyIDlessSubscription{
		filter:       map[string]string{},
		stream:       make(chan pubsublib.Message),
		messagesSeen: map[string]bool{},
	}
}

func (s *dummyIDlessSubscription) SetFilter(ctx context.Context, f pubsublib.Filter) error {
	s.filter = f.FilterAttributes()
	return nil
}

func (s *dummyIDlessSubscription) Receive(ctx context.Context, h func(context.Context, pubsublib.Message)) error {
	msg := <-s.stream
	id := msg.ID()
	if _, present := s.messagesSeen[id]; present {
		return nil
	}
	if s.messagesSeen == nil {
		s.messagesSeen = map[string]bool{}
	}
	s.messagesSeen[id] = true
	d := dummyFilter(s.filter)
	if !pubsublib.MatchesFilter(&d, msg) {
		return nil
	}
	h(ctx, msg)
	return nil
}

func (s *dummyIDlessSubscription) inputMessages(msgs []pubsublib.Message) {
	for _, msg := range msgs {
		go func(m pubsublib.Message) {
			s.stream <- m
		}(msg)
	}
}

type dummyFilter map[string]string

func (f *dummyFilter) FilterAttributes() map[string]string {
	return *f
}

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

func TestSubscribeOneMessage(t *testing.T) {
	sub := newDummySubscription()
	msg := dummyMessage{
		attrs: nil,
		body:  []byte("test message"),
		id:    "1",
	}
	sub.inputMessages([]pubsublib.Message{&msg})
	bodies, err := Subscribe(context.Background(), sub, 1)
	if err != nil {
		t.Fatalf("Got error %v from Subscribe (shouldn't be possible)", err)
	}
	if len(bodies) != 1 {
		t.Errorf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
}

func TestIgnoreDuplicateMessages(t *testing.T) {
	sub := newDummySubscription()
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
	published := []pubsublib.Message{&msg1, &msg1, &msg2, &msg2, &msg3}
	sub.inputMessages(published)
	bodies, err := Subscribe(context.Background(), sub, 3)
	if err != nil {
		t.Fatalf("Got error %v from Subscribe (shouldn't be possible)", err)
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
	ctx := context.Background()
	sub := newDummySubscription()
	sub.SetFilter(ctx, &dummyFilter{})
	msg := dummyMessage{
		attrs: map[string]string{
			"foo": "bar",
		},
		body: []byte("test message"),
		id:   "1",
	}
	sub.inputMessages([]pubsublib.Message{&msg})
	bodies, err := Subscribe(ctx, sub, 1)
	if err != nil {
		t.Fatalf("Got error %v from Subscribe (shouldn't be possible)", err)
	}
	if len(bodies) != 1 {
		t.Errorf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
}

func TestRejectMessagesWithoutAttrs(t *testing.T) {
	ctx := context.Background()
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
	// Order is not guaranteed, so it is possible for this test to pass when it should fail.
	// Therefore, run three times and pass only if all three pass
	for i := 0; i < 3; i++ {
		sub := newDummySubscription()
		sub.SetFilter(ctx, &dummyFilter{
			"foo": "bar",
		})
		published := []pubsublib.Message{&msg1, &msg1, &msg2, &msg2, &msg3}
		sub.inputMessages(published)
		bodies, err := Subscribe(ctx, sub, 1)
		if err != nil {
			t.Fatalf("Got error %v from Subscribe (shouldn't be possible)", err)
		}
		if len(bodies) != 1 {
			t.Fatalf("Wrong number of messages read: expected 1, got %v", len(bodies))
		}
		body := string(bodies[0])
		if body != "jack" {
			t.Fatalf("Accepted a message which should have been rejected by the filter.")
		}
	}
}
