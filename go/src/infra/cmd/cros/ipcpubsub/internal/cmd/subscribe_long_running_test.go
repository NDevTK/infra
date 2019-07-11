// +build slow

package cmd

import (
	"context"
	"infra/cmd/cros/ipcpubsub/pubsublib"
	"testing"
)

func TestReceiveMessageSentLongAfterSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test with multi-second waits")
	}
	t.Log("testing subscribe - wait 5m - send")
	sub := newSlowSubscription("5m")
	msg := dummyMessage{
		attrs: nil,
		body:  []byte("test message"),
		id:    "1",
	}
	go func() {
		sub.inputMessages([]pubsublib.Message{&msg})
	}()
	bodies, err := Subscribe(context.Background(), sub, 1)
	if err != nil {
		t.Fatalf("Got error %v from Subscribe (shouldn't be possible)", err)
	}
	if len(bodies) != 1 {
		t.Errorf("Wrong number of messages read: expected 1, got %v", len(bodies))
	}
}
