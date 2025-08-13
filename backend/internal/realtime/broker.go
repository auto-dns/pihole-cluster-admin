package realtime

import (
	"sync"
	"sync/atomic"
)

type Broker struct {
	mu                 sync.RWMutex
	subscriptions      map[string]map[chan Event]struct{}
	subscribersChanged chan struct{}
	n                  atomic.Int64
}

func NewBroker() BrokerInterface {
	return &Broker{
		subscriptions:      make(map[string]map[chan Event]struct{}),
		subscribersChanged: make(chan struct{}, 1),
	}
}

func (b *Broker) poke() {
	select {
	case b.subscribersChanged <- struct{}{}:
	default:
	}
}

func (b *Broker) Subscribe(topics []string) (<-chan Event, func()) {
	ch := make(chan Event, 16)
	b.mu.Lock()
	for _, topic := range topics {
		if b.subscriptions[topic] == nil {
			b.subscriptions[topic] = make(map[chan Event]struct{})
		}
		b.subscriptions[topic][ch] = struct{}{}
	}
	b.mu.Unlock()
	b.n.Add(1)
	b.poke()

	cancel := func() {
		b.mu.Lock()
		for _, t := range topics {
			if set := b.subscriptions[t]; set != nil {
				delete(set, ch)
				if len(set) == 0 {
					delete(b.subscriptions, t)
				}
			}
		}
		b.mu.Unlock()
		b.n.Add(-1)
		b.poke()
		close(ch)
	}

	return ch, cancel
}

func (b *Broker) Publish(topic string, data []byte) {
	event := Event{
		Topic: topic,
		Data:  data,
	}
	b.mu.RLock()
	set := b.subscriptions[topic]
	for ch := range set {
		select {
		case ch <- event:
		default:
		}
	}
	b.mu.RUnlock()
}

func (b *Broker) SubscriberCount() int64 {
	return b.n.Load()
}

func (b *Broker) SubscribersChanged() <-chan struct{} {
	return b.subscribersChanged
}
