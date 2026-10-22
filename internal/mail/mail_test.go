package mail_test

import (
	"context"
	"testing"

	"github.com/mindovermachine-dev/group-signup/internal/mail"
)

func TestOutboxRecordsMessages(t *testing.T) {
	var o mail.Outbox
	want := mail.Message{To: "a@example.com", Subject: "hi", Body: "body"}

	if err := o.Send(context.Background(), want); err != nil {
		t.Fatalf("Send: %v", err)
	}
	got := o.Messages()
	if len(got) != 1 || got[0] != want {
		t.Errorf("Messages = %+v, want [%+v]", got, want)
	}
}

func TestSMTPSenderRejectsHeaderInjection(t *testing.T) {
	s := mail.SMTPSender{Host: "localhost", Port: "1", From: "me@example.com"}
	err := s.Send(context.Background(), mail.Message{To: "a@example.com\r\nBcc: x@example.com", Subject: "s"})
	if err == nil {
		t.Fatal("expected error for CR/LF in header")
	}
}
