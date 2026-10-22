// Command server runs the posse-utils HTTP API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mindovermachine-dev/group-signup/internal/config"
	"github.com/mindovermachine-dev/group-signup/internal/mail"
	"github.com/mindovermachine-dev/group-signup/internal/newsletter"
	"github.com/mindovermachine-dev/group-signup/internal/server"
	"github.com/mindovermachine-dev/group-signup/internal/store"
	"github.com/mindovermachine-dev/group-signup/internal/token"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	var st store.Store = store.NewMemory()
	if cfg.FirestoreProject != "" {
		fs, err := store.NewFirestore(ctx, cfg.FirestoreProject)
		if err != nil {
			return err
		}
		defer fs.Close() //nolint:errcheck // shutting down
		st = fs
	} else {
		log.Warn("FIRESTORE_PROJECT unset: using in-memory store")
	}

	var mailer mail.Sender = mail.LogSender{Log: log}
	if cfg.SMTP != nil {
		mailer = *cfg.SMTP
	} else {
		log.Warn("SMTP_HOST unset: emails are only logged")
	}

	svc := newsletter.New(
		newsletter.Config{BaseURL: cfg.BaseURL, Lists: cfg.Lists},
		st, mailer, token.NewSigner(cfg.TokenSecret, nil),
	)
	srv := &http.Server{
		Addr:              net.JoinHostPort("", cfg.Port),
		Handler:           server.New(cfg.SiteURL, newsletter.NewHandler(svc, cfg.SiteURL, log)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("listening", "addr", srv.Addr)

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
