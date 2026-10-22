---
title: Email with Google Workspace
description: How to send confirmation email through Google Workspace SMTP using an app password, plus limits and alternatives.
---

# Email with Google Workspace

posse-utils sends mail through an authenticated SMTP submission server. No mail
server is needed. The default target is Google Workspace.

## Set up SMTP with an app password

1. Choose or create a mailbox on your domain to send from, for example
   `news@example.com`.
2. Turn on 2-Step Verification for that account. A Workspace admin may need to
   allow it.
3. Create an **app password** in the Google Account security settings. Admins can
   disable app passwords for the organisation; if the option is missing, ask the
   admin.
4. Store the password in Secret Manager (see [Deployment](deployment.md)).
5. Set:

   ```text
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=news@example.com
   SMTP_PASSWORD=<app password>
   MAIL_FROM=news@example.com
   ```

Port 587 with STARTTLS is what Go's `net/smtp` negotiates automatically. Cloud Run
allows outbound 587 and 465; only port 25 is blocked.

## Deliverability

- Send from an address on your own domain.
- Keep SPF, DKIM and DMARC configured for the domain. They normally are for a
  Workspace domain. Verify with the Google Admin Toolbox or by sending to a test
  mailbox and viewing the headers.
- Keep the From address equal to the authenticated mailbox, otherwise Gmail rewrites
  or rejects it.

## Limits

Workspace limits sending to roughly 2,000 messages per day per user (lower for
trial accounts). This is ample for confirmation emails and a small newsletter.
Sending to thousands of subscribers needs a dedicated provider.

## Switching provider later

Sending sits behind the `mail.Sender` interface:

```go
type Sender interface {
    Send(ctx context.Context, m Message) error
}
```

To use the Gmail API (service account with domain-wide delegation, no password to
rotate), a transactional provider, or another SMTP host, add an implementation and
select it in `cmd/server/main.go`. Nothing in `newsletter` changes.

## Testing without sending

- Unset `SMTP_HOST`: `LogSender` writes the message, including the link, to the
  log.
- In unit tests use `mail.Outbox`, which records messages in memory.
