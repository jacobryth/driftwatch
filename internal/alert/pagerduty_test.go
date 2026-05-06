package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPDHandler(url string) *PagerDutyHandler {
	h := NewPagerDutyHandler("test-key-123")
	h.eventsURL = url
	return h
}

func TestPagerDutyHandler_Success(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	h := newPDHandler(ts.URL)
	a := makeEvent(EventModified, "/etc/app/config.yaml")

	if err := h.Handle(a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received.RoutingKey != "test-key-123" {
		t.Errorf("routing key = %q, want %q", received.RoutingKey, "test-key-123")
	}
	if received.EventAction != "trigger" {
		t.Errorf("event_action = %q, want trigger", received.EventAction)
	}
	if received.Payload.Source != "driftwatch" {
		t.Errorf("source = %q, want driftwatch", received.Payload.Source)
	}
}

func TestPagerDutyHandler_CriticalSeverity(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	h := newPDHandler(ts.URL)
	a := makeEvent(EventDeleted, "/etc/app/secret.yaml")

	if err := h.Handle(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Payload.Severity != "critical" {
		t.Errorf("severity = %q, want critical", received.Payload.Severity)
	}
}

func TestPagerDutyHandler_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	h := newPDHandler(ts.URL)
	a := makeEvent(EventModified, "/etc/app/config.yaml")

	if err := h.Handle(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestPagerDutyHandler_Unreachable(t *testing.T) {
	h := newPDHandler("http://127.0.0.1:1")
	a := makeEvent(EventModified, "/etc/app/config.yaml")

	if err := h.Handle(a); err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}
