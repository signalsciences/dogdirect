package dogdirect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBasicTest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		want := `{"series":[{"metric":"foobar","points":[[1640995200,123.4]],"type":"gauge","host":"hostname"}]}`
		if string(got) != want {
			t.Errorf("got %s, want %s", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	endpointv1 = ts.URL
	api := NewAPI("foo", "bar", 0)
	c := New("hostname", api)
	c.now = func() float64 {
		return float64(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC).Unix())
	}

	c.Incr("counter", []string{"tag1", "role:foo"})
	c.Incr("anotherc", nil)
	c.Incr("counter", nil)
	c.Incr("anotherc", nil)

	c.Gauge("foobar", 123.4, nil)
	c.Gauge("foobar", 666.6, nil)

	for i := 0; i < 10; i++ {
		c.Histogram("histo", float64(i), nil)
	}
	time.Sleep(time.Second * 2)
	c.Incr("counter", nil)
	c.Incr("counter", nil)
	snap := c.Snapshot()

	snap.finalize(c.now())
	if err := c.Flush(); err != nil {
		t.Errorf("c.Flush() with nil snapshot error: %v", err)
	}

	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("can't marshal: %s", err)
	}
	fmt.Printf("raw = %s", string(raw))

	c.Gauge("foobar", 123.4, nil)
	if err := c.Flush(); err != nil {
		t.Fatalf("c.Flush(): %v", err)
	}
}
