package youtube

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientLookup_ViaResolveChannel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/youtube/v3/channels" && r.URL.Path != "/" {
			// allow any path; we'll point client HTTP to rewrite
		}
		q := r.URL.Query()
		if q.Get("forHandle") != "veritasium" {
			t.Fatalf("forHandle=%q", q.Get("forHandle"))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"id": "UCHnyfMqiRRG1u-2MsSQLbXA", "snippet": map[string]string{"title": "Veritasium"}},
			},
		})
	}))
	defer srv.Close()

	client := NewClient("k")
	client.HTTP = srv.Client()

	// Monkey-patch by calling lookup through a thin wrapper: replace api host using custom RoundTripper.
	client.HTTP.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = srv.Listener.Addr().String()
		req.URL.Path = "/"
		return http.DefaultTransport.RoundTrip(req)
	})

	ch, err := client.ResolveChannel("https://www.youtube.com/@veritasium")
	if err != nil {
		t.Fatalf("ResolveChannel: %v", err)
	}
	if ch.Title != "Veritasium" || ch.ID != "UCHnyfMqiRRG1u-2MsSQLbXA" {
		t.Fatalf("unexpected: %+v", ch)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
