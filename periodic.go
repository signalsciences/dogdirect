package dogdirect

import (
	"time"
)

// FlushCloser is similar to a io.WriteCloser
type FlushCloser interface {
	Flush() error
	Close() error
}

// Multi is a list of a FlushClosers
type Multi []FlushCloser

// Flush flushes all tasks
//  All tasks are attempted to be flushed, first error if any is returned.
//  Currently done in serial, although could be done in parallel
func (m Multi) Flush() error {
	var errout error
	for _, f := range m {
		if err := f.Flush(); err != nil && errout == nil {
			errout = err
		}
	}
	return errout
}

// Close closes all tasks.
// All tasks are attempted to be closed, first error if any is returned
func (m Multi) Close() error {
	var errout error
	for _, f := range m {
		if err := f.Close(); err != nil && errout == nil {
			errout = err
		}
	}
	return errout
}

// Periodic handles flushing data periodically
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
