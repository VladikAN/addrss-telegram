package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/vladikan/addrss-telegram/database"
)

const (
	defaultFeedsLimit = 50
	maxFeedsLimit     = 200
)

type feedResponse struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Normalized string     `json:"normalized"`
	URI        string     `json:"uri"`
	Updated    *time.Time `json:"updated,omitempty"`
	Healthy    bool       `json:"healthy"`
	Blocked    bool       `json:"blocked"`
	LastPub    *time.Time `json:"last_pub,omitempty"`
	LastPubURI string     `json:"last_pub_uri,omitempty"`
}

type feedsListResponse struct {
	Feeds  []feedResponse `json:"feeds"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
	Total  int            `json:"total"`
}

func registerValidationAPI(mux *http.ServeMux, opt *Options) {
	mux.Handle("/api/feeds", authMiddleware(opt.APIToken, http.HandlerFunc(listFeedsHandler)))
	mux.Handle("/api/feeds/", authMiddleware(opt.APIToken, http.HandlerFunc(feedActionHandler)))
}

func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			http.Error(w, "api token is not configured", http.StatusServiceUnavailable)
			return
		}

		if !authorizeRequest(r, token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authorizeRequest(r *http.Request, token string) bool {
	if header := r.Header.Get("X-API-Token"); header != "" {
		return header == token
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:]) == token
	}

	return false
}

func listFeedsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	offset, err := parseNonNegativeInt(r.URL.Query().Get("offset"), 0)
	if err != nil {
		http.Error(w, "invalid offset", http.StatusBadRequest)
		return
	}

	limit, err := parseNonNegativeInt(r.URL.Query().Get("limit"), defaultFeedsLimit)
	if err != nil || limit == 0 {
		http.Error(w, "invalid limit", http.StatusBadRequest)
		return
	}
	if limit > maxFeedsLimit {
		limit = maxFeedsLimit
	}

	feeds, total, err := db.GetFeedsPage(offset, limit)
	if err != nil {
		log.Printf("ERROR api list feeds: %s", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	out := feedsListResponse{
		Feeds:  toFeedResponses(feeds),
		Offset: offset,
		Limit:  limit,
		Total:  total,
	}

	writeJSON(w, http.StatusOK, out)
}

func feedActionHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/feeds/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		http.Error(w, "invalid feed id", http.StatusBadRequest)
		return
	}

	switch parts[1] {
	case "block":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		setFeedBlockedHandler(w, id, true)
	case "unblock":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		setFeedBlockedHandler(w, id, false)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func setFeedBlockedHandler(w http.ResponseWriter, id int, blocked bool) {
	feed, err := db.GetFeedByID(id)
	if err != nil {
		log.Printf("ERROR api get feed %d: %s", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if feed == nil {
		http.Error(w, "feed not found", http.StatusNotFound)
		return
	}

	if err := db.SetFeedBlocked(id, blocked); err != nil {
		log.Printf("ERROR api set feed %d blocked=%v: %s", id, blocked, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	feed.Blocked = blocked
	writeJSON(w, http.StatusOK, toFeedResponse(*feed))
}

func toFeedResponses(feeds []database.Feed) []feedResponse {
	out := make([]feedResponse, 0, len(feeds))
	for _, feed := range feeds {
		out = append(out, toFeedResponse(feed))
	}
	return out
}

func toFeedResponse(feed database.Feed) feedResponse {
	return feedResponse{
		ID:         feed.ID,
		Name:       feed.Name,
		Normalized: feed.Normalized,
		URI:        feed.URI,
		Updated:    feed.Updated,
		Healthy:    feed.Healthy,
		Blocked:    feed.Blocked,
		LastPub:    feed.LastPub,
		LastPubURI: feed.LastPubURI,
	}
}

func parseNonNegativeInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
