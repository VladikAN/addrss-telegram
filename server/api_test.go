package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vladikan/addrss-telegram/database"
)

func TestAuthMiddleware_MissingTokenConfig(t *testing.T) {
	handler := authMiddleware("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_BearerToken(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_XAPIToken(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	req.Header.Set("X-API-Token", "secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestListFeedsHandler_Success(t *testing.T) {
	db = &dbMock{
		getFeedsPageMock: func() ([]database.Feed, int, error) {
			return []database.Feed{
				{ID: 1, Name: "Feed A", URI: "https://a.example/rss", Healthy: true, Blocked: false},
				{ID: 2, Name: "Feed B", URI: "https://b.example/rss", Healthy: false, Blocked: true},
			}, 10, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/feeds?offset=0&limit=2", nil)
	w := httptest.NewRecorder()
	listFeedsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp feedsListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Total != 10 {
		t.Errorf("Expected total 10, got %d", resp.Total)
	}
	if resp.Offset != 0 || resp.Limit != 2 {
		t.Errorf("Unexpected pagination: offset=%d limit=%d", resp.Offset, resp.Limit)
	}
	if len(resp.Feeds) != 2 {
		t.Fatalf("Expected 2 feeds, got %d", len(resp.Feeds))
	}
	if !resp.Feeds[1].Blocked {
		t.Error("Expected second feed to be blocked")
	}
}

func TestListFeedsHandler_InvalidOffset(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/feeds?offset=-1", nil)
	w := httptest.NewRecorder()
	listFeedsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestBlockFeedHandler_Success(t *testing.T) {
	blocked := false
	db = &dbMock{
		getFeedByIDMock: func() (*database.Feed, error) {
			return &database.Feed{ID: 7, Name: "Feed", URI: "https://example.com/rss", Healthy: true}, nil
		},
		setFeedBlockedMock: func() error {
			blocked = true
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/7/block", nil)
	w := httptest.NewRecorder()
	feedActionHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}
	if !blocked {
		t.Error("Expected SetFeedBlocked to be called")
	}

	var resp feedResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if !resp.Blocked || resp.ID != 7 {
		t.Errorf("Unexpected response: %+v", resp)
	}
}

func TestBlockFeedHandler_NotFound(t *testing.T) {
	db = &dbMock{
		getFeedByIDMock: func() (*database.Feed, error) { return nil, nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/99/block", nil)
	w := httptest.NewRecorder()
	feedActionHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUnblockFeedHandler_Success(t *testing.T) {
	db = &dbMock{
		getFeedByIDMock: func() (*database.Feed, error) {
			return &database.Feed{ID: 3, Name: "Feed", URI: "https://example.com/rss", Blocked: true}, nil
		},
		setFeedBlockedMock: func() error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/3/unblock", nil)
	w := httptest.NewRecorder()
	feedActionHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var resp feedResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Blocked {
		t.Error("Expected feed to be unblocked")
	}
}

func TestRegisterValidationAPI_Routes(t *testing.T) {
	db = &dbMock{
		getFeedsPageMock: func() ([]database.Feed, int, error) {
			return []database.Feed{}, 0, nil
		},
	}

	mux := http.NewServeMux()
	registerValidationAPI(mux, &Options{APIToken: "secret"})

	req := httptest.NewRequest(http.MethodGet, "/api/feeds", nil)
	req.Header.Set("X-API-Token", "secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
