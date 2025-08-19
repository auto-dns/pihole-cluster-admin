package cluster

type Service struct {
	cluster cluster
}

func New(cluster cluster) *Service {
	return &Service{
		cluster: cluster,
	}
}
