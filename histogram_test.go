package dogdirect

import (
	"testing"
)

var result float64

func benchmarkExactHistogram(sz int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := NewExactHistogram(sz, nil)
		for i := sz - 1; i >= 0; i-- {
			h.Add(float64(i))
		}
		hr := h.Flush()
		if int(hr.Count) != sz {
			b.Fatalf("benchmark exact histogram failed.  count Expected %d got %d", sz, int(hr.Count))
		}

		// make absolutely sure compiler doesn't optimize something away
		result = hr.Count
	}

}
func BenchmarkExactHistogram1k(b *testing.B)   { benchmarkExactHistogram(1000, b) }
func BenchmarkExactHistogram10k(b *testing.B)  { benchmarkExactHistogram(10000, b) }
func BenchmarkExactHistogram100k(b *testing.B) { benchmarkExactHistogram(100000, b) }
