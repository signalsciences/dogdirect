package dogdirect

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

	// error case is normal, and occurs with newly created hosts
	// Need to distinguish between "host does not exist" and something
	// more horrible.  Perhaps limit number of attempts.
	if err := ht.api.AddHostTags(ht.hostname, "", ht.tags); err == nil {
		ht.tagged = true
	}
	return nil
}

func (ht *HostTagger) Close() error {
	return nil
}
