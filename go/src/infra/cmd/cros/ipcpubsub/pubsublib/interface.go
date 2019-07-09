package pubsublib

import "context"

// A Client manages any connection to an external service necessary for this implementation of the interface.
type Client interface {
	CreateTopic(context.Context, string) (Topic, error)
	CreateSubscription(context.Context, Topic, string) (Subscription, error)
}

// A Subscription listens for messages on a particular topic and may filter them
type Subscription interface {
	SetFilter(context.Context, Filter) error
	Receive(context.Context, func(context.Context, Message)) error
}

// A Message has an ID for detecting multiple deliveries, a possibly-empty message body, and a possibly-empty map of attributes
type Message interface {
	Attributes() map[string]string
	Body() []byte
	ID() string
}

// A Topic is a channel for messages to be published and subscribed to.
type Topic interface {
	Publish(context.Context, Message) (string, error)
	ID() string
}

// A Filter is a set of attributes to require messages to match in order to be received by the subscriber.
type Filter interface {
	FilterAttributes() map[string]string
}
