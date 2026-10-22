package server_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mindovermachine-dev/group-signup/internal/mail"
	"github.com/mindovermachine-dev/group-signup/internal/newsletter"
	"github.com/mindovermachine-dev/group-signup/internal/server"
	"github.com/mindovermachine-dev/group-signup/internal/store"
	"github.com/mindovermachine-dev/group-signup/internal/token"
)

const site = "https://example.com"

func newHandler() http.Handler {
	svc := newsletter.New(
		newsletter.Config{BaseURL: "https://api.example.com", Lists: []string{"news"}},
		store.NewMemory(), &mail.Outbox{}, token.NewSigner([]byte("secret"), nil),
	)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return server.New(site, newsletter.NewHandler(svc, site, log))
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("code = %d, want 200", rec.Code)
	}
}

func TestCORSPreflightAllowedOrigin(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/v1/subscribe", nil)
	req.Header.Set("Origin", site)
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("code = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != site {
		t.Errorf("Allow-Origin = %q, want %q", got, site)
	}
}

func TestCORSForeignOriginGetsNoHeaders(t *testing.T) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodOptions, "/v1/subscribe", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty", got)
	}
}
