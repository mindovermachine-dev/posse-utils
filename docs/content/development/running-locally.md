---
title: Running locally
description: How to run the server on your machine and exercise the full subscribe, confirm and unsubscribe flow by hand.
sidebar:
  order: 2
---

# Running locally

## Start the server

```sh
make run
```

`make run` supplies development defaults and starts `go run ./cmd/server`:

| Variable | Default used by `make run` |
| --- | --- |
| `BASE_URL` | `http://localhost:8080` |
| `SITE_URL` | `http://localhost:4321` (Astro's default dev port) |
| `TOKEN_SECRET` | a fixed dev-only string |

Override any of them from your shell, for example
`SITE_URL=http://localhost:3000 make run`.

Persistence is in memory and email is logged, because `FIRESTORE_PROJECT` and
`SMTP_HOST` are unset. Restarting the server forgets all subscribers.

## Walk through the flow

1. Subscribe:

   ```sh
   curl -i -X POST http://localhost:8080/v1/subscribe \
     -H 'Content-Type: application/json' \
     -d '{"email":"me@example.com","list":"newsletter"}'
   ```

   Expect `202 Accepted`.

2. Find the confirmation link in the server log. The JSON log line has
   `"msg":"mail (not sent)"` and the link in the `body` field.

3. Open the link (copy the URL from the log, it contains the token):

   ```sh
   curl -i 'http://localhost:8080/v1/confirm?token=PASTE_TOKEN'
   ```

   Expect `303 See Other` with `Location: http://localhost:4321/newsletter/confirmed/`.

4. Confirm that an invalid token is handled:

   ```sh
   curl -i 'http://localhost:8080/v1/confirm?token=nonsense'
   ```

   Expect a `303` to `/newsletter/invalid/`.

## Error cases worth trying

```sh
# Invalid email -> 400
curl -i -X POST localhost:8080/v1/subscribe -H 'Content-Type: application/json' \
  -d '{"email":"nope","list":"newsletter"}'

# Unknown list -> 400
curl -i -X POST localhost:8080/v1/subscribe -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","list":"other"}'

# Honeypot filled -> 202, but nothing is stored or sent
curl -i -X POST localhost:8080/v1/subscribe -H 'Content-Type: application/json' \
  -d '{"email":"me@example.com","list":"newsletter","website":"x"}'
```

## CORS check

The API only answers cross-origin preflights from `SITE_URL`:

```sh
curl -i -X OPTIONS localhost:8080/v1/subscribe -H 'Origin: http://localhost:4321'
curl -i -X OPTIONS localhost:8080/v1/subscribe -H 'Origin: https://evil.example'
```

The first returns `204` with `Access-Control-Allow-Origin`; the second does not.

## Real email locally

To send real mail from your machine, set the SMTP variables (see
[Email](email.md)). Use your own address as the recipient.

```sh
SMTP_HOST=smtp.gmail.com SMTP_USER=you@yourdomain.com \
SMTP_PASSWORD='app-password' MAIL_FROM=you@yourdomain.com \
BASE_URL=http://localhost:8080 make run
```

Never commit credentials. `.env` is ignored by git, but `make run` does not read
it; export the variables in your shell or use a tool such as `direnv`.

## Local Firestore

Use the Firestore emulator when you want persistence without a real project:

```sh
gcloud emulators firestore start --host-port=localhost:8081
# in another terminal
FIRESTORE_EMULATOR_HOST=localhost:8081 FIRESTORE_PROJECT=demo-posse make run
```

The Firestore client honours `FIRESTORE_EMULATOR_HOST` automatically. The emulator
needs a Java runtime and the `cloud-firestore-emulator` gcloud component.
