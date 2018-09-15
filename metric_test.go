package dogdirect

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestBasicTest(t *testing.T) {
	c, err := New("hostname", "myapp", "foo", []string{"tag1", "tag2"})
	if err != nil {
		t.Fatalf("unable to create: %s", err)
	}
	c.Incr("counter")
	c.Incr("counter")
	c.Gauge("foobar", 123.4)
	c.Gauge("foobar", 666.6)
	for i := 0; i < 10; i++ {
		c.Histogram("histo", float64(i))
	}
	time.Sleep(time.Second)
	c.Incr("counter")
	c.Incr("counter")
	snap := c.Snapshot()
	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("can't marhsall: %s", err)
	}
	fmt.Printf("raw = %s", string(raw))
}
