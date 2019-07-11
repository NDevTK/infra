package pubsublib

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/option"
)

// RealMessage wraps pubsub messages as a Message
type RealMessage struct {
	message pubsub.Message
}

var _ Message = &RealMessage{}

// Attributes of the message
func (m *RealMessage) Attributes() map[string]string {
	return m.message.Attributes
}

// Body of message, as unformatted bytes
func (m *RealMessage) Body() []byte {
	return m.message.Data
}

//ID is a unique message identifier to identify messages being sent more than once
func (m *RealMessage) ID() string {
	return m.message.ID
}

// NewMessage creates a message from the critical payload pieces
func NewMessage(bs []byte, attrs map[string]string) *RealMessage {
	msg := pubsub.Message{
		Data:       bs,
		Attributes: attrs,
	}
	return &RealMessage{message: msg}
}

// RealTopic wraps pubsub topics as a Topic
type RealTopic struct {
	topic *pubsub.Topic
}

// Publish broadcasts a message on the topic's channel
func (t *RealTopic) Publish(ctx context.Context, msg Message) (string, error) {
	realMessage := NewMessage(msg.Body(), msg.Attributes())
	result := t.topic.Publish(ctx, &realMessage.message)
	return result.Get(ctx)
}

// ID is the topic's name
func (t *RealTopic) ID() string {
	return t.topic.ID()
}

var _ Topic = &RealTopic{}

// RealClient wraps pubsub clients as a Client
type RealClient struct {
	client pubsub.Client
}

// CreateTopic creates a Topic on the external service with the given name.
func (c *RealClient) CreateTopic(ctx context.Context, name string) (Topic, error) {
	t := c.client.Topic(name)
	exists, err := t.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return &RealTopic{topic: t}, nil
	}

	t, err = c.client.CreateTopic(ctx, name)
	if err != nil {
		return nil, err
	}
	return &RealTopic{topic: t}, nil
}

// CreateSubscription creates a Subscription on the supplied Topic with the specified name.
func (c *RealClient) CreateSubscription(ctx context.Context, t Topic, subName string) (Subscription, error) {
	sub := c.client.Subscription(subName)
	realTopic, err := c.transmuteTopic(ctx, t)
	if err != nil {
		return nil, err
	}
	if ex, err := sub.Exists(ctx); err != nil {
		return nil, err
	} else if !ex {
		sub, err = c.client.CreateSubscription(ctx, subName, defaultSubscriptionConfig(realTopic.topic))
		if err != nil {
			return nil, err
		}
		return &RealSubscription{subscription: sub, filter: nil}, nil
	}
	if conf, err := sub.Config(ctx); err != nil {
		return nil, err
	} else if conf.Topic.ID() != realTopic.ID() {
		return nil, errors.Reason(
			"Failed to create subscription %v in topic %v. This subscription identifier is in use within a different topic.",
			subName, t.ID()).Err()
	}
	return &RealSubscription{subscription: sub, filter: nil}, nil
}

var _ Client = &RealClient{}

// NewClient creates an interface-conformant Client which can be used for direct Pubsub interfacing
func NewClient(ctx context.Context, projectName string, opts ...option.ClientOption) (*RealClient, error) {
	cli, err := pubsub.NewClient(ctx, projectName, opts...)
	if err != nil {
		return nil, err
	}
	return &RealClient{client: *cli}, nil
}

func (c *RealClient) transmuteTopic(ctx context.Context, t Topic) (*RealTopic, error) {
	switch t.(type) {
	case *RealTopic:
		return t.(*RealTopic), nil
	default:
		tp, err := c.CreateTopic(ctx, t.ID())
		if err != nil {
			return nil, err
		}
		return tp.(*RealTopic), nil
	}
}

// RealSubscription wraps pubsub subscriptions as a Subscription
type RealSubscription struct {
	subscription *pubsub.Subscription
	filter       map[string]string
}

// Receive handles a single incoming message and surfaces any errors from the message-handling process
func (s *RealSubscription) Receive(ctx context.Context, handler func(context.Context, Message)) error {
	wrappedHandler := func(ctx context.Context, m *pubsub.Message) {
		message := NewMessage(m.Data, m.Attributes)
		if !s.matchesFilter(message) {
			m.Ack()
			return
		}
		handler(ctx, message)
		m.Ack()
	}
	return s.subscription.Receive(ctx, wrappedHandler)
}

// SetFilter sets which message attributes to require in received messages
func (s *RealSubscription) SetFilter(ctx context.Context, f Filter) error {
	s.filter = f.FilterAttributes()
	return nil
}

// MatchesFilter takes an arbitrary pubsublib.Filter and pubsublib.Message and
//  checks if the message satisfies the filter.
func MatchesFilter(f Filter, m Message) bool {
	mAttrs := m.Attributes()
	fAttrs := f.FilterAttributes()
	if len(fAttrs) == 0 {
		return true
	}
	if len(mAttrs) == 0 {
		return false
	}
	for k, v := range fAttrs {
		if mAttrs[k] != v {
			return false
		}
	}
	return true
}

func (s *RealSubscription) matchesFilter(m *RealMessage) bool {
	f := RealFilter(s.filter)
	return MatchesFilter(&f, m)
}

var _ Subscription = &RealSubscription{}

func defaultSubscriptionConfig(t *pubsub.Topic) pubsub.SubscriptionConfig {
	return pubsub.SubscriptionConfig{
		Topic:               t,
		AckDeadline:         10 * time.Second,
		RetainAckedMessages: false,
		Labels:              map[string]string{},
	}
}

// RealFilter is a concrete implementation of a Filter
type RealFilter map[string]string

// FilterAttributes supplies the required attributes
func (f *RealFilter) FilterAttributes() map[string]string {
	return *f
}

var _ Filter = &RealFilter{}
