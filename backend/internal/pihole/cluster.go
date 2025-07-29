package pihole

import "sync"

type Cluster struct {
	clients []ClientInterface
}

func NewCluster(clients ...ClientInterface) *Cluster {
	return &Cluster{clients: clients}
}

func (c *Cluster) FetchLogs(from, until int64) ([]*QueryLogResponse, []error) {
	var wg sync.WaitGroup
	results := make([]*QueryLogResponse, len(c.clients))
	errs := make([]error, len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			r, err := ci.FetchLogs(from, until)
			results[i] = r
			errs[i] = err
		}(i, client)
	}

	wg.Wait()
	return results, errs
}

func (c *Cluster) Logout() []error {
	var wg sync.WaitGroup
	errs := make([]error, len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			err := ci.Logout()
			errs[i] = err
		}(i, client)
	}

	wg.Wait()
	return errs
}
