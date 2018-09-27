package dogdirect

import (
	"reflect"
	"sort"
	"testing"
)

var cases = []struct {
	orig []string
	want []string
}{
	{nil, nil},
	{[]string{"foo"}, []string{"foo"}},
	{[]string{"foo", "bar"}, []string{"bar", "foo"}},
	{[]string{"bar", "foo"}, []string{"bar", "foo"}},
	{[]string{"foo", "foo"}, []string{"foo"}},
	{[]string{"foo", "bar", "foo"}, []string{"bar", "foo"}},
	{[]string{"foo", "foo", "bar"}, []string{"bar", "foo"}},
	{[]string{"bar", "foo", "foo"}, []string{"bar", "foo"}},
	{[]string{"foo", "foo", "foo"}, []string{"foo"}},
}

func TestUnique(t *testing.T) {
	for i, c := range cases {
		got := unique(c.orig)
		sort.Strings(got)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Case %d: got %v want %v", i, got, c.want)
		}
	}
}

var fruitlist = []string{
	"apricot",
	"fig",
	"avocado",
	"apple",
	"banana",
	"lime",
	"date",
	"lemon",
}

// for comparison
func unique_map(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

func BenchmarkUniqMap(b *testing.B) {
	fruits := fruitlist
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		unique_map(fruits)
	}
}

func BenchmarkUniqSort(b *testing.B) {
	fruits := fruitlist
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		unique(fruits)
	}
}
