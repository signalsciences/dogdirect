package dogdirect

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var hostTaggerCases = []struct {
	hostname string
	tags     []string
}{
	{"hostname", []string{"tag1", "tag2", "tag3"}},
	{"hostname", []string{}},
}

func TestHostTagger(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		want := `{"tags":["tag1","tag2","tag3"]}`
		if string(got) != want {
			t.Errorf("got %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	endpointv1 = ts.URL
	api := NewAPI("foo", "bar", 0)

	for _, tt := range hostTaggerCases {
		ht := NewHostTagger(api, tt.hostname, tt.tags)

		if err := ht.Flush(); err != nil {
			t.Fatal(err)
		}

		// Flush() twice to cover the case of returning nil
		// when ht.tagged is already set
		if err := ht.Flush(); err != nil {
			t.Fatal(err)
		}

		if err := ht.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
