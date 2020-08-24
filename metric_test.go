package dogdirect

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestBasicTest(t *testing.T) {
	api := NewAPI("foo", "bar", 0)
	c := New("hostname", api, 18*time.Second)
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

	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("can't marhsall: %s", err)
	}
	fmt.Printf("raw = %s\n", string(raw))
}
