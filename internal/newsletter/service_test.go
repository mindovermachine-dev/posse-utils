package newsletter_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mindovermachine-dev/group-signup/internal/mail"
	"github.com/mindovermachine-dev/group-signup/internal/newsletter"
	"github.com/mindovermachine-dev/group-signup/internal/store"
	"github.com/mindovermachine-dev/group-signup/internal/token"
)

const baseURL = "https://api.example.com"

type fixture struct {
	svc    *newsletter.Service
	store  *store.Memory
	outbox *mail.Outbox
	now    time.Time
}

// newFixture wires real collaborators; only I/O (store, mail, clock) is in-memory.
func newFixture() *fixture {
	f := &fixture{
		store:  store.NewMemory(),
		outbox: &mail.Outbox{},
		now:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	clock := func() time.Time { return f.now }
	f.svc = newsletter.New(
		newsletter.Config{BaseURL: baseURL, Lists: []string{"news"}, Now: clock},
		f.store, f.outbox, token.NewSigner([]byte("secret"), clock),
	)
	return f
}

// lastToken extracts the token from the most recent email's link.
func (f *fixture) lastToken(t *testing.T) string {
	t.Helper()
	msgs := f.outbox.Messages()
	if len(msgs) == 0 {
		t.Fatal("no email sent")
	}
	for _, field := range strings.Fields(msgs[len(msgs)-1].Body) {
		if strings.HasPrefix(field, baseURL) {
			u, err := url.Parse(field)
			if err != nil {
				t.Fatalf("parse link: %v", err)
			}
			return u.Query().Get("token")
		}
	}
	t.Fatal("no link in email")
	return ""
}

func (f *fixture) status(t *testing.T, email string) store.Status {
	t.Helper()
	s, err := f.store.Get(context.Background(), "news", email)
	if err != nil {
		t.Fatalf("store.Get: %v", err)
	}
	return s.Status
}

func TestSubscribeThenConfirm(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	if err := f.svc.Subscribe(ctx, "  Alice@Example.com ", "news"); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if got := f.status(t, "alice@example.com"); got != store.Pending {
		t.Fatalf("status = %q, want pending", got)
	}
	if n := len(f.outbox.Messages()); n != 1 {
		t.Fatalf("emails sent = %d, want 1", n)
	}

	if err := f.svc.Confirm(ctx, f.lastToken(t)); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got := f.status(t, "alice@example.com"); got != store.Confirmed {
		t.Errorf("status = %q, want confirmed", got)
	}
}

func TestSubscribeValidation(t *testing.T) {
	tests := []struct {
		name, email, list string
		want              error
	}{
		{"empty email", "", "news", newsletter.ErrInvalidEmail},
		{"not an email", "nope", "news", newsletter.ErrInvalidEmail},
		{"display name", "Bob <bob@example.com>", "news", newsletter.ErrInvalidEmail},
		{"unknown list", "a@example.com", "other", newsletter.ErrUnknownList},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			err := f.svc.Subscribe(context.Background(), tt.email, tt.list)
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
			if n := len(f.outbox.Messages()); n != 0 {
				t.Errorf("emails sent = %d, want 0", n)
			}
		})
	}
}

func TestSubscribeConfirmedAddressSendsNothing(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_ = f.svc.Subscribe(ctx, "a@example.com", "news")
	_ = f.svc.Confirm(ctx, f.lastToken(t))

	if err := f.svc.Subscribe(ctx, "a@example.com", "news"); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if n := len(f.outbox.Messages()); n != 1 {
		t.Errorf("emails sent = %d, want 1 (only the original)", n)
	}
	if got := f.status(t, "a@example.com"); got != store.Confirmed {
		t.Errorf("status = %q, want confirmed", got)
	}
}

func TestConfirmRejectsBadTokens(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_ = f.svc.Subscribe(ctx, "a@example.com", "news")
	tok := f.lastToken(t)

	if err := f.svc.Confirm(ctx, "garbage"); !errors.Is(err, newsletter.ErrInvalidToken) {
		t.Errorf("garbage: err = %v", err)
	}

	f.now = f.now.Add(25 * time.Hour)
	if err := f.svc.Confirm(ctx, tok); !errors.Is(err, newsletter.ErrInvalidToken) {
		t.Errorf("expired: err = %v", err)
	}
	if got := f.status(t, "a@example.com"); got != store.Pending {
		t.Errorf("status = %q, want pending", got)
	}
}

func TestUnsubscribe(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_ = f.svc.Subscribe(ctx, "a@example.com", "news")
	_ = f.svc.Confirm(ctx, f.lastToken(t))

	link, err := url.Parse(f.svc.UnsubscribeURL("a@example.com", "news"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := f.svc.Unsubscribe(ctx, link.Query().Get("token")); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if got := f.status(t, "a@example.com"); got != store.Unsubscribed {
		t.Errorf("status = %q, want unsubscribed", got)
	}
	if err := f.svc.Unsubscribe(ctx, "garbage"); !errors.Is(err, newsletter.ErrInvalidToken) {
		t.Errorf("garbage: err = %v", err)
	}
}

func TestConfirmTokenCannotUnsubscribe(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_ = f.svc.Subscribe(ctx, "a@example.com", "news")

	if err := f.svc.Unsubscribe(ctx, f.lastToken(t)); !errors.Is(err, newsletter.ErrInvalidToken) {
		t.Errorf("err = %v, want ErrInvalidToken", err)
	}
}

func TestResubscribeAfterUnsubscribe(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	_ = f.svc.Subscribe(ctx, "a@example.com", "news")
	_ = f.svc.Confirm(ctx, f.lastToken(t))
	link, _ := url.Parse(f.svc.UnsubscribeURL("a@example.com", "news"))
	_ = f.svc.Unsubscribe(ctx, link.Query().Get("token"))
	oldConfirm := f.lastToken(t)

	if err := f.svc.Confirm(ctx, oldConfirm); !errors.Is(err, newsletter.ErrInvalidToken) {
		t.Errorf("old confirm link must not resubscribe: err = %v", err)
	}
	if err := f.svc.Subscribe(ctx, "a@example.com", "news"); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	if got := f.status(t, "a@example.com"); got != store.Pending {
		t.Errorf("status = %q, want pending", got)
	}
}
