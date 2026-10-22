package token_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mindovermachine-dev/group-signup/internal/token"
)

var epoch = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func claims() token.Claims {
	return token.Claims{
		Email:   "a@example.com",
		List:    "newsletter",
		Purpose: token.Confirm,
		Expires: epoch.Add(time.Hour),
	}
}

func TestRoundTrip(t *testing.T) {
	s := token.NewSigner([]byte("secret"), func() time.Time { return epoch })

	got, err := s.Verify(s.Sign(claims()), token.Confirm)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if want := claims(); got.Email != want.Email || got.List != want.List || !got.Expires.Equal(want.Expires) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestVerifyRejects(t *testing.T) {
	now := epoch
	clock := func() time.Time { return now }
	s := token.NewSigner([]byte("secret"), clock)
	valid := s.Sign(claims())

	tests := []struct {
		name string
		tok  string
		want error
		prep func()
	}{
		{name: "no separator", tok: "garbage", want: token.ErrMalformed},
		{name: "bad base64 signature", tok: "abc.!!!", want: token.ErrMalformed},
		{name: "tampered body", tok: "x" + valid, want: token.ErrSignature},
		{
			name: "other secret",
			tok:  token.NewSigner([]byte("other"), clock).Sign(claims()),
			want: token.ErrSignature,
		},
		{
			name: "wrong purpose",
			tok:  s.Sign(token.Claims{Purpose: token.Unsubscribe, Expires: epoch.Add(time.Hour)}),
			want: token.ErrPurpose,
		},
		{
			name: "expired",
			tok:  valid,
			want: token.ErrExpired,
			prep: func() { now = epoch.Add(2 * time.Hour) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now = epoch
			if tt.prep != nil {
				tt.prep()
			}
			_, err := s.Verify(tt.tok, token.Confirm)
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTokenIsURLSafe(t *testing.T) {
	s := token.NewSigner([]byte("secret"), nil)
	tok := s.Sign(claims())
	if strings.ContainsAny(tok, "+/= ") {
		t.Errorf("token %q is not URL-safe", tok)
	}
}
