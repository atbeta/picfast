package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	stdmail "net/mail"
	"net/smtp"
	"strings"

	"github.com/atbeta/picfast/internal/config"
)

type smtpSender struct {
	cfg *config.MailConfig
}

func NewSender(cfg *config.MailConfig) Sender {
	if cfg == nil || !cfg.IsConfigured() {
		return NewNoopSender(false)
	}
	return &smtpSender{cfg: cfg}
}

func (s *smtpSender) Ready() bool {
	return s != nil && s.cfg != nil && s.cfg.IsConfigured()
}

func (s *smtpSender) Send(ctx context.Context, msg Message) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	payload := s.buildMessage(msg)

	switch strings.ToLower(strings.TrimSpace(s.cfg.Encryption)) {
	case "tls":
		return s.sendWithImplicitTLS(ctx, addr, msg.ToEmail, payload)
	default:
		return s.sendWithSMTP(addr, msg.ToEmail, payload)
	}
}

func (s *smtpSender) sendWithSMTP(addr, toEmail string, payload []byte) error {
	var auth smtp.Auth
	if strings.TrimSpace(s.cfg.Username) != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	return smtp.SendMail(addr, auth, s.cfg.FromEmail, []string{toEmail}, payload)
}

func (s *smtpSender) sendWithImplicitTLS(ctx context.Context, addr, toEmail string, payload []byte) error {
	dialer := &tls.Dialer{
		Config: &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		},
	}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if strings.TrimSpace(s.cfg.Username) != "" {
		if err := client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return err
		}
	}
	if err := client.Mail(s.cfg.FromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}
	wc, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(payload); err != nil {
		_ = wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (s *smtpSender) buildMessage(msg Message) []byte {
	headers := []string{
		fmt.Sprintf("From: %s", formatAddress(s.cfg.FromName, s.cfg.FromEmail)),
		fmt.Sprintf("To: %s", formatAddress(msg.ToName, msg.ToEmail)),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", msg.Subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		msg.Text,
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func formatAddress(name, email string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return (&stdmail.Address{Name: name, Address: email}).String()
}

var _ Sender = (*smtpSender)(nil)
