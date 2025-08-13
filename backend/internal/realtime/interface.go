package realtime

type BrokerInterface interface {
	Subscribe(topics []string) (ch <-chan Event, cancel func())
	Publish(topic string, data []byte)
	SubscriberCount() int64
	SubscribersChanged() <-chan struct{}
}
