package dogdirect

import (
	"fmt"
)

type HostTagger struct {
	api      API
	hostname string
	tags     []string
	tagged   bool
}

func NewHostTagger(api API, hostname string, tags []string) *HostTagger {
	return &HostTagger{
		api:      api,
		hostname: hostname,
		tags:     tags,
	}
}

func (ht *HostTagger) Flush() error {
	if ht.tagged {
		return nil
	}
	if len(ht.tags) == 0 {
		ht.tagged = true
		return nil
	}
	if err := ht.api.AddHostTags(ht.hostname, "", ht.tags); err != nil {
		return fmt.Errorf("unable to set hosttags for %q: %v", ht.hostname, err)
	}
	ht.tagged = true
	return nil
}

func (ht *HostTagger) Close() error {
	return nil
}
