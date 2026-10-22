// Package newsletter implements double opt-in newsletter subscriptions.
package newsletter

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"slices"
	"strings"
	"time"

	postmail "github.com/mindovermachine-dev/group-signup/internal/mail"
	"github.com/mindovermachine-dev/group-signup/internal/store"
	"github.com/mindovermachine-dev/group-signup/internal/token"
)

var (
	ErrInvalidEmail = errors.New("newsletter: invalid email")
	ErrUnknownList  = errors.New("newsletter: unknown list")
	ErrInvalidToken = errors.New("newsletter: invalid token")
)

// Config holds the service settings.
type Config struct {
	BaseURL    string        // public URL of this service, used in email links
	Lists      []string      // accepted list names
	ConfirmTTL time.Duration // validity of a confirmation link
	Now        func() time.Time
}

// Service holds the subscription rules.
type Service struct {
	cfg    Config
	store  store.Store
	mailer postmail.Sender
	tokens *token.Signer
}

// New returns a Service. A zero ConfirmTTL defaults to 24h and a nil Now to time.Now.
func New(cfg Config, st store.Store, m postmail.Sender, tk *token.Signer) *Service {
	if cfg.ConfirmTTL == 0 {
		cfg.ConfirmTTL = 24 * time.Hour
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &Service{cfg: cfg, store: st, mailer: m, tokens: tk}
}

// Subscribe records a pending subscription and emails a confirmation link.
// Already-confirmed addresses are silently ignored so callers cannot probe the list.
func (s *Service) Subscribe(ctx context.Context, email, list string) error {
	email, err := normalizeEmail(email)
	if err != nil {
		return err
	}
	if !slices.Contains(s.cfg.Lists, list) {
		return ErrUnknownList
	}

	sub, err := s.store.Get(ctx, list, email)
	switch {
	case err == nil && sub.Status == store.Confirmed:
		return nil
	case err != nil && !errors.Is(err, store.ErrNotFound):
		return err
	case err != nil:
		sub = store.Subscriber{Email: email, List: list, CreatedAt: s.cfg.Now()}
	}
	sub.Status = store.Pending
	if err := s.store.Put(ctx, sub); err != nil {
		return err
	}

	tok := s.tokens.Sign(token.Claims{
		Email: email, List: list, Purpose: token.Confirm,
		Expires: s.cfg.Now().Add(s.cfg.ConfirmTTL),
	})
	link := s.cfg.BaseURL + "/v1/confirm?token=" + url.QueryEscape(tok)
	return s.mailer.Send(ctx, postmail.Message{
		To:      email,
		Subject: "Please confirm your subscription",
		Body: fmt.Sprintf("Confirm your subscription to %q by opening this link:\n\n%s\n\n"+
			"If you did not request this, ignore this email.\n", list, link),
	})
}

// Confirm activates the subscription named by a confirmation token.
func (s *Service) Confirm(ctx context.Context, tok string) error {
	c, err := s.tokens.Verify(tok, token.Confirm)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	sub, err := s.store.Get(ctx, c.List, c.Email)
	if errors.Is(err, store.ErrNotFound) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	switch sub.Status {
	case store.Confirmed:
		return nil
	case store.Unsubscribed:
		return ErrInvalidToken
	}
	sub.Status = store.Confirmed
	sub.ConfirmedAt = s.cfg.Now()
	return s.store.Put(ctx, sub)
}

// Unsubscribe deactivates the subscription named by an unsubscribe token.
func (s *Service) Unsubscribe(ctx context.Context, tok string) error {
	c, err := s.tokens.Verify(tok, token.Unsubscribe)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	sub, err := s.store.Get(ctx, c.List, c.Email)
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	sub.Status = store.Unsubscribed
	return s.store.Put(ctx, sub)
}

// UnsubscribeURL returns a long-lived opt-out link for use in outgoing newsletters.
func (s *Service) UnsubscribeURL(email, list string) string {
	tok := s.tokens.Sign(token.Claims{
		Email: email, List: list, Purpose: token.Unsubscribe,
		Expires: s.cfg.Now().AddDate(10, 0, 0),
	})
	return s.cfg.BaseURL + "/v1/unsubscribe?token=" + url.QueryEscape(tok)
}

func normalizeEmail(raw string) (string, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	addr, err := mail.ParseAddress(raw)
	if err != nil || addr.Address != raw {
		return "", ErrInvalidEmail
	}
	return raw, nil
}
