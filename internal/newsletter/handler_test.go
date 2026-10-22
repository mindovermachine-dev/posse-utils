package newsletter_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mindovermachine-dev/group-signup/internal/newsletter"
	"github.com/mindovermachine-dev/group-signup/internal/store"
)

const siteURL = "https://site.example.com"

func newServer(f *fixture) http.Handler {
	mux := http.NewServeMux()
	newsletter.NewHandler(f.svc, siteURL, slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return mux
}

func do(h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHTTPSubscribeAndConfirm(t *testing.T) {
	f := newFixture()
	h := newServer(f)

	rec := do(h, http.MethodPost, "/v1/subscribe", `{"email":"a@example.com","list":"news"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("subscribe code = %d, want 202", rec.Code)
	}

	rec = do(h, http.MethodGet, "/v1/confirm?token="+url.QueryEscape(f.lastToken(t)), "")
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != siteURL+"/newsletter/confirmed/" {
		t.Errorf("confirm = %d -> %q", rec.Code, rec.Header().Get("Location"))
	}
	if got := f.status(t, "a@example.com"); got != store.Confirmed {
		t.Errorf("status = %q, want confirmed", got)
	}
}

func TestHTTPSubscribeRejects(t *testing.T) {
	tests := []struct {
		name, body string
		want       int
	}{
		{"malformed json", `{`, http.StatusBadRequest},
		{"bad email", `{"email":"x","list":"news"}`, http.StatusBadRequest},
		{"unknown list", `{"email":"a@example.com","list":"zzz"}`, http.StatusBadRequest},
		{"oversized", `{"email":"` + strings.Repeat("a", 8<<10) + `"}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			if rec := do(newServer(f), http.MethodPost, "/v1/subscribe", tt.body); rec.Code != tt.want {
				t.Errorf("code = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestHTTPHoneypotIsSilentlyDropped(t *testing.T) {
	f := newFixture()
	rec := do(newServer(f), http.MethodPost, "/v1/subscribe",
		`{"email":"a@example.com","list":"news","website":"http://spam"}`)

	if rec.Code != http.StatusAccepted {
		t.Errorf("code = %d, want 202", rec.Code)
	}
	if n := len(f.outbox.Messages()); n != 0 {
		t.Errorf("emails sent = %d, want 0", n)
	}
}

func TestHTTPInvalidTokenRedirects(t *testing.T) {
	h := newServer(newFixture())
	for _, path := range []string{"/v1/confirm?token=bad", "/v1/unsubscribe?token=bad"} {
		rec := do(h, http.MethodGet, path, "")
		if got := rec.Header().Get("Location"); rec.Code != http.StatusSeeOther || got != siteURL+"/newsletter/invalid/" {
			t.Errorf("%s = %d -> %q", path, rec.Code, got)
		}
	}
}

func TestHTTPUnsubscribe(t *testing.T) {
	f := newFixture()
	h := newServer(f)
	_ = f.svc.Subscribe(context.Background(), "a@example.com", "news")

	rec := do(h, http.MethodGet, strings.TrimPrefix(f.svc.UnsubscribeURL("a@example.com", "news"), baseURL), "")
	if got := rec.Header().Get("Location"); got != siteURL+"/newsletter/unsubscribed/" {
		t.Errorf("Location = %q", got)
	}
	if got := f.status(t, "a@example.com"); got != store.Unsubscribed {
		t.Errorf("status = %q, want unsubscribed", got)
	}
}
