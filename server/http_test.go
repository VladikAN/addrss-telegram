package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vladikan/addrss-telegram/templates"
)

func TestCommandHandler_Start(t *testing.T) {
	templates.SetCustomOutput(func(lang string, name string, data interface{}) (string, error) {
		return name, nil
	})

	opt := &Options{BotAdmin: 0}
	handler := makeCommandHandler(opt)

	body, _ := json.Marshal(httpRequest{UserID: 1, Text: "/start"})
	req := httptest.NewRequest(http.MethodPost, "/command", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var replies []httpReply
	json.NewDecoder(w.Body).Decode(&replies)

	if len(replies) != 1 {
		t.Fatalf("Expected 1 reply, got %d", len(replies))
	}
	if replies[0].Text != "start-success" {
		t.Errorf("Expected 'start-success', got '%s'", replies[0].Text)
	}
	if replies[0].ChatID != 1 {
		t.Errorf("Expected chat_id 1, got %d", replies[0].ChatID)
	}
}

func TestCommandHandler_MethodNotAllowed(t *testing.T) {
	opt := &Options{}
	handler := makeCommandHandler(opt)

	req := httptest.NewRequest(http.MethodGet, "/command", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestCommandHandler_InvalidJSON(t *testing.T) {
	opt := &Options{}
	handler := makeCommandHandler(opt)

	req := httptest.NewRequest(http.MethodPost, "/command", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCommandHandler_DefaultUserID(t *testing.T) {
	templates.SetCustomOutput(func(lang string, name string, data interface{}) (string, error) {
		return name, nil
	})

	opt := &Options{}
	handler := makeCommandHandler(opt)

	body, _ := json.Marshal(httpRequest{Text: "/help"})
	req := httptest.NewRequest(http.MethodPost, "/command", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	var replies []httpReply
	json.NewDecoder(w.Body).Decode(&replies)

	if len(replies) != 1 {
		t.Fatalf("Expected 1 reply, got %d", len(replies))
	}
	if replies[0].ChatID != 1 {
		t.Errorf("Expected default chat_id 1, got %d", replies[0].ChatID)
	}
}
