package poller

import (
	"context"
	"time"
)

type Config struct {
	Interval    time.Duration
	GracePeriod time.Duration // how long to wait for subs to come back
	JitterRatio float64       // e.g. 0.2 for ±20%, 0 = none
}

type Runner struct {
	broker broker
	cfg    Config
}

func New(b broker, cfg Config) *Runner {
	return &Runner{broker: b, cfg: cfg}
}

func (r *Runner) Run(ctx context.Context, sweep func(context.Context)) {
	jitter := func(d time.Duration) time.Duration {
		if r.cfg.JitterRatio <= 0 {
			return d
		}
		j := time.Duration(float64(d) * r.cfg.JitterRatio)
		n := time.Now().UnixNano() % int64(2*j+1)
		return d - j + time.Duration(n)
	}

	for {
		// idle until we have at least one subscriber
		if r.broker.SubscriberCount() == 0 {
			select {
			case <-r.broker.SubscribersChanged():
				if r.broker.SubscriberCount() == 0 {
					continue
				}
			case <-ctx.Done():
				return
			}
		}

		// active
		sweep(ctx)
		t := time.NewTicker(jitter(r.cfg.Interval))
		running := true

		for running {
			select {
			case <-t.C:
				sweep(ctx)

			case <-r.broker.SubscribersChanged():
				if r.broker.SubscriberCount() == 0 {
					if r.cfg.GracePeriod > 0 {
						g := time.NewTimer(r.cfg.GracePeriod)
						select {
						case <-g.C:
							if r.broker.SubscriberCount() == 0 {
								running = false
							}
						case <-r.broker.SubscribersChanged():
							// subs came back; keep running
						case <-ctx.Done():
							t.Stop()
							g.Stop()
							return
						}
						g.Stop()
					} else {
						running = false
					}
				}

			case <-ctx.Done():
				t.Stop()
				return
			}
		}

		t.Stop()
	}
}
