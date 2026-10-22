// Package config loads service settings from the environment.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mindovermachine-dev/group-signup/internal/mail"
)

const minSecretLen = 32

// Config is the validated service configuration.
type Config struct {
	Port             string
	BaseURL          string // public URL of this service
	SiteURL          string // origin of the static site; also the CORS origin
	TokenSecret      []byte
	Lists            []string
	FirestoreProject string           // empty selects the in-memory store
	SMTP             *mail.SMTPSender // nil selects the log-only mailer
}

// Load reads configuration using getenv (typically os.Getenv).
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		Port:             orDefault(getenv("PORT"), "8080"),
		BaseURL:          strings.TrimRight(getenv("BASE_URL"), "/"),
		SiteURL:          strings.TrimRight(getenv("SITE_URL"), "/"),
		TokenSecret:      []byte(getenv("TOKEN_SECRET")),
		Lists:            splitList(orDefault(getenv("LISTS"), "newsletter")),
		FirestoreProject: getenv("FIRESTORE_PROJECT"),
	}

	var errs []error
	if c.BaseURL == "" {
		errs = append(errs, errors.New("BASE_URL is required"))
	}
	if c.SiteURL == "" {
		errs = append(errs, errors.New("SITE_URL is required"))
	}
	if len(c.TokenSecret) < minSecretLen {
		errs = append(errs, fmt.Errorf("TOKEN_SECRET must be at least %d bytes", minSecretLen))
	}

	if host := getenv("SMTP_HOST"); host != "" {
		c.SMTP = &mail.SMTPSender{
			Host:     host,
			Port:     orDefault(getenv("SMTP_PORT"), "587"),
			User:     getenv("SMTP_USER"),
			Password: getenv("SMTP_PASSWORD"),
			From:     getenv("MAIL_FROM"),
		}
		if c.SMTP.User == "" || c.SMTP.Password == "" || c.SMTP.From == "" {
			errs = append(errs, errors.New("SMTP_USER, SMTP_PASSWORD and MAIL_FROM are required with SMTP_HOST"))
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return c, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
