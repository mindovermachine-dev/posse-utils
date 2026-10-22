---
title: Getting started
description: How to open the project in the devcontainer, build and run it, and run it as a Docker container.
sidebar:
  order: 1
---

# Getting started

## Prerequisites

- Docker Desktop (or another Docker engine) on the host
- VS Code with the **Dev Containers** extension
- Optional: a `GH_TOKEN` environment variable on the host, so the GitHub CLI is
  authenticated inside the container

## Option 1: devcontainer (recommended)

1. Open the repository folder in VS Code.
2. Run **Dev Containers: Reopen in Container** from the Command Palette.
3. Wait for `postCreateCommand.sh` to finish. It installs `golangci-lint` and
   `goimports`, sets up the GitHub CLI extension and aliases, marks the repo as a
   safe git directory, and downloads Go modules.
4. In the container terminal:

   ```sh
   make tidy     # sync go.mod / go.sum (first time, or after changing imports)
   make check    # vet + lint + tests + coverage gate
   make run      # start the server on :8080
   ```

The container image is `mcr.microsoft.com/devcontainers/base:noble` with these
features: common-utils, GitHub CLI, Go, Docker-outside-of-Docker and the gcloud
CLI. Port 8080 is forwarded automatically.

The configuration lives in [`.devcontainer/`](../.devcontainer/devcontainer.json).

> Chat history in VS Code is stored per workspace identity. A folder opened in the
> container has a different identity from the local folder, so history does not
> carry over. See the VS Code chat export/import commands if you need to move a
> session.

## Option 2: Go on the host

Install Go (the version in `go.mod`) and `golangci-lint` v2, then use the same
`make` targets.

## Option 3: run the production image

The [Dockerfile](../Dockerfile) builds a static binary and runs it in a
distroless, non-root image.

```sh
make docker-build            # builds the image "posse-utils"

docker run --rm -p 8080:8080 \
  -e BASE_URL=http://localhost:8080 \
  -e SITE_URL=http://localhost:4321 \
  -e TOKEN_SECRET=dev-only-secret-dev-only-secret-0000 \
  posse-utils
```

Without `FIRESTORE_PROJECT` the container keeps subscribers in memory (lost on
restart). Without `SMTP_HOST` it only logs emails. Both fall back with a warning,
which is exactly what you want for local work. See [Configuration](configuration.md).

Check that it is up:

```sh
curl -i http://localhost:8080/healthz
```

If you run the image inside the devcontainer, the Docker CLI talks to the
**host's** Docker daemon (Docker-outside-of-Docker). `-p 8080:8080` therefore
binds on the host, not inside the devcontainer, and may collide with the
forwarded port 8080 of `make run`. Stop one before starting the other, or map to
another host port (`-p 8081:8080`).

