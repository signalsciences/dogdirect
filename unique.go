package dogdirect

import (
	"sort"
)

// unique computes the slice of unique elements in-place.
// Original input is destroyed.
// PUBLIC DOMAIN
func unique(s []string) []string {
	if len(s) < 2 {
		return s
	}
	sort.Strings(s)
	j := 1
	for i := 1; i < len(s); i++ {
		if s[j-1] != s[i] {
			s[j] = s[i]
			j++
		}
	}
	return s[:j]
}
