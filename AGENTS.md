# AGENTS.md

Guidance for coding agents working in this repository (`github.com/thezmc/edmotion`).

## Project overview

- This is a Go HTTP service that serves Vim-based code-fix challenges from
  `challenges/`.
- Entrypoint is `cmd/edmotion/main.go`.
- Runtime bootstrapping is in `internal/runtime/runtime.go`.
- Config loading is in `internal/config/settings.go`.
- HTTP router and middleware are in `internal/httpapi/`.
- Challenge loading, route handling, and solution execution are in
  `internal/challenge/`.
- Admin auth/state and admin routes are in `internal/admin/`.

## Current HTTP behavior

- Challenge routes are dynamic and registered as `/{challengeID}`:
  - `GET /<challenge-id>` returns challenge metadata plus broken source.
  - `POST /<challenge-id>` accepts a Vim normal-mode edit script and returns
    the challenge flag when the edited output matches `fixed`.
  - `OPTIONS /<challenge-id>` returns `Allow: GET, POST, OPTIONS`.
- Admin routes:
  - `PUT /admin/toggle-fixed-files`
  - `PUT /admin/set-password`
- Admin auth is exact-match on raw `Authorization` header against current admin
  password.
- Middleware stack includes request logging, panic recovery, request IDs,
  `RealIP`, per-IP rate limiting, and fixed-file visibility context injection.

## Challenge catalog behavior

- Challenges are loaded from top-level subdirectories of `CHALLENGE_DIR`
  (default `challenges`).
- Challenge ID is the directory name.
- Required files for a loadable challenge are: `broken`, `fixed`, `flag`, and
  `max`.
- `solution` is used by local validation tooling but is not required by the
  runtime loader.
- Invalid challenge directories are skipped with logs (they do not fail startup
  if at least one challenge is valid).
- Catalog auto-reload runs continuously using fsnotify (when available) plus a
  periodic poll fallback.

## Environment variables

- `HTTP_ADDR` (default `:8080`)
- `CHALLENGE_DIR` (default `challenges`)
- `ADMIN_PASSWORD` (auto-generated if unset at startup)
- `GIVE_FIXED_FILES` (`true` enables fixed-file output in challenge `GET`)
- `LOG_LEVEL` (`debug|info|warn|error`, default `info`)
- `REQUEST_LIMIT_PER_MINUTE` (default `12`)
- `CATALOG_RELOAD_INTERVAL` (Go duration string, default `60s`)
- `VIM_PATH` (optional path to vim binary; otherwise discovered via `PATH`)

## Build, test, and run

- Run tests: `go test ./...`
- Build local binary: `go build -o edmotion ./cmd/edmotion`
- Run locally: `go run ./cmd/edmotion`
- Build container image: `docker build -t edmotion .`
- Optional challenge-data tooling via mise:
  - `mise run challenges:max:validate`
  - `mise run challenges:max:sync`
  - `mise run challenges:validate`

## Editing guidelines

- Keep the service small and direct; avoid broad architecture refactors unless
  requested.
- Preserve current package boundaries (`internal/admin`, `internal/challenge`,
  `internal/httpapi`, `internal/runtime`, `internal/config`, `internal/logging`).
- Preserve logging and middleware conventions (`log/slog` + `tint`, chi-based
  router/middleware).
- Be careful when changing solution execution semantics:
  - body size limit and per-challenge max character checks
  - allowed/disallowed control characters in Vim scripts
  - Vim invocation flags/timeouts and output comparison behavior

## Security and ops notes

- Do not introduce real secrets into source or challenge assets.
- Existing challenge `flag` values are challenge artifacts; do not rotate or
  alter them unless explicitly requested.
- Admin auth intentionally uses raw header equality; do not replace with a
  different auth scheme unless explicitly requested.
