package pubsublib

import (
	"context"
	"sync"
	"testing"

	"cloud.google.com/go/pubsub"
)

type dummyReceiver struct {
	input []pubsub.Message
}

func (r *dummyReceiver) Receive(ctx context.Context, handler func(context.Context, *pubsub.Message)) error {
	wg := sync.WaitGroup{}
	c := make(chan pubsub.Message)

	go func() {
		for _, m := range r.input {
			c <- m
		}
		close(c)
	}()

	wg.Add(len(r.input))
	for m := range c {
		go func(msg pubsub.Message) {
			handler(ctx, &msg)
			wg.Done()
		}(m)
	}
	wg.Wait()
	<-ctx.Done()
	return nil
}

var _ Receiver = &dummyReceiver{}

func newReceiver(queue []pubsub.Message) Receiver {
	return &dummyReceiver{
		input: queue,
	}
}

func TestReadOneMessage(t *testing.T) {
	in := []pubsub.Message{{}}
	s := newReceiver(in)
	ctx := context.Background()
	ms, e, can := readToChannel(ctx, s)
	defer can()
loop:
	for {
		select {
		case m, open := <-ms:
			if m != nil {
				break loop
			}
			if !open {
				break loop
			}
		case err := <-e:
			if err != nil {
				t.Fatalf("Error while reading messages to channel: %v\n", err)
			}
		}
	}
}
