package dogdirect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	Name       string        `json:"metric"`
	Value      [1][2]float64 `json:"points"`
	MetricType string        `json:"type"`
	Hostname   string        `json:"host,omitempty"`
	Tags       []string      `json:"tags,omitempty"`
	Interval   int           `json:"interval,omitempty"`
}

func now() float64 {
	return float64(time.Now().UTC().Unix())
}

// NewMetric creates a new metric
func NewMetric(name string, mtype string, tags []string) *Metric {
	return &Metric{
		Name:       name,
		MetricType: mtype,
		Tags:       tags,
	}
}

// Client is the main datastructure of metrics to upload
type Client struct {
	Series     []*Metric          `json:"series"` // raw data
	hostname   string             // hostname
	namespace  string             // namespace prefix if any
	tags       []string           // global tags, if any
	metrics    map[string]*Metric // map of name to metric for fast lookup
	histograms map[string]*ExactHistogram
	now        func() float64 // for testing
	writer     io.WriteCloser // where output goes
	flushTime  int            // how often to upload in seconds

	stop chan struct{}
	sync.Mutex
}

// New creates a new datadog client
func New(hostname string, apikey string, namespace string, tags []string) (*Client, error) {

	// if we have a namespace, and it doesn't end in a "." then add one
	if namespace != "" && namespace[len(namespace)-1] != '.' {
		namespace += "."
	}

	client := &Client{
		now:        now,
		hostname:   hostname,
		namespace:  namespace,
		tags:       tags,
		metrics:    make(map[string]*Metric),
		histograms: make(map[string]*ExactHistogram),
		flushTime:  15,
		stop:       make(chan struct{}, 1),
		writer:     NewWriter(apikey, time.Second*5),
	}
	go client.watch()
	return client, nil
}

func (c *Client) watch() {
	ticker := time.NewTicker(time.Second * time.Duration(c.flushTime))

	for {
		select {
		case <-ticker.C:
			// TODO error is squashed
			if err := c.Flush(); err != nil {
				// TODO: need call out
			}
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

// Gauge represent an observation
func (c *Client) Gauge(name string, value float64) error {
	c.Lock()
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(c.namespace+name, TypeGauge, c.tags)
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	m.Value[0][1] = value
	c.Unlock()
	return nil
}

// Count represents a count of events
func (c *Client) Count(name string, value float64) error {
	c.Lock()
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(c.namespace+name, TypeRate, c.tags)
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	m.Value[0][1] += value / float64(c.flushTime)
	c.Unlock()
	return nil
}

// Incr adds one event count, same as Count(name, 1)
func (c *Client) Incr(name string) error {
	return c.Count(name, 1.0)
}

// Decr subtracts one event, same as Count(name, -1)
func (c *Client) Decr(name string) error {
	return c.Count(name, -1.0)
}

// Timing records a duration
func (c *Client) Timing(name string, val time.Duration) error {
	// datadog works in milliseconds
	return c.Histogram(name, val.Seconds()*1000)
}

// Histogram records a value that will be used in aggregate
func (c *Client) Histogram(name string, val float64) error {
	c.Lock()
	h := c.histograms[name]
	if h == nil {
		h = NewExactHistogram(1000)
		c.histograms[name] = h
	}
	h.Add(val)
	c.Unlock()
	return nil
}

// Snapshot makes a copy of the data and resets everything locally
func (c *Client) Snapshot() *Client {
	c.Lock()
	if len(c.Series) == 0 && len(c.histograms) == 0 {
		c.Unlock()
		return nil
	}
	snap := Client{
		Series:     c.Series,
		metrics:    c.metrics,
		histograms: c.histograms,
		flushTime: c.flushTime,
	}
	c.metrics = make(map[string]*Metric)
	c.histograms = make(map[string]*ExactHistogram)
	c.Series = nil
	c.Unlock()

	// now for histograms, convert to various descriptive statistic guages
	for name, h := range snap.histograms {
		hr := h.Flush()
		if hr.count == 0 {
			continue
		}
		snap.Count(name+".count", hr.count)
		snap.Gauge(name+".max", hr.max)
		snap.Gauge(name+".avg", hr.avg)
		snap.Gauge(name+".median", hr.median)
		snap.Gauge(name+".95percentile", hr.p95)
	}

	currentTime := c.now()
	for i := 0; i < len(snap.Series); i++ {
		snap.Series[i].Value[0][0] = currentTime
		snap.Series[i].Hostname = c.hostname
		snap.Series[i].Interval = c.flushTime
	}

	return &snap
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

	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	_, err = c.writer.Write(raw)
	return err
}

// Close the client connection.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	select {
	case c.stop <- struct{}{}:
	default:
	}

	// make best attempt at closing writer
	err1 := c.Flush()
	err2 := c.writer.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// DirectAPIWriter handles uploading data to DataDog's api endpoint
type DirectAPIWriter struct {
	endpoint string
	client   *http.Client
}

// NewWriter creates a new uploader to DataDog's api
func NewWriter(apikey string, timeout time.Duration) *DirectAPIWriter {
	return &DirectAPIWriter{
		endpoint: "https://api.datadoghq.com/api/v1/series?api_key=" + apikey,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Close satisfies io.Closer
func (d *DirectAPIWriter) Close() error {
	d = nil
	return nil
}

// Write satifies io.Writer
func (d *DirectAPIWriter) Write(data []byte) (int, error) {
	body := bytes.NewReader(data)
	req, err := http.NewRequest(http.MethodPost, d.endpoint, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			// if the error has the url in it, then retrieve the inner error
			// and ditch the url (which might contain secrets)
			err = urlErr.Err
		}
		return 0, err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return 0, fmt.Errorf("http status %v: %s", resp.StatusCode, string(responseBody))
	}

	return len(data), nil
}
