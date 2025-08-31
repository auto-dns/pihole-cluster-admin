package realtime

type Topic = string

// TODO: ChatGPT recommending decoupling v1 from here and introducing it at edge only - search for "SSE topics and versioning"
const (
	TopicClusterBlockingV1 Topic = "v1.cluster_blocking"
	TopicHealthSummaryV1   Topic = "v1.health_summary"
	TopicNodeHealthV1      Topic = "v1.node_health"
)
