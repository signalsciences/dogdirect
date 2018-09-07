package dogdirect

import (
	"testing"
)

var result float64

func benchmarkExactHistogram(sz int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := NewExactHistogram(sz)
		for i := sz - 1; i >= 0; i-- {
			h.Add(float64(i))
		}
		hr := h.Flush()
		if int(hr.count) != sz {
			b.Fatalf("benchmark exact histogram failed.  count Expected %d got %d", sz, int(hr.count))
		}

		// make absolutely sure compiler doesn't optimize something away
		result = hr.count
	}

}
func BenchmarkExactHistogram1k(b *testing.B)   { benchmarkExactHistogram(1000, b) }
func BenchmarkExactHistogram10k(b *testing.B)  { benchmarkExactHistogram(10000, b) }
func BenchmarkExactHistogram100k(b *testing.B) { benchmarkExactHistogram(100000, b) }
