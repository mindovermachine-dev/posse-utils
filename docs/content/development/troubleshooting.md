---
title: Troubleshooting
description: Common problems with building, testing, running and deploying posse-utils, and how to fix them.
---

# Troubleshooting

## Build and tooling

| Symptom | Cause and fix |
| --- | --- |
| `make: golangci-lint: command not found` | Rebuild the devcontainer, or install golangci-lint v2 |
| `File is not properly formatted (gofmt)` | Run `make fmt` |
| `noctx` complains about `httptest.NewRequest` | Use `httptest.NewRequestWithContext(context.Background(), ...)` |
| `total coverage ... (min 80%)` | Add tests for the new code, or see what is uncovered with `make cover` |
| Docker build fails at `COPY go.sum` | Run `make tidy` first |
| Docker build fails with a Go version error | The `FROM golang:` tag in the Dockerfile must be at least the `go` directive in `go.mod` |

## Running

| Symptom | Cause and fix |
| --- | --- |
| `config: BASE_URL is required` and similar | A required variable is missing; the message lists all of them |
| `TOKEN_SECRET must be at least 32 bytes` | Use `openssl rand -hex 32` |
| `bind: address already in use` on 8080 | Another process, or the `make run` forward and a Docker container, share port 8080 |
| Subscribers disappear on restart | `FIRESTORE_PROJECT` is unset, so the in-memory store is in use |
| No email arrives locally | `SMTP_HOST` is unset; the link is in the server log |

## Browser

| Symptom | Cause and fix |
| --- | --- |
| CORS error in the console | `SITE_URL` does not exactly match the page's origin (scheme, host, port, no trailing slash) |
| Preflight fails | The request must be `POST` with `Content-Type: application/json` |
| `400` from `/v1/subscribe` | Invalid email, or the `list` is not in `LISTS` |
| Confirmation link goes to `/newsletter/invalid/` | Token expired (24 hours), used after unsubscribing, or `TOKEN_SECRET` was rotated |
| Confirmation link 404s on the site | The result pages are missing from the Astro site |

## Email

| Symptom | Cause and fix |
| --- | --- |
| `smtp send: 535` authentication error | Wrong app password, or app passwords are disabled for the account |
| `smtp send: 534` or similar | 2-Step Verification is not enabled |
| Mail goes to spam | Check SPF, DKIM and DMARC, and that `MAIL_FROM` is the authenticated mailbox |
| `mail: header injection attempt` | The recipient or subject contained a newline |

## Cloud Run

| Symptom | Cause and fix |
| --- | --- |
| `PERMISSION_DENIED` from Firestore | The service account lacks `roles/datastore.user`, or the Firestore API or database is not created |
| Revision fails to start | Read the logs (`gcloud run services logs read`); usually a missing variable or secret access |
| `Permission denied on secret` | Grant `roles/secretmanager.secretAccessor` on the specific secret to the runtime service account |
| Links in emails point to `placeholder.invalid` | Do step 7 of [Deployment](deployment.md) |
| Cold start delay on first request | Expected with `--min-instances 0`; set `1` if it matters (costs money) |
