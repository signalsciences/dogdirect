package dogdirect

import (
	"sync"
	"time"
)

/*
 * https://docs.datadoghq.com/api/?lang=python#metrics
 * https://github.com/stripe/veneur/blob/master/sinks/datadog/datadog.go
 */

/* https://docs.datadoghq.com/api/?lang=python#metrics
 * ARGUMENTS

series [required]:
Pass a JSON array where each item in the array contains the following arguments:

metric [required]:
The name of the timeseries

type [optional, default=gauge]:
Type of your metric either: gauge, rate, or count

interval [optional, default=None]:
If the type of the metric is rate or count, define the corresponding interval.

points [required]:
A JSON array of points. Each point is of the form:
[[POSIX_timestamp, numeric_value], ...]
Note: The timestamp should be in seconds, current, and its format should be a 32bit float gauge-type value. Current is defined as not more than 10 minutes in the future or more than 1 hour in the past.

host [optional]:
The name of the host that produced the metric.

tags [optional, default=None]:
A list of tags associated with the metric

*/

// todo: make proper type
const (
	TypeGauge = "gauge"
	TypeRate  = "rate"
	TypeCount = "count"
)

// Metric is a data structure that represents the JSON that Datadog
// wants when posting to the API
type Metric struct {
	Name     string        `json:"metric"`
	Value    [1][2]float64 `json:"points"`
	Type     string        `json:"type"`
	Hostname string        `json:"host,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
	Interval int           `json:"interval,omitempty"`
}

func now() float64 {
	return float64(time.Now().Unix())
}

// NewMetric creates a new metric
func NewMetric(name string, mtype string, tags []string) *Metric {
	return &Metric{
		Name: name,
		Type: mtype,
		Tags: tags,
	}
}

// Client is the main datastructure of metrics to upload
type Client struct {
	Series     []*Metric          `json:"series"` // raw data
	hostname   string             // hostname
	tags       []string           // global tags, if any
	metrics    map[string]*Metric // map of name to metric for fast lookup
	histograms map[string]*ExactHistogram
	now        func() float64 // for testing
	writer     API            // where output goes
	lastFlush  float64        // unix epoch as float64(t.Now().Unix())

	sync.Mutex
}

// New creates a new datadog metrics client
func New(hostname string, api API) (*Client, error) {
	client := &Client{
		now:        now,
		hostname:   hostname,
		metrics:    make(map[string]*Metric),
		histograms: make(map[string]*ExactHistogram),
		writer:     api,
		lastFlush:  now(),
	}
	return client, nil
}

// Gauge represents an observation
func (c *Client) Gauge(name string, value float64, tags []string) error {
	c.Lock()
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(name, TypeGauge, unique(tags))
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	m.Value[0][1] = value
	c.Unlock()
	return nil
}

// Count represents a count of events
func (c *Client) Count(name string, value float64, tags []string) error {
	c.Lock()
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(name, TypeRate, unique(tags))
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	// note, this sum must be divided by the interval length
	//  before sending.
	m.Value[0][1] += value
	c.Unlock()
	return nil
}

// Incr adds one event count, same as Count(name, 1)
func (c *Client) Incr(name string, tags []string) error {
	return c.Count(name, 1.0, tags)
}

// Decr subtracts one event, same as Count(name, -1)
func (c *Client) Decr(name string, tags []string) error {
	return c.Count(name, -1.0, tags)
}

// Timing records a duration
func (c *Client) Timing(name string, val time.Duration, tags []string) error {
	// datadog works in milliseconds
	return c.Histogram(name, val.Seconds()*1000, tags)
}

// Histogram records a value that will be used in aggregate
func (c *Client) Histogram(name string, val float64, tags []string) error {
	c.Lock()
	h := c.histograms[name]
	if h == nil {
		h = NewExactHistogram(1000, tags)
		c.histograms[name] = h
	}
	h.Add(val)
	c.Unlock()
	return nil
}

// Snapshot makes a copy of the data and resets everything locally
func (c *Client) Snapshot() *Client {
	c.Lock()
	defer func() {
		c.lastFlush = c.now()
		c.Unlock()
	}()

	if len(c.Series) == 0 && len(c.histograms) == 0 {
		return nil
	}
	snap := Client{
		hostname:   c.hostname,
		Series:     c.Series,
		metrics:    c.metrics,
		histograms: c.histograms,
		lastFlush:  c.lastFlush,
	}
	c.metrics = make(map[string]*Metric)
	c.histograms = make(map[string]*ExactHistogram)
	c.Series = nil
	return &snap
}

// not locked.. for use locally with snapshots
func (c *Client) finalize(nowUnix float64) {
	interval := nowUnix - c.lastFlush

	// histograms: convert to various descriptive statistic gauges
	for name, h := range c.histograms {
		hr := h.Flush()
		if hr.count == 0 {
			continue
		}
		c.Count(name+".count", hr.count, h.tags)
		c.Gauge(name+".max", hr.max, h.tags)
		c.Gauge(name+".avg", hr.avg, h.tags)
		c.Gauge(name+".median", hr.median, h.tags)
		c.Gauge(name+".95percentile", hr.p95, h.tags)
	}
	for i := 0; i < len(c.Series); i++ {
		c.Series[i].Value[0][0] = nowUnix
		c.Series[i].Hostname = c.hostname
		c.Series[i].Interval = int(interval)
		if c.Series[i].Type == "rate" {
			c.Series[i].Value[0][1] /= interval
		}
	}
}

// Flush forces a flush of the pending commands in the buffer
func (c *Client) Flush() error {
	if c == nil {
		return nil
	}
	snap := c.Snapshot()
	if snap == nil {
		return nil
	}

	// c.lastFlush is "now"
	snap.finalize(c.lastFlush)

	return c.writer.AddPoints(snap.Series)
}

// Close the client connection.
func (c *Client) Close() error {
	// make best attempt at closing writer
	return c.Flush()
}
