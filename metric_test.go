package dogdirect

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestBasicTest(t *testing.T) {
	c, err := New("hostname", "apikey")
	if err != nil {
		t.Fatalf("unable to create: %s", err)
	}
	c.Incr("counter", []string{"tag1", "tag2"})
	c.Incr("anotherc", nil)
	c.Incr("counter", nil)
	c.Incr("anotherc", nil)

	c.Gauge("foobar", 123.4, nil)
	c.Gauge("foobar", 666.6, nil)

	for i := 0; i < 10; i++ {
		c.Histogram("histo", float64(i), nil)
	}
	time.Sleep(time.Second)
	c.Incr("counter", nil)
	c.Incr("counter", nil)
	snap := c.Snapshot()
	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("can't marhsall: %s", err)
	}
	fmt.Printf("raw = %s", string(raw))
}
