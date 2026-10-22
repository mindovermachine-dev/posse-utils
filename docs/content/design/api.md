---
title: API reference
description: HTTP endpoints exposed by posse-utils, with request and response formats and status codes.
---

# API reference

Base URL: the value of `BASE_URL`, for example `https://api.example.com`.

All responses are small. Browser-facing links redirect to pages on the static site
(`SITE_URL`); the form endpoint returns JSON errors.

## POST /v1/subscribe

Starts a subscription and sends a confirmation email.

Request (`Content-Type: application/json`, body at most 4 KiB):

```json
{ "email": "reader@example.com", "list": "newsletter", "website": "" }
```

| Field | Required | Notes |
| --- | --- | --- |
| `email` | yes | Trimmed and lower-cased. A bare address only; `Name <a@b.c>` is rejected |
| `list` | yes | Must be one of the configured `LISTS` |
| `website` | no | Honeypot. Leave empty; keep the input hidden from humans |

Responses:

| Status | Meaning |
| --- | --- |
| `202 Accepted` | Accepted. Also returned for already-confirmed addresses and honeypot hits, on purpose |
| `400 Bad Request` | Malformed JSON, oversized body, invalid email or unknown list. Body: `{"error":"..."}` |
| `500 Internal Server Error` | Storage or email failure. Body: `{"error":"internal error"}` |

## GET /v1/confirm?token=TOKEN

Linked from the confirmation email. Confirms the subscription.

| Outcome | Response |
| --- | --- |
| Success, or already confirmed | `303` to `SITE_URL/newsletter/confirmed/` |
| Invalid, expired, wrong-purpose token, or the address has since unsubscribed | `303` to `SITE_URL/newsletter/invalid/` |
| Storage failure | `500 internal error` |

## GET /v1/unsubscribe?token=TOKEN

Linked from newsletters. Unsubscribes the address; idempotent.

| Outcome | Response |
| --- | --- |
| Success, or the address is unknown | `303` to `SITE_URL/newsletter/unsubscribed/` |
| Invalid or expired token | `303` to `SITE_URL/newsletter/invalid/` |
| Storage failure | `500 internal error` |

The service can create these links with `Service.UnsubscribeURL(email, list)`;
nothing calls it yet because newsletter sending is not implemented.

## GET /healthz

Returns `200` with an empty body. Use it for uptime checks.

## CORS

Requests whose `Origin` header equals `SITE_URL` get
`Access-Control-Allow-Origin`. Preflight `OPTIONS` requests from that origin
receive `204` with `Access-Control-Allow-Methods: POST`,
`Access-Control-Allow-Headers: Content-Type` and `Access-Control-Max-Age: 600`.
Other origins get no CORS headers, so browsers block them.

## Example

```sh
curl -i -X POST https://api.example.com/v1/subscribe \
  -H 'Content-Type: application/json' \
  -H 'Origin: https://example.com' \
  -d '{"email":"reader@example.com","list":"newsletter"}'
```
