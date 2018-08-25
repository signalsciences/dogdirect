# ddd (DataDog Direct)

Directly send metrics to datadog via the HTTP api using golang

# What Problem Are We Solving?

Using DataDog's SasS offering to upload time series data when you can't use the official DataDog agent.

In general, you should use the official DataDog agent.  It's great.  Use it.  But, sometimes you can't or won't run the official DataDog agent.  This could be because:

* You are on a constrained environment with low resources
* You don't want alien code running in your environment
* You don't want to run agent process along side your application.
* You don't want to run dedicated container to run the DataDog agent (per cluster or per availability zone).

So if you can't (or won't run the datadog), package provides a simple interface for counters and gauges and uploads data using the HTTP API (the same one the DataDog agent uses).

# What does this do?

* Stores your metrics locally at **per second** resolution, supporting both counters and guages.  This might be a better than the statsd interface you've been using previously.
* Uploads your metrics to DataDog every 15 seconds

# What doesn't this do?

* Anything that's not a guage or counter: histograms, logs, traces, service checks, events.  Also not supported are tags (but would be easy enough to add).
* Error handling is probably not awesome.  Pull requests welcome.

# References and Credits

## Offical HTTP API Documentation

THe offical HTTP API documentation on [metrics](https://docs.datadoghq.com/api/?lang=bash#metrics)

> We store metric points at the 1 second resolution, but we’d prefer if you only submitted points every 15 seconds. Any metrics with fractions of a second timestamps gets rounded to the nearest second, and if any points have the same timestamp, the latest point overwrites the previous ones.

https://docs.datadoghq.com/api/?lang=bash#post-timeseries-points

## datadog-go

In particular, much of this code is based on [statsd.go](https://github.com/DataDog/datadog-go/blob/master/statsd/statsd.go).   I have mixed feelings about the "watcher" being glued into the main object but seems to work for now.

Interestingly, the buffered implimentation doesn't consolidate anything.  If you do 100 increments of a single stat, it will send 100 statsd messages.

## strip/veneue

[stripe/veneue](https://github.com/stripe/veneur) is a statds server on sterioids.  It uses the HTTP API as well.

See the [datadog section]( https://github.com/stripe/veneur/tree/master/sinks/datadog) and the [http.go](https://github.com/stripe/veneur/blob/master/http/http.go) for details.
