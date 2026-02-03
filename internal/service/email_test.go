package service

import (
	"net/smtp"
	"strings"
	"testing"

	"reforest/config"
)

func TestNewMailtrapSender_NilConfig(t *testing.T) {
	if got := NewMailtrapSender(nil); got != nil {
		t.Fatalf("expected nil sender for nil config")
	}
}

func TestNewMailtrapSender_FillsFields(t *testing.T) {
	cfg := &config.Config{
		MailtrapHost: "smtp.example.com",
		MailtrapPort: "2525",
		MailtrapUser: "u",
		MailtrapPass: "p",
		MailtrapFrom: "from@example.com",
	}

	s := NewMailtrapSender(cfg)
	if s.Host != cfg.MailtrapHost || s.Port != cfg.MailtrapPort || s.Username != cfg.MailtrapUser ||
		s.Password != cfg.MailtrapPass || s.From != cfg.MailtrapFrom {
		t.Fatalf("sender fields not populated from config: %+v", s)
	}
}

func TestMailtrapSender_validateErrors(t *testing.T) {
	tests := []struct {
		name string
		s    *MailtrapSender
		to   string
		sub  string
		body string
	}{
		{"nil receiver", nil, "a@b.com", "sub", "body"},
		{"missing config", &MailtrapSender{}, "a@b.com", "sub", "body"},
		{"missing recipient", &MailtrapSender{Host: "h", Port: "p", Username: "u", Password: "p", From: "f"}, "", "sub", "body"},
		{"invalid recipient", &MailtrapSender{Host: "h", Port: "p", Username: "u", Password: "p", From: "f"}, "invalid-email", "sub", "body"},
		{"missing subject", &MailtrapSender{Host: "h", Port: "p", Username: "u", Password: "p", From: "f"}, "a@b.com", "", "body"},
		{"missing body", &MailtrapSender{Host: "h", Port: "p", Username: "u", Password: "p", From: "f"}, "a@b.com", "sub", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.validate(tt.to, tt.sub, tt.body)
			if err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestMailtrapSender_SendSuccess(t *testing.T) {
	origSend := sendMail
	defer func() { sendMail = origSend }()

	var capturedAddr, capturedFrom, capturedMsg string
	var capturedTo []string

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedAddr = addr
		capturedFrom = from
		capturedTo = append([]string{}, to...)
		capturedMsg = string(msg)
		return nil
	}

	s := &MailtrapSender{
		Host:     "smtp.test",
		Port:     "2525",
		Username: "user",
		Password: "pass",
		From:     "from@example.com",
	}

	err := s.Send("to@example.com", "Hello", "Test body")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if capturedAddr != "smtp.test:2525" {
		t.Fatalf("addr passed to SendMail mismatch, got %q", capturedAddr)
	}
	if capturedFrom != "from@example.com" {
		t.Fatalf("MAIL FROM mismatch, got %q", capturedFrom)
	}
	if len(capturedTo) != 1 || capturedTo[0] != "to@example.com" {
		t.Fatalf("RCPT TO mismatch, got %v", capturedTo)
	}
	if !strings.Contains(capturedMsg, "Subject: Hello") {
		t.Fatalf("message subject not found in payload: %q", capturedMsg)
	}
	if !strings.Contains(capturedMsg, "Test body") {
		t.Fatalf("message body not found in payload: %q", capturedMsg)
	}
}

func TestMailtrapSender_SendValidationError(t *testing.T) {
	s := &MailtrapSender{Host: "h", Port: "p", Username: "u", Password: "p", From: "f"}
	if err := s.Send("", "sub", "body"); err == nil {
		t.Fatalf("expected validation error when recipient is empty")
	}
}
