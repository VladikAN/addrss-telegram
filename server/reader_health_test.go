package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vladikan/addrss-telegram/database"
)

func TestReadFeeds_ClientErrorsMarkBroken(t *testing.T) {
	for _, code := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()

			broken := false
			updated := false
			now := time.Now()
			rd := &Reader{DB: &dbMock{
				getFeedsMock: func() ([]database.Feed, error) {
					return []database.Feed{{
						ID:         1,
						Normalized: "broken",
						URI:        srv.URL,
						SourceType: database.SourceRSS,
						LastPub:    &now,
						Healthy:    true,
					}}, nil
				},
				setFeedBrokenMock:  func() error { broken = true; return nil },
				setFeedUpdatedMock: func() error { updated = true; return nil },
			}}

			if err := rd.readFeeds(); err != nil {
				t.Fatal(err)
			}
			if !broken {
				t.Fatal("expected feed to be marked broken")
			}
			if updated {
				t.Fatal("expected feed not to be marked healthy")
			}
		})
	}
}

func TestReadFeeds_TLSErrorMarksBroken(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	broken := false
	updated := false
	now := time.Now()
	rd := &Reader{DB: &dbMock{
		getFeedsMock: func() ([]database.Feed, error) {
			return []database.Feed{{
				ID:         2,
				Normalized: "tls",
				URI:        srv.URL,
				SourceType: database.SourceRSS,
				LastPub:    &now,
			}}, nil
		},
		setFeedBrokenMock:  func() error { broken = true; return nil },
		setFeedUpdatedMock: func() error { updated = true; return nil },
	}}

	if err := rd.readFeeds(); err != nil {
		t.Fatal(err)
	}
	if !broken || updated {
		t.Fatalf("broken=%v updated=%v", broken, updated)
	}
}
