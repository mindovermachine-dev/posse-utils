// Package mail sends transactional email.
package mail

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"sync"
)

// Message is a plain-text email.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender delivers messages.
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// SMTPSender sends through an authenticated SMTP submission server
// (for example smtp.gmail.com:587 with a Google Workspace app password).
type SMTPSender struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func (s SMTPSender) Send(_ context.Context, m Message) error {
	msg, err := s.build(m)
	if err != nil {
		return err
	}
	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	if err := smtp.SendMail(net.JoinHostPort(s.Host, s.Port), auth, s.From, []string{m.To}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func (s SMTPSender) build(m Message) ([]byte, error) {
	if strings.ContainsAny(m.To+m.Subject+s.From, "\r\n") {
		return nil, errors.New("mail: header injection attempt")
	}
	var b strings.Builder
	b.WriteString("From: " + s.From + "\r\n")
	b.WriteString("To: " + m.To + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", m.Subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(strings.ReplaceAll(m.Body, "\n", "\r\n"))
	return []byte(b.String()), nil
}

// LogSender writes messages to the log; for local development only.
type LogSender struct{ Log *slog.Logger }

func (l LogSender) Send(ctx context.Context, m Message) error {
	l.Log.InfoContext(ctx, "mail (not sent)", "to", m.To, "subject", m.Subject, "body", m.Body)
	return nil
}

// Outbox records messages instead of sending them; a fake for tests.
type Outbox struct {
	mu   sync.Mutex
	msgs []Message
}

func (o *Outbox) Send(_ context.Context, m Message) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.msgs = append(o.msgs, m)
	return nil
}

// Messages returns a copy of everything sent so far.
func (o *Outbox) Messages() []Message {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]Message(nil), o.msgs...)
}
