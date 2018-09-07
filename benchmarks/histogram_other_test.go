package ddd

import (
	"testing"

	"github.com/VividCortex/gohistogram"
	"github.com/codahale/hdrhistogram"
)

var xresult float64

func benchmarkStreamingHistogram(sz int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := gohistogram.NewHistogram(20)
		for i := sz - 1; i >= 0; i-- {
			h.Add(float64(i))
		}
		xresult = h.Count()
	}
}

func BenchmarkStreamingHistogram1k(b *testing.B)   { benchmarkStreamingHistogram(1000, b) }
func BenchmarkStreamingHistogram10k(b *testing.B)  { benchmarkStreamingHistogram(10000, b) }
func BenchmarkStreamingHistogram100k(b *testing.B) { benchmarkStreamingHistogram(100000, b) }

func benchmarkHdrHistogram(sz int, b *testing.B) {
	hr := HistogramResult{}

	h := hdrhistogram.New(0, int64(sz), 1)
	for n := 0; n < b.N; n++ {
		for i := sz - 1; i >= 0; i-- {
			h.RecordValue(int64(i))
		}
		hr.min = float64(h.Min())
		hr.max = float64(h.Max())
		hr.count = float64(h.TotalCount())
		hr.p95 = float64(h.ValueAtQuantile(0.95))
		hr.median = float64(h.ValueAtQuantile(0.50))
		hr.avg = float64(h.Mean())

		h.Reset()
	}

	xresult = hr.count
}
func BenchmarkHdrHistogram1k(b *testing.B)   { benchmarkHdrHistogram(1000, b) }
func BenchmarkHdrHistogram10k(b *testing.B)  { benchmarkHdrHistogram(10000, b) }
func BenchmarkHdrHistogram100k(b *testing.B) { benchmarkHdrHistogram(100000, b) }
