package mail

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestBuildFormatsMessage(t *testing.T) {
	s := SMTPSender{From: "me@example.com"}

	raw, err := s.build(Message{To: "a@example.com", Subject: "Hej æøå", Body: "line1\nline2"})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	got := string(raw)

	for _, want := range []string{
		"From: me@example.com\r\n",
		"To: a@example.com\r\n",
		"Subject: =?utf-8?q?",
		"Content-Type: text/plain; charset=utf-8\r\n\r\nline1\r\nline2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("message missing %q:\n%s", want, got)
		}
	}
}

func TestLogSenderNeverFails(t *testing.T) {
	l := LogSender{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := l.Send(context.Background(), Message{To: "a@example.com"}); err != nil {
		t.Errorf("Send: %v", err)
	}
}
