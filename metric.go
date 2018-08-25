package ddd

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

// DDMetric is a data structure that represents the JSON that Datadog
// wants when posting to the API
type DDMetric struct {
	Name       string       `json:"metric"`
	Value      [][2]float64 `json:"points"`
	MetricType string       `json:"type"`
	Hostname   string       `json:"host,omitempty"`
	Tags       []string     `json:"tags,omitempty"`
	Interval   int32        `json:"interval,omitempty"`
}

func now() float64 {
	return float64(time.Now().UTC().Unix())
}

// NewMetric creates a new metric
func NewMetric(name, mtype, host string) *DDMetric {
	return &DDMetric{
		Name:       name,
		MetricType: mtype,
		Hostname:   host,
	}
}

// Add uses a new observation and adjusts the metric accordingly
func (m *DDMetric) Add(now float64, val float64) {
	vlast := len(m.Value) - 1

	// if first point, OR if last point is in a previous interval
	if vlast == -1 || m.Value[vlast][0] != now {
		m.Value = append(m.Value, [2]float64{now, val})
		return
	}

	// last point is in current interval
	switch m.MetricType {
	case "counter":
		// add new observation to existing
		m.Value[vlast][1] += val
	case "gauge":
		// overwrite with new observation
		m.Value[vlast][1] = val
	default:
		panic("metric type not supported")
	}
}

// Client is the main datastructure of metrics to upload
type Client struct {
	Series   []*DDMetric          `json:"series"` // raw data
	hostname string               // hostname
	metrics  map[string]*DDMetric // map of name to metric for fast lookup

	now       func() float64 // for testing
	writer    io.WriteCloser // where output goes
	flushTime time.Duration  // how often to upload

	stop chan struct{}
	sync.Mutex
}

// New creates a new datadog client
func New(hostname string, apikey string) (*Client, error) {
	client := &Client{
		now:       now,
		hostname:  hostname,
		metrics:   make(map[string]*DDMetric),
		flushTime: time.Second * 15,
		stop:      make(chan struct{}, 1),
		writer:    NewWriter(apikey, time.Second*5),
	}
	go client.watch()
	return client, nil
}

func (c *Client) watch() {
	ticker := time.NewTicker(c.flushTime)

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
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(name, "gauge", c.hostname)
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	m.Add(c.now(), value)
	return nil
}

// Count represents a count of events
func (c *Client) Count(name string, value float64) error {
	c.Lock()
	m, ok := c.metrics[name]
	if !ok {
		m = NewMetric(name, "counter", c.hostname)
		c.Series = append(c.Series, m)
		c.metrics[name] = m
	}
	m.Add(c.now(), value)
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

// Snapshot makes a copy of the data and resets everything locally
func (c *Client) Snapshot() *Client {
	c.Lock()
	if len(c.Series) == 0 {
		c.Unlock()
		return nil
	}
	ccopy := Client{
		Series: c.Series,
	}
	c.metrics = make(map[string]*DDMetric)
	c.Series = nil
	c.Unlock()
	return &ccopy
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
