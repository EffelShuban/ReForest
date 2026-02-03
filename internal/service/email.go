package service

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"reforest/config"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type MailtrapSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// sendMail is kept as a variable so tests can stub it without opening real network connections.
var sendMail = smtp.SendMail

func NewMailtrapSender(cfg *config.Config) *MailtrapSender {
	if cfg == nil {
		return nil
	}
	return &MailtrapSender{
		Host:     cfg.MailtrapHost,
		Port:     cfg.MailtrapPort,
		Username: cfg.MailtrapUser,
		Password: cfg.MailtrapPass,
		From:     cfg.MailtrapFrom,
	}
}

func (s *MailtrapSender) Send(to, subject, body string) error {
	if err := s.validate(to, subject, body); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s\r\n",
		s.From, to, subject, body))

	return sendMail(addr, auth, s.From, []string{to}, msg)
}

func (s *MailtrapSender) validate(to, subject, body string) error {
	if s == nil {
		return errors.New("email sender is nil")
	}
	if s.Host == "" || s.Port == "" || s.Username == "" || s.Password == "" || s.From == "" {
		return errors.New("mailtrap configuration is incomplete")
	}
	if to == "" {
		return errors.New("recipient email is required")
	}
	if !strings.Contains(to, "@") {
		return errors.New("recipient email looks invalid")
	}
	if subject == "" {
		return errors.New("subject is required")
	}
	if body == "" {
		return errors.New("email body is empty")
	}
	return nil
}
