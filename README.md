# gemini-api

A lightweight HTTP gateway that exposes Google Gemini's web conversation endpoint as a standard **OpenAI-compatible REST API**. Written in Go, shipped as a single static binary — nothing to configure, no upstream account, no tokens.

The build pipeline compiles the server core from the upstream Go module and wraps it with a portable runtime layer (dynamic config generation, container health checks, environment-driven auth).

## What you get

| | |
|---|---|
| Endpoints | `POST /v1/chat/completions`, `GET /v1/models`, `POST /v1/responses` |
| Auth | Optional Bearer keys via env (open by default) |
| Streaming | Server-Sent Events |
| Function calling | OpenAI tool-call schema |
| Vision | Base64 and `image_url` inputs |
| Models | Flash 3.6 / 3.7, Extended Thinking, Pro, Auto, Lite |
| Reasoning control | `@think=N` suffix, 0 (deep) → 4 (fast) |
| Codex CLI | Native Responses API |

## Running on Render

Deploy as a **Web Service** with the Docker runtime (free tier works):

1. Push this repository to GitHub.
2. Render → **New +** → **Web Service**.
3. Pick the repo. Runtime: **Docker**, plan: **Free**.
4. Add env vars (see below) if needed.

The container listens on `$PORT` automatically and answers a health check at `/v1/models`.

### Environment

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port (Render injects it) |
| `HOST` | `0.0.0.0` | Bind address |
| `API_KEYS` | empty | Comma-separated bearer tokens; empty means the endpoint is public |
| `DEFAULT_MODEL` | `gemini-3.6-flash` | Model used when the request omits one |
| `IMPERSONATE` | empty | Chrome/Edge TLS fingerprint (`chrome_120`…) for WAF-protected networks |

## Running locally

With Docker:

```bash
docker compose up --build
```

Without Docker (needs Go 1.23+):

```bash
./build.sh
./gemini-api --port 8080
```

## Smoke test

```bash
curl http://localhost:8080/v1/models

curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"ping"}]}'
```

Point any OpenAI SDK or agent (Cline, Codex CLI, OpenWebUI) at `http://localhost:8080/v1` — or your Render URL — and go.

## Notes on hosting

Google applies risk-based filtering on its web endpoint. Requests originating from residential networks pass through routinely (verified on Android/Termux). Traffic from cloud/datacenter ranges may occasionally trigger 403/429 responses; setting `IMPERSONATE` to a browser fingerprint resolves most cases.

## License

MIT. See [LICENSE](LICENSE).
