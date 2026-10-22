---
title: Astro integration
description: How to add the newsletter signup form and the confirmation, unsubscribe and invalid-link pages to the Astro site.
sidebar:
  order: 5
---

The site needs one form and three static pages.

## Environment

Expose the API URL to the site, for example in `.env`:

```text
PUBLIC_POSSE_API=https://api.example.com
```

`SITE_URL` on the service must equal the site's origin, otherwise the browser
blocks the request (CORS).

## Signup form component

`src/components/Signup.astro`:

```astro
---
const api = import.meta.env.PUBLIC_POSSE_API;
---
<form id="signup" data-api={api}>
  <label>
    Email
    <input type="email" name="email" required autocomplete="email" />
  </label>
  <!-- Honeypot: hidden from people, bots fill it in -->
  <input type="text" name="website" tabindex="-1" autocomplete="off"
         style="position:absolute;left:-9999px" aria-hidden="true" />
  <button type="submit">Subscribe</button>
  <p id="signup-status" role="status"></p>
</form>

<script>
  const form = document.querySelector<HTMLFormElement>("#signup")!;
  const status = document.querySelector<HTMLElement>("#signup-status")!;

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    const data = new FormData(form);
    status.textContent = "Sending...";
    try {
      const res = await fetch(`${form.dataset.api}/v1/subscribe`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: data.get("email"),
          list: "newsletter",
          website: data.get("website"),
        }),
      });
      status.textContent = res.ok
        ? "Check your inbox to confirm your subscription."
        : "That did not look like a valid email address.";
    } catch {
      status.textContent = "Something went wrong. Please try again later.";
    }
  });
</script>
```

Always show the same success message when the response is `202`, even if the
address was already subscribed. The API intentionally does not reveal that.

## Result pages

The API redirects the browser to these paths on `SITE_URL`. Create them as plain
static pages:

| Path | Shown when |
| --- | --- |
| `/newsletter/confirmed/` | The confirmation link worked |
| `/newsletter/unsubscribed/` | The unsubscribe link worked |
| `/newsletter/invalid/` | The link is invalid, expired or used after unsubscribing; suggest subscribing again |

Example `src/pages/newsletter/confirmed.astro`:

```astro
---
import Layout from "../../layouts/Layout.astro";
---
<Layout title="Subscription confirmed">
  <h1>You are subscribed</h1>
  <p>Thanks for confirming. You will hear from me soon.</p>
</Layout>
```

With Astro's default `trailingSlash` behaviour, `confirmed.astro` is served at
`/newsletter/confirmed/`. If you configure `trailingSlash: "never"`, adjust the
paths in `internal/newsletter/handler.go` to match.

## Privacy

Link to your privacy policy from the form and state what the address is used for.
Subscribers can unsubscribe at any time through the link in every email.
