package alert

import (
	"errors"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func newEmailHandler(dialFn func(string, smtp.Auth, string, []string, []byte) error) *emailHandler {
	return &emailHandler{
		cfg: EmailConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "secret",
			From:     "alerts@example.com",
			To:       []string{"ops@example.com"},
		},
		dial: dialFn,
	}
}

func TestEmailHandler_Success(t *testing.T) {
	var capturedMsg []byte
	h := newEmailHandler(func(_ string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedMsg = msg
		return nil
	})

	evt := makeEvent(KindModified, "critical")
	evt.Time = time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	if err := h.Handle(evt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(capturedMsg)
	if !strings.Contains(s, "Subject: [driftwatch]") {
		t.Error("expected Subject header in message")
	}
	if !strings.Contains(s, evt.Path) {
		t.Errorf("expected path %q in message body", evt.Path)
	}
}

func TestEmailHandler_AddrFormat(t *testing.T) {
	var capturedAddr string
	h := newEmailHandler(func(addr string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		capturedAddr = addr
		return nil
	})

	_ = h.Handle(makeEvent(KindModified, "warning"))

	if capturedAddr != "smtp.example.com:587" {
		t.Errorf("unexpected addr %q", capturedAddr)
	}
}

func TestEmailHandler_PropagatesError(t *testing.T) {
	sentinel := errors.New("smtp unavailable")
	h := newEmailHandler(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		return sentinel
	})

	if err := h.Handle(makeEvent(KindModified, "critical")); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestEmailHandler_NoAuthWhenNoUsername(t *testing.T) {
	var capturedAuth smtp.Auth
	h := &emailHandler{
		cfg: EmailConfig{
			Host: "localhost",
			Port: 25,
			From: "a@b.com",
			To:   []string{"c@d.com"},
		},
		dial: func(_ string, auth smtp.Auth, _ string, _ []string, _ []byte) error {
			capturedAuth = auth
			return nil
		},
	}

	_ = h.Handle(makeEvent(KindDeleted, "warning"))

	if capturedAuth != nil {
		t.Error("expected nil auth when Username is empty")
	}
}
