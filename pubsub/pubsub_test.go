package pubsub

import (
	"testing"
	"time"
)

func TestSubscribeAndPublish(t *testing.T) {
	ps := NewPubSub()

	sub := ps.Subscribe("news", "client1")

	// Publish a message
	count := ps.Publish("news", "Hello World")
	if count != 1 {
		t.Errorf("Expected 1 subscriber, got %d", count)
	}

	// Check message received
	select {
	case msg := <-sub.Messages:
		if msg != "Hello World" {
			t.Errorf("Expected 'Hello World', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	ps := NewPubSub()

	sub1 := ps.Subscribe("news", "client1")
	sub2 := ps.Subscribe("news", "client2")

	count := ps.Publish("news", "Breaking news!")
	if count != 2 {
		t.Errorf("Expected 2 subscribers, got %d", count)
	}

	// Both should receive
	for _, sub := range []*Subscriber{sub1, sub2} {
		select {
		case msg := <-sub.Messages:
			if msg != "Breaking news!" {
				t.Errorf("Expected 'Breaking news!', got '%s'", msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for message")
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	ps := NewPubSub()

	ps.Subscribe("news", "client1")
	ps.Subscribe("news", "client2")

	if ps.SubscriberCount("news") != 2 {
		t.Errorf("Expected 2 subscribers")
	}

	ps.Unsubscribe("news", "client1")

	if ps.SubscriberCount("news") != 1 {
		t.Errorf("Expected 1 subscriber after unsubscribe")
	}
}

func TestPublishToNonexistentChannel(t *testing.T) {
	ps := NewPubSub()

	count := ps.Publish("nonexistent", "message")
	if count != 0 {
		t.Errorf("Expected 0 subscribers, got %d", count)
	}
}
