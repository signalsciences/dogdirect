package ddd

import (
	"sort"
	"sync"
)

// HistogramResults returns some descriptive statistics
// Add other stats as needed
type HistogramResult struct {
	count  float64
	min    float64
	max    float64
	avg    float64
	median float64
	p95    float64
}

// ExactHistogram is the dumbest way possible to compute various descriptive statistics
//  It keeps all data, does a sort, and the figures out various stats.
//  That said for 1000 elements, it takes under 1/20 of a millisecond to compute.
//
// Also the "sort" method is what datadog's agent does, so it can't be too painful.
//
type ExactHistogram struct {
	samples []float64
	sync.Mutex
}

func NewExactHistogram() *ExactHistogram {
	return &ExactHistogram{}
}

func (he *ExactHistogram) Add(val float64) {
	he.Lock()
	he.samples = append(he.samples, val)
	he.Unlock()
}

// Snap makes a copy of the data and resets internal state
func (he *ExactHistogram) Snap() *ExactHistogram {
	he.Lock()
	newh := &ExactHistogram{
		samples: he.samples,
	}
	he.samples = make([]float64, 0, len(he.samples))
	he.Unlock()
	return newh
}

// Flush needs to be renamed, but computes the data
func (he *ExactHistogram) Flush() HistogramResult {
	h := he.Snap()
	if len(h.samples) == 0 {
		// caller can check to see if count = 0
		return HistogramResult{}
	}

	sort.Float64s(h.samples)
	count := len(h.samples)
	sum := 0.0
	for _, val := range h.samples {
		sum += val
	}

	return HistogramResult{
		count:  float64(count),
		min:    h.samples[0],
		max:    h.samples[count-1],
		avg:    sum / float64(count),
		median: h.samples[count/2],
		p95:    h.samples[(count*95)/100],
	}
}
