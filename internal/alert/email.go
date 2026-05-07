package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP connection and addressing settings.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// emailHandler sends alert events via SMTP.
type emailHandler struct {
	cfg  EmailConfig
	dial func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailHandler returns a Handler that delivers alerts by email.
func NewEmailHandler(cfg EmailConfig) Handler {
	return &emailHandler{
		cfg:  cfg,
		dial: smtp.SendMail,
	}
}

func (h *emailHandler) Handle(evt Event) error {
	addr := fmt.Sprintf("%s:%d", h.cfg.Host, h.cfg.Port)
	var auth smtp.Auth
	if h.cfg.Username != "" {
		auth = smtp.PlainAuth("", h.cfg.Username, h.cfg.Password, h.cfg.Host)
	}

	subject := fmt.Sprintf("[driftwatch] %s – %s", evt.Severity, evt.Path)
	body := fmt.Sprintf(
		"Path:     %s\nEvent:    %s\nSeverity: %s\nTime:     %s\n",
		evt.Path, evt.Kind, evt.Severity, evt.Time.Format("2006-01-02T15:04:05Z07:00"),
	)
	msg := []byte(strings.Join([]string{
		"From: " + h.cfg.From,
		"To: " + strings.Join(h.cfg.To, ", "),
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		body,
	}, "\r\n"))

	return h.dial(addr, auth, h.cfg.From, h.cfg.To, msg)
}
