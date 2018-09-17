package dogdirect

import (
	"sort"
)

// HistogramResult returns some descriptive statistics
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
}

// NewExactHistogram creates a new object
func NewExactHistogram(points int) *ExactHistogram {
	if points == 0 {
		return &ExactHistogram{}
	}
	return &ExactHistogram{
		samples: make([]float64, 0, points),
	}
}

// Add adds a data point
func (he *ExactHistogram) Add(val float64) {
	he.samples = append(he.samples, val)
}

// Flush needs to be renamed, but computes the data
func (he *ExactHistogram) Flush() HistogramResult {
	if len(he.samples) == 0 {
		// caller can check to see if count = 0
		return HistogramResult{}
	}

	sort.Float64s(he.samples)
	count := len(he.samples)
	sum := 0.0
	for _, val := range he.samples {
		sum += val
	}

	return HistogramResult{
		count:  float64(count),
		min:    he.samples[0],
		max:    he.samples[count-1],
		avg:    sum / float64(count),
		median: he.samples[count/2],
		p95:    he.samples[(count*95)/100],
	}
}
