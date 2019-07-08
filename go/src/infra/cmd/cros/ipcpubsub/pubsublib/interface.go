package pubsublib

import "context"

type Client interface {
	CreateTopic(context.Context, string) (Topic, error)
	CreateSubscription(context.Context, Topic, string) (Subscription, error)
}

type Subscription interface {
	SetFilter(context.Context, Filter) error
	Receive(context.Context, func(context.Context, Message)) error
}

type Message interface {
	Attributes() map[string]string
	Body() []byte
	ID() string
}

type Topic interface {
	Publish(context.Context, Message) (string, error)
	ID() string
}

type Filter interface {
	FilterAttributes() map[string]string
}
