---
title: Deployment
description: Step-by-step deployment of posse-utils to Google Cloud Run with Firestore, Secret Manager and a least-privilege service account.
sidebar:
  order: 4
---

# Deployment

Target: Google Cloud Run, scale to zero, Firestore for storage, Secret Manager for
secrets.

Commands below use `gcloud`, which is installed in the devcontainer. Replace the
placeholders in the variables block.

## 1. Variables

```sh
export PROJECT_ID=my-gcp-project
export REGION=europe-west1
export SERVICE=posse-utils
export SA=posse-utils
export SITE_URL=https://example.com
export MAIL_FROM=news@example.com

gcloud auth login
gcloud config set project "$PROJECT_ID"
```

## 2. Enable APIs

```sh
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  firestore.googleapis.com \
  secretmanager.googleapis.com
```

## 3. Create the Firestore database

```sh
gcloud firestore databases create --location="$REGION" --type=firestore-native
```

Choose the location carefully: it cannot be changed afterwards.

## 4. Service account (least privilege)

```sh
gcloud iam service-accounts create "$SA" --display-name="posse-utils runtime"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:$SA@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/datastore.user"
```

The service account needs only Firestore access and, in the next step, access to
two secrets.

## 5. Secrets

```sh
openssl rand -hex 32 | tr -d '\n' | gcloud secrets create posse-token-secret --data-file=-
printf '%s' 'YOUR-GOOGLE-APP-PASSWORD' | gcloud secrets create posse-smtp-password --data-file=-

for s in posse-token-secret posse-smtp-password; do
  gcloud secrets add-iam-policy-binding "$s" \
    --member="serviceAccount:$SA@$PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor"
done
```

To rotate later, add a new version with `gcloud secrets versions add`. Rotating
`posse-token-secret` invalidates every outstanding confirmation and unsubscribe
link.

## 6. First deploy

`BASE_URL` is the service's own URL, which Cloud Run assigns on first deploy, so
set a placeholder first and correct it in step 7.

```sh
gcloud run deploy "$SERVICE" \
  --source . \
  --region "$REGION" \
  --service-account "$SA@$PROJECT_ID.iam.gserviceaccount.com" \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 3 \
  --set-env-vars "^@^BASE_URL=https://placeholder.invalid@SITE_URL=$SITE_URL@LISTS=newsletter@FIRESTORE_PROJECT=$PROJECT_ID@SMTP_HOST=smtp.gmail.com@SMTP_USER=$MAIL_FROM@MAIL_FROM=$MAIL_FROM" \
  --set-secrets "TOKEN_SECRET=posse-token-secret:latest,SMTP_PASSWORD=posse-smtp-password:latest"
```

Notes:

- `--source .` builds with Cloud Build using the repository's `Dockerfile`.
- `--allow-unauthenticated` is required because browsers and email links call the
  API directly. Protection comes from the application (CORS, validation, honeypot)
  and from `--max-instances`.
- The `^@^` prefix changes the env-var separator so values containing commas (a
  multi-list `LISTS`) are safe.
- `--max-instances 3` is a cost and abuse guard while there is no rate limiting.

## 7. Set BASE_URL

```sh
URL=$(gcloud run services describe "$SERVICE" --region "$REGION" --format='value(status.url)')
gcloud run services update "$SERVICE" --region "$REGION" --update-env-vars "BASE_URL=$URL"
echo "$URL"
```

If you map a custom domain (next section), use that domain as `BASE_URL` instead.

## 8. Smoke test

```sh
curl -i "$URL/healthz"
curl -i -X POST "$URL/v1/subscribe" \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","list":"newsletter"}'
```

Open the email, click the link, and check the document in the Firestore console
under `subscribers`.

## Custom domain (optional)

```sh
gcloud beta run domain-mappings create \
  --service "$SERVICE" --domain api.example.com --region "$REGION"
```

Add the DNS records it prints, then update `BASE_URL` to `https://api.example.com`.
Domain mapping is available only in some regions; a global external load balancer
is the alternative.

## Subsequent deploys

Re-run the deploy command from step 6 without changing env vars:

```sh
gcloud run deploy "$SERVICE" --source . --region "$REGION"
```

Existing settings, secrets and service account are kept.

## Roll back

```sh
gcloud run revisions list --service "$SERVICE" --region "$REGION"
gcloud run services update-traffic "$SERVICE" --region "$REGION" --to-revisions REVISION=100
```

## Logs

```sh
gcloud run services logs read "$SERVICE" --region "$REGION" --limit 50
```

Logs are JSON, so they are structured in Cloud Logging.

## Cost and limits

- Scale-to-zero with `--min-instances 0` costs nothing while idle.
- Cloud Run, Firestore and Secret Manager all have free tiers that comfortably
  cover newsletter-scale traffic.
- Set a budget alert on the project.

## CI deployment (recommended later)

Use GitHub Actions with Workload Identity Federation so no service account key is
stored in GitHub: create a workload identity pool and provider for the repository,
let a deployer service account (roles `run.admin`, `iam.serviceAccountUser`,
`cloudbuild.builds.editor`, plus storage and artifact registry access as needed) be
impersonated by it, and run `make check` followed by `gcloud run deploy`.

## Checklist

- [ ] Firestore database created
- [ ] Service account has only `datastore.user` and secret accessor
- [ ] `TOKEN_SECRET` and `SMTP_PASSWORD` come from Secret Manager
- [ ] `BASE_URL` matches the public URL
- [ ] `SITE_URL` is exactly the site origin (scheme and host, no trailing slash)
- [ ] SPF, DKIM and DMARC pass for the sending domain
- [ ] `--max-instances` is low
- [ ] The three result pages exist on the site ([Astro integration](astro-integration.md))
