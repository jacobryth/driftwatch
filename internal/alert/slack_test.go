package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSlackHandler_Success(t *testing.T) {
	var received slackPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := NewSlackHandler(ts.URL)
	e := makeEvent(KindModified, "/etc/app/config.yaml")

	if err := h.Handle(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Text == "" {
		t.Error("expected non-empty Slack message text")
	}
	if len(received.Attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
	attachment := received.Attachments[0]
	if attachment.Color != "#ff0000" {
		t.Errorf("expected critical color, got %s", attachment.Color)
	}
	foundPath := false
	for _, f := range attachment.Fields {
		if f.Title == "File" && f.Value == "/etc/app/config.yaml" {
			foundPath = true
		}
	}
	if !foundPath {
		t.Error("expected File field with correct path in attachment")
	}
}

func TestSlackHandler_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := NewSlackHandler(ts.URL)
	e := makeEvent(KindModified, "/etc/app/config.yaml")

	if err := h.Handle(e); err == nil {
		t.Error("expected error for non-OK status, got nil")
	}
}

func TestSlackHandler_Unreachable(t *testing.T) {
	h := NewSlackHandler("http://127.0.0.1:0/nonexistent")
	e := makeEvent(KindDeleted, "/etc/app/config.yaml")

	if err := h.Handle(e); err == nil {
		t.Error("expected error for unreachable server, got nil")
	}
}

func TestSlackHandler_WarningColor(t *testing.T) {
	var received slackPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := NewSlackHandler(ts.URL)
	e := makeEvent(KindUnknown, "/etc/app/config.yaml")

	if err := h.Handle(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(received.Attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
	if received.Attachments[0].Color != "#ffcc00" {
		t.Errorf("expected warning color #ffcc00, got %s", received.Attachments[0].Color)
	}
}
