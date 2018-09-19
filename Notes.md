# Notes on Datadog API

## Magic 'host:' tag for metrics

For metrics and metrics only (not events or service checks), a tag starting with `host:` should set the `hostname`, and be removed from the tags list.

## Tags should not have duplicates

In both the python and golang agents, tags are deduplicated.

## Bucketing and Flush intervals

The golang client appears to quantize stats into 10 second intervals, but upload every 15 seconds (see [pkg/aggregator/aggregator.go](https://github.com/DataDog/datadog-agent/blob/6300fb2afbe0570ea437399951691db76a45bcc4/pkg/aggregator/aggregator.go) ).  Not sure how this works exactly, it uploads one or two sets of metrics every 15s?

```
// DefaultFlushInterval aggregator default flush interval
const DefaultFlushInterval = 15 * time.Second // flush interval
const bucketSize = 10                         // fixed for now
```

The python agent (version 5), has [dd-agent/dogstatsd.py](https://github.com/DataDog/dd-agent/blob/33afda662aade99500f454b33f208e8289818d7b/dogstatsd.py):

```
# Dogstatsd constants in seconds
77	DOGSTATSD_FLUSH_INTERVAL = 10
78	DOGSTATSD_AGGREGATOR_BUCKET_SIZE = 10
```
