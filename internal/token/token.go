// Package token issues and verifies stateless, HMAC-signed, expiring tokens.
package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Purpose binds a token to one action so it cannot be replayed on another.
type Purpose string

const (
	Confirm     Purpose = "confirm"
	Unsubscribe Purpose = "unsubscribe"
)

var (
	ErrMalformed = errors.New("token: malformed")
	ErrSignature = errors.New("token: bad signature")
	ErrExpired   = errors.New("token: expired")
	ErrPurpose   = errors.New("token: wrong purpose")
)

// Claims is the verified content of a token.
type Claims struct {
	Email   string
	List    string
	Purpose Purpose
	Expires time.Time
}

type wire struct {
	Email   string  `json:"e"`
	List    string  `json:"l"`
	Purpose Purpose `json:"p"`
	Expires int64   `json:"x"`
}

// Signer signs and verifies tokens with a shared secret.
type Signer struct {
	secret []byte
	now    func() time.Time
}

// NewSigner returns a Signer. A nil now defaults to time.Now.
func NewSigner(secret []byte, now func() time.Time) *Signer {
	if now == nil {
		now = time.Now
	}
	return &Signer{secret: secret, now: now}
}

// Sign returns a URL-safe token for c.
func (s *Signer) Sign(c Claims) string {
	payload, _ := json.Marshal(wire{c.Email, c.List, c.Purpose, c.Expires.Unix()}) //nolint:errchkjson // fixed struct cannot fail
	enc := base64.RawURLEncoding
	body := enc.EncodeToString(payload)
	return body + "." + enc.EncodeToString(s.mac(body))
}

// Verify checks signature, purpose and expiry, and returns the claims.
func (s *Signer) Verify(tok string, want Purpose) (Claims, error) {
	body, sig, ok := strings.Cut(tok, ".")
	if !ok {
		return Claims{}, ErrMalformed
	}
	enc := base64.RawURLEncoding
	gotSig, err := enc.DecodeString(sig)
	if err != nil {
		return Claims{}, ErrMalformed
	}
	if !hmac.Equal(gotSig, s.mac(body)) {
		return Claims{}, ErrSignature
	}
	payload, err := enc.DecodeString(body)
	if err != nil {
		return Claims{}, ErrMalformed
	}
	var w wire
	if err := json.Unmarshal(payload, &w); err != nil {
		return Claims{}, ErrMalformed
	}
	if w.Purpose != want {
		return Claims{}, ErrPurpose
	}
	c := Claims{Email: w.Email, List: w.List, Purpose: w.Purpose, Expires: time.Unix(w.Expires, 0)}
	if !s.now().Before(c.Expires) {
		return Claims{}, ErrExpired
	}
	return c, nil
}

func (s *Signer) mac(body string) []byte {
	h := hmac.New(sha256.New, s.secret)
	h.Write([]byte(body))
	return h.Sum(nil)
}
