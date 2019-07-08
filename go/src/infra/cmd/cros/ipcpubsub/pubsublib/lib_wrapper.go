package pubsublib

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/api/option"
)

type RealMessage struct {
	message pubsub.Message
}

var _ Message = &RealMessage{}

func (m *RealMessage) Attributes() map[string]string {
	return m.message.Attributes
}

func (m *RealMessage) Body() []byte {
	return m.message.Data
}

func (m *RealMessage) ID() string {
	return m.message.ID
}

func NewMessage(bs []byte, attrs map[string]string) *RealMessage {
	msg := pubsub.Message{
		Data:       bs,
		Attributes: attrs,
	}
	return &RealMessage{message: msg}
}

type RealTopic struct {
	topic *pubsub.Topic
}

func (t *RealTopic) Publish(ctx context.Context, msg Message) (string, error) {
	realMessage := NewMessage(msg.Body(), msg.Attributes())
	result := t.topic.Publish(ctx, &realMessage.message)
	return result.Get(ctx)
}

func (t *RealTopic) ID() string {
	return t.topic.ID()
}

var _ Topic = &RealTopic{}

type RealClient struct {
	client pubsub.Client
}

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

func (c *RealClient) CreateSubscription(ctx context.Context, t Topic, subName string) (Subscription, error) {
	sub := c.client.Subscription(subName)
	realTopic, err := c.TransmuteTopic(ctx, t)
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

func NewClient(ctx context.Context, projectName string, opts ...option.ClientOption) (*RealClient, error) {
	cli, err := pubsub.NewClient(ctx, projectName, opts...)
	if err != nil {
		return nil, err
	}
	return &RealClient{client: *cli}, nil
}

func (c *RealClient) TransmuteTopic(ctx context.Context, t Topic) (*RealTopic, error) {
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

type RealSubscription struct {
	subscription *pubsub.Subscription
	filter       map[string]string
}

func (s *RealSubscription) Receive(ctx context.Context, handler func(context.Context, Message)) error {
	wrappedHandler := func(ctx context.Context, m *pubsub.Message) {
		if !s.matchesFilter(m) {
			m.Ack()
			return
		}
		message := NewMessage(m.Data, m.Attributes)
		handler(ctx, message)
		m.Ack()
	}
	return s.subscription.Receive(ctx, wrappedHandler)
}
func (s *RealSubscription) SetFilter(ctx context.Context, f Filter) error {
	s.filter = f.FilterAttributes()
	return nil
}

func (s *RealSubscription) matchesFilter(m *pubsub.Message) bool {
	if len(s.filter) == 0 {
		return true
	}
	if len(m.Attributes) == 0 {
		return false
	}
	for k, v := range s.filter {
		if m.Attributes[k] != v {
			return false
		}
	}
	return true
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
