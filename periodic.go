package dogdirect

import (
	"time"
)

// FlushCloser is similar to a io.WriteCloser
type FlushCloser interface {
	Flush() error
	Close() error
}

// Periodic handles flushing data periodically, also satifies the
// FlushCloser interface
type Periodic struct {
	client FlushCloser
	stop   chan struct{}
}

// NewPeriodic task create a new ticket to flush data at regular intervals
func NewPeriodic(client FlushCloser, duration time.Duration) *Periodic {
	c := &Periodic{
		client: client,
	}
	go c.watch(duration)
	return c
}

func (p *Periodic) watch(duration time.Duration) {
	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			// TODO error is squashed
			if err := p.Flush(); err != nil {
				// TODO: need call out
			}
		case <-p.stop:
			ticker.Stop()
			return
		}
	}
}

// Flush causes data to be written out
func (p *Periodic) Flush() error {
	return p.client.Flush()
}

// Close stops the ticket and closes the client
func (p *Periodic) Close() error {
	if p == nil {
		return nil
	}
	select {
	case p.stop <- struct{}{}:
	default:
	}
	return p.client.Close()
}
