---
title: Configuration
description: Reference for every environment variable read by posse-utils, with defaults, validation rules and examples.
---

# Configuration

All configuration comes from environment variables, read once at startup by
[`internal/config`](../internal/config/config.go). Invalid configuration stops the
process with a message listing every problem.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `BASE_URL` | yes | | Public URL of this service. Used to build links in emails. Trailing slash is trimmed |
| `SITE_URL` | yes | | Origin of the static site, for example `https://example.com`. Used as the CORS origin and as the redirect target. No path |
| `TOKEN_SECRET` | yes | | HMAC key, at least 32 bytes. Generate with `openssl rand -hex 32` |
| `PORT` | no | `8080` | Listen port. Cloud Run sets this for you |
| `LISTS` | no | `newsletter` | Comma-separated allow-list of list names |
| `FIRESTORE_PROJECT` | no | | GCP project ID. Unset means an in-memory store (data lost on restart) |
| `SMTP_HOST` | no | | SMTP server. Unset means emails are only logged |
| `SMTP_PORT` | no | `587` | SMTP port (STARTTLS) |
| `SMTP_USER` | with `SMTP_HOST` | | SMTP username, usually the mailbox address |
| `SMTP_PASSWORD` | with `SMTP_HOST` | | SMTP password or app password |
| `MAIL_FROM` | with `SMTP_HOST` | | From address, for example `news@example.com` |

## Behaviour by combination

| `FIRESTORE_PROJECT` | `SMTP_HOST` | Result |
| --- | --- | --- |
| unset | unset | Local development: in-memory data, emails in the log |
| set | unset | Real storage, emails still only logged (useful for staging) |
| unset | set | Real emails, data lost on restart (do not use in production) |
| set | set | Production |

Firestore credentials come from Application Default Credentials: the Cloud Run
service account in production, or `gcloud auth application-default login` on a
workstation. `FIRESTORE_EMULATOR_HOST` redirects the client to the emulator.

## Secrets

Treat `TOKEN_SECRET` and `SMTP_PASSWORD` as secrets: keep them out of the image,
the repository and shell history. In production use Secret Manager (see
[Deployment](deployment.md)).

## Example

```sh
BASE_URL=https://api.example.com
SITE_URL=https://example.com
LISTS=newsletter,events
FIRESTORE_PROJECT=my-gcp-project
SMTP_HOST=smtp.gmail.com
SMTP_USER=news@example.com
MAIL_FROM=news@example.com
# TOKEN_SECRET and SMTP_PASSWORD come from Secret Manager
```
