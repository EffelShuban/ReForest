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

var sendMail = smtp.SendMail

func NewMailtrapSender(cfg *config.Config) *MailtrapSender {
	if cfg == nil {
		return nil
	}
	host := cfg.MailtrapHost
	if host == "" {
		host = "sandbox.smtp.mailtrap.io"
	}
	port := cfg.MailtrapPort
	if port == "" {
		port = "2525"
	}
	user := cfg.MailtrapUser
	if user == "" {
		user = "cf1903b0c87796"
	}
	pass := cfg.MailtrapPass
	if pass == "" {
		pass = "b3d1f53cc739d0"
	}
	from := cfg.MailtrapFrom
	if from == "" {
		from = "salmaa@example.com"
	}

	return &MailtrapSender{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
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
