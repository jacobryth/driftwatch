package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/alert"
)

func TestWebhookHandler_Success(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	h := alert.NewWebhookHandler(srv.URL, 5*time.Second)
	a := makeEvent("modified")

	if err := h.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["event_type"] != "modified" {
		t.Errorf("event_type: want modified, got %s", received["event_type"])
	}
	if received["path"] == "" {
		t.Error("expected non-empty path")
	}
}

func TestWebhookHandler_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	h := alert.NewWebhookHandler(srv.URL, 5*time.Second)
	if err := h.Handle(makeEvent("modified")); err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestWebhookHandler_Unreachable(t *testing.T) {
	h := alert.NewWebhookHandler("http://127.0.0.1:1", 1*time.Second)
	if err := h.Handle(makeEvent("deleted")); err == nil {
		t.Fatal("expected error for unreachable endpoint")
	}
}
