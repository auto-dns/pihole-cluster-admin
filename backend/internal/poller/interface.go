package poller

type broker interface {
	SubscriberCount() int64
	SubscribersChanged() <-chan struct{}
}
