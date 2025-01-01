package pubsub

import (
	"sync"
)

// Subscriber represents a client subscribed to channels.
type Subscriber struct {
	ID       string
	Messages chan string
}

// PubSub manages channel subscriptions and message publishing.
type PubSub struct {
	mu       sync.RWMutex
	channels map[string]map[string]*Subscriber // channel -> subscriberID -> Subscriber
}

// NewPubSub creates a new PubSub instance.
func NewPubSub() *PubSub {
	return &PubSub{
		channels: make(map[string]map[string]*Subscriber),
	}
}

// Subscribe adds a subscriber to a channel.
func (ps *PubSub) Subscribe(channel, subscriberID string) *Subscriber {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, ok := ps.channels[channel]; !ok {
		ps.channels[channel] = make(map[string]*Subscriber)
	}

	sub := &Subscriber{
		ID:       subscriberID,
		Messages: make(chan string, 100), // Buffered channel
	}
	ps.channels[channel][subscriberID] = sub
	return sub
}

// Unsubscribe removes a subscriber from a channel.
func (ps *PubSub) Unsubscribe(channel, subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.channels[channel]; ok {
		if sub, ok := subs[subscriberID]; ok {
			close(sub.Messages)
			delete(subs, subscriberID)
		}
		if len(subs) == 0 {
			delete(ps.channels, channel)
		}
	}
}

// UnsubscribeAll removes a subscriber from all channels.
func (ps *PubSub) UnsubscribeAll(subscriberID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for channel, subs := range ps.channels {
		if sub, ok := subs[subscriberID]; ok {
			close(sub.Messages)
			delete(subs, subscriberID)
		}
		if len(subs) == 0 {
			delete(ps.channels, channel)
		}
	}
}

// Publish sends a message to all subscribers of a channel.
// Returns the number of subscribers that received the message.
func (ps *PubSub) Publish(channel, message string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subs, ok := ps.channels[channel]
	if !ok {
		return 0
	}

	count := 0
	for _, sub := range subs {
		select {
		case sub.Messages <- message:
			count++
		default:
			// Channel buffer full, skip this subscriber
		}
	}
	return count
}

// SubscriberCount returns the number of subscribers for a channel.
func (ps *PubSub) SubscriberCount(channel string) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, ok := ps.channels[channel]; ok {
		return len(subs)
	}
	return 0
}
