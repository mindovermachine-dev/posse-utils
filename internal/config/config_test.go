package config_test

import (
	"strings"
	"testing"

	"github.com/mindovermachine-dev/group-signup/internal/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func validEnv() map[string]string {
	return map[string]string{
		"BASE_URL":     "https://api.example.com/",
		"SITE_URL":     "https://example.com/",
		"TOKEN_SECRET": strings.Repeat("s", 32),
	}
}

func TestLoadDefaults(t *testing.T) {
	c, err := config.Load(env(validEnv()))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != "8080" || c.BaseURL != "https://api.example.com" || c.SiteURL != "https://example.com" {
		t.Errorf("unexpected config: %+v", c)
	}
	if len(c.Lists) != 1 || c.Lists[0] != "newsletter" {
		t.Errorf("Lists = %v", c.Lists)
	}
	if c.SMTP != nil {
		t.Errorf("SMTP should be nil without SMTP_HOST")
	}
}

func TestLoadSMTPAndLists(t *testing.T) {
	e := validEnv()
	e["LISTS"] = " news , events ,"
	e["SMTP_HOST"] = "smtp.gmail.com"
	e["SMTP_USER"] = "me@example.com"
	e["SMTP_PASSWORD"] = "pw"
	e["MAIL_FROM"] = "me@example.com"

	c, err := config.Load(env(e))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(c.Lists) != 2 || c.Lists[1] != "events" {
		t.Errorf("Lists = %v", c.Lists)
	}
	if c.SMTP == nil || c.SMTP.Port != "587" {
		t.Errorf("SMTP = %+v", c.SMTP)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]string)
		want   string
	}{
		{"missing base url", func(m map[string]string) { delete(m, "BASE_URL") }, "BASE_URL"},
		{"missing site url", func(m map[string]string) { delete(m, "SITE_URL") }, "SITE_URL"},
		{"short secret", func(m map[string]string) { m["TOKEN_SECRET"] = "short" }, "TOKEN_SECRET"},
		{"smtp without credentials", func(m map[string]string) { m["SMTP_HOST"] = "h" }, "SMTP_USER"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := validEnv()
			tt.mutate(e)
			_, err := config.Load(env(e))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want mention of %s", err, tt.want)
			}
		})
	}
}
