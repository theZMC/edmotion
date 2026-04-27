# edmotion

`edmotion` is a small Go HTTP service for Vim-based code-fix challenges.

The server and challenges were originally written for the 2026
[JOLT Cyber Challenge](https://www.venturecenter.co/programs/jolt/).

Each challenge lives in `challenges/<challenge-id>/` and includes:

- `broken` - code presented to players
- `fixed` - reference output after a correct edit
- `solution` - maintainer helper file (not required by runtime)
- `max` - max number of characters accepted in a submitted Vim script
- `flag` - response returned for a correct solution

## How it works

- `GET /<challenge-id>` returns challenge metadata and the broken source.
- `POST /<challenge-id>` accepts a Vim normal-mode key script.
- `OPTIONS /<challenge-id>` returns `Allow: GET, POST, OPTIONS`.
- `PUT /admin/toggle-fixed-files` toggles whether GET includes the fixed source.
- `PUT /admin/set-password` updates the admin password.

Admin endpoints require an exact raw `Authorization` header value.

## Quick start (local)

1. Start the server:

```bash
ADMIN_PASSWORD=devpass HTTP_ADDR=:8080 go run ./cmd/edmotion
```

2. In another terminal:

```bash
BASE_URL="http://127.0.0.1:8080"
ADMIN_PASSWORD="devpass"
```

3. List available challenge IDs:

```bash
ls challenges
```

## API examples

Fetch a challenge:

```bash
curl -i "$BASE_URL/auth-bypass"
```

Submit a solution script:

```bash
curl -i -X POST \
  --data '20jeXp' \
  "$BASE_URL/signal-intercept"
```

Check allowed methods:

```bash
curl -i -X OPTIONS "$BASE_URL/auth-bypass"
```

Toggle fixed-file visibility:

```bash
curl -i -X PUT \
  -H "Authorization: $ADMIN_PASSWORD" \
  "$BASE_URL/admin/toggle-fixed-files"
```

Set a new admin password:

```bash
curl -i -X PUT \
  -H "Authorization: $ADMIN_PASSWORD" \
  --data 'password=newpass123' \
  "$BASE_URL/admin/set-password"
```

## Configuration

- `HTTP_ADDR` (default `:8080`)
- `CHALLENGE_DIR` (default `challenges`)
- `ADMIN_PASSWORD` (auto-generated if unset)
- `GIVE_FIXED_FILES` (`true` enables fixed output in `GET`)
- `LOG_LEVEL` (`debug|info|warn|error`, default `info`)
- `REQUEST_LIMIT_PER_MINUTE` (default `12`)
- `CATALOG_RELOAD_INTERVAL` (Go duration, default `60s`)
- `VIM_PATH` (optional path to `vim` binary)

## Build and run

Run tests:

```bash
go test ./...
```

Build binary:

```bash
go build -o edmotion ./cmd/edmotion
```

Run binary:

```bash
./edmotion
```

Build container:

```bash
docker build -t edmotion .
```

Run with compose:

```bash
docker compose up --build
```

## Maintainer utilities (optional)

If you use `mise`, the repository includes challenge maintenance tasks:

- `mise run challenges:max:validate`
- `mise run challenges:max:sync`
- `mise run challenges:validate`
