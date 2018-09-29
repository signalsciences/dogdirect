package dogdirect

import (
	"time"
)

// FlushCloser is similar to a io.WriteCloser
type FlushCloser interface {
	Flush() error
	Close() error
}

// MultiTask is a sequence of FlushClosers
//  It operates in serial, although the underlyding implimentation
//  can work in parallel.  See Periodic below
type MultiTask []FlushCloser

// Flush writes out all data and returns first error if any
func (mt *MultiTask) Flush() error {
	var errout error
	for _, task := range *mt {
		if err := task.Close(); err != nil && errout == nil {
			errout = err
		}
	}
	return errout
}

// Close writes out any buffered data and shutdown client
func (mt *MultiTask) Close() error {
	var errout error
	for _, task := range *mt {
		if err := task.Close(); err != nil && errout == nil {
			errout = err
		}
	}
	return errout
}

// Periodic handles flushing data periodically, also satifies the
// FlushCloser interface so it can be used in MutliTask
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
