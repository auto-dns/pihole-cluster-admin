package pihole

import "sync"

type Cluster struct {
	clients []ClientInterface
}

func NewCluster(clients ...ClientInterface) *Cluster {
	return &Cluster{clients: clients}
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (c *Cluster) FetchQueryLogs(opts FetchQueryLogOptions) []*NodeResult[FetchQueryLogResponse] {
	var wg sync.WaitGroup
	results := make([]*NodeResult[FetchQueryLogResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			r, err := ci.FetchQueryLogs(opts)
			node := ci.GetNodeInfo()
			results[i] = &NodeResult[FetchQueryLogResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      errorString(err),
				Response:   r,
			}
		}(i, client)
	}

	wg.Wait()
	return results
}

func (c *Cluster) AddDomainRule(opts AddDomainRuleOptions) []*NodeResult[AddDomainRuleResponse] {
	var wg sync.WaitGroup
	results := make([]*NodeResult[AddDomainRuleResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			r, err := ci.AddDomainRule(opts)
			node := ci.GetNodeInfo()
			results[i] = &NodeResult[AddDomainRuleResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      errorString(err),
				Response:   r,
			}
		}(i, client)
	}

	wg.Wait()
	return results
}

func (c *Cluster) RemoveDomainRule(opts RemoveDomainRuleOptions) []*NodeResult[RemoveDomainRuleResponse] {
	var wg sync.WaitGroup
	results := make([]*NodeResult[RemoveDomainRuleResponse], len(c.clients))

	for i, client := range c.clients {
		wg.Add(1)
		go func(i int, ci ClientInterface) {
			defer wg.Done()
			err := ci.RemoveDomainRule(opts)
			node := ci.GetNodeInfo()

			results[i] = &NodeResult[RemoveDomainRuleResponse]{
				PiholeNode: node,
				Success:    err == nil,
				Error:      errorString(err),
				// Response is nil because 204 has no body
			}
		}(i, client)
	}

	wg.Wait()
	return results
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
