package dogdirect

import (
	"sort"
)

// HistogramResult returns some descriptive statistics
// Add other stats as needed
type HistogramResult struct {
	Count  float64
	Min    float64
	Max    float64
	Avg    float64
	Median float64
	P95    float64
}

// ExactHistogram is the dumbest way possible to compute various descriptive statistics
//  It keeps all data, does a sort, and the figures out various stats.
//  That said for 1000 elements, it takes under 1/20 of a millisecond to compute.
//
// Also the "sort" method is what datadog's agent does, so it can't be too painful.
//
type ExactHistogram struct {
	samples []float64
	tags    []string
}

// NewExactHistogram creates a new object
func NewExactHistogram(points int, tags []string) *ExactHistogram {
	if points == 0 {
		return &ExactHistogram{}
	}
	return &ExactHistogram{
		samples: make([]float64, 0, points),
		tags:    tags,
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
		Count:  float64(count),
		Min:    he.samples[0],
		Max:    he.samples[count-1],
		Avg:    sum / float64(count),
		Median: he.samples[count/2],
		P95:    he.samples[(count*95)/100],
	}
}
