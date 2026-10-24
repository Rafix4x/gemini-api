# gemini-api — OpenAI-compatible Gemini gateway (Go)

<div align="center">

![GitHub release](https://img.shields.io/github/v/release/Rafix4x/gemini-api?style=for-the-badge&color=4285F4)
![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![Render](https://img.shields.io/badge/Render-deployable-46E3B7?style=for-the-badge&logo=render&logoColor=white)
![License](https://img.shields.io/github/license/Rafix4x/gemini-api?style=for-the-badge&color=yellow)

**One small static binary — two full API shapes.**

OpenAI-compatible **and** native Gemini (`v1beta`) in a single Go binary. Anonymous out of the box, optional cookie for real Pro routing, optional bearer keys.

</div>

---

## Table of contents

1. [Features](#-features)
2. [Quick start](#-quick-start)
3. [API reference](#-api-reference)
4. [Examples](#-examples)
5. [SDK & tooling compatibility](#-sdk--tooling-compatibility)
6. [Deploy](#-deploy)
7. [Configuration](#-configuration)
8. [Images](#-images)
9. [Pro models](#-pro-models)
10. [Reasoning levels](#-reasoning-levels)
11. [Observability](#-observability--logs)
12. [Security & hardening](#-security--hardening)
13. [How it works](#-how-it-works)
14. [Troubleshooting](#-troubleshooting)
15. [Development](#-development)
16. [License](#-license)

---

## ✨ Features

| Area | What it delivers |
|---|---|
| **Dual protocol** | OpenAI shape `/v1/…` **and** native Gemini `/v1beta/…` from one server |
| **Streaming** | Server-Sent Events (`"stream": true`) + `:streamGenerateContent` (ndjson) |
| **Function calling** | OpenAI `tools` / `tool_calls` *and* Gemini `functionDeclarations` / `functionCall` |
| **Vision** | `data:` URLs, `http(s)` image URLs, base64 PNG, inline Gemini `inlineData` |
| **Reasoning control** | `@think=N` override (`0` deepest → `4` fastest) |
| **Responses API** | `/v1/responses` — native Codex CLI support |
| **Images** | `/v1/images/generations` (OpenAI images shape) + edits via reference image |
| **Self-healing** | Auto-updates Gemini upstream build id on startup = survives upstream changes |
| **Zero config** | Anonymous mode by default; optional bearer keys; static single binary |
| **Env-first** | Every flag overridable via env vars |

---

## ⚡ Quick start

### Requirements

- Go ≥ 1.26.5 (bundled in official Docker image)
- Docker 24+ (optional)

### Build from source

```bash
./build.sh              # emits ./gemini-api
./gemini-api --port 8080

# or, with a config file:
cp config.json.example config.json   # if present
./gemini-api --config config.json
```

### Docker

```bash
docker compose up --build
# or, handheld:
docker build -t gemini-api .
docker run --rm -p 8080:8080 gemini-api
```

### Verify

```bash
curl http://localhost:8080/
curl http://localhost:8080/v1/models | jq .
```

### First call

```bash
curl -s http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"Capital of Bangladesh?"}]}' \
  | jq '.choices[0].message.content'
```

---

## 📡 API reference

### OpenAI-compatible surface

| Method | Endpoint | Description |
|:---:|---|---|
| `GET`  | `/`                        | health / build / status |
| `GET`  | `/v1/models`               | list OpenAI-style model objects |
| `POST` | `/v1/chat/completions`     | chat + streaming |
| `POST` | `/v1/responses`            | OpenAI Responses API (Codex-native) |
| `POST` | `/v1/images/generations`   | image generation + edit |
| `GET`  | `/v1/images/{id}.png`      | serve a generated image |

### Native Gemini surface (`v1beta`)

| Method | Endpoint | Description |
|:---:|---|---|
| `GET`  | `/v1beta/models`                                | list models in Gemini shape |
| `POST` | `/v1beta/models/{model}:generateContent`         | generation (sync) |
| `POST` | `/v1beta/models/{model}:streamGenerateContent`  | incremental stream (ndjson) |

### Model registry

| Model | Notes |
|---|---|
| `gemini-4-argon`                | flagship reasoning, unified 1M context |
| `gemini-3.8-flash`              | most intelligent Flash (coding / agents) |
| `gemini-3.6-flash`              | all-rounder *(default)* |
| `gemini-3.7-flash`              | latest all-rounder, very strong reasoning |
| `gemini-3.5-flash`              | alias → `gemini-3.6-flash` |
| `gemini-3.5-flash-thinking`     | deep thinking, longest outputs |
| `gemini-3.5-flash-thinking-lite`| adaptive-depth reasoning |
| `gemini-3.1-pro`                | Pro *(cookie needed for real routing)* |
| `gemini-3.1-pro-enhanced`       | Pro enhanced *(experimental)* |
| `gemini-auto`                   | automatic model selection |
| `gemini-flash-lite`             | lightweight & fastest |
| `gemini-image`                  | image generation *(signed cookie needed)* |

> 💡 **Reasoning override** – append `@think=N` (e.g. `gemini-3.6-flash@think=0` deepest … `@think=4` fastest). See [Reasoning levels](#-reasoning-levels).

---

## 🧪 Examples

### Sync chat

```bash
curl -s http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model":"gemini-3.6-flash",
    "messages":[
      {"role":"system","content":"Terse docs assistant."},
      {"role":"user","content":"Explain SSE in 2 sentences."}
    ],
    "temperature":0.7,
    "max_tokens":256
  }'
```

### SSE streaming (OpenAI wire)

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","stream":true,"messages":[{"role":"user","content":"Count 1-5"}]}'
```

### Vision via data URL

```bash
curl -s http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model":"gemini-3.6-flash",
    "messages":[
      {"role":"user","content":[
        {"type":"text","text":"Describe this image"},
        {"type":"image_url","image_url":{"url":
           "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA…"}}
      ]}
    ]
  }'
```

### OpenAI function calling

```bash
curl -s http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model":"gemini-3.6-flash",
    "messages":[{"role":"user","content":"Weather in Dhaka?"}],
    "tools":[{"type":"function","function":{
      "name":"get_weather","description":"Current weather of a city",
      "parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}
    }}],
    "tool_choice":"auto"
  }'
```

### Native Gemini shape

```bash
curl -s http://localhost:8080/v1beta/models/gemini-3.6-flash:generateContent \
  -H 'Content-Type: application/json' \
  -d '{"contents":[{"parts":[{"text":"ping"}]}]}'
```

---

## 🧰 SDK & tooling compatibility

Point **any OpenAI/SSE-compatible client** to `http://HOST:PORT/v1`. Works great with:

- OpenAI SDK (Python, Node, Go)
- LangChain, LiteLLM
- `curl`, `httpx`, fetch-style clients
- Codex CLI (via `/v1/responses`)
- Agent Zero and any OpenAI-compat consumers
- Any Gemini SDK pointed at `/v1beta`

---

## 🚀 Deploy

### Render (free Web Service, Docker runtime)

1. Push the repo to GitHub.
2. Render → **New +** → Web Service → pick the repo.
3. Runtime **Docker**, plan **Free**.
4. Add env vars if desired → **Create**.
5. Service URLs from `$PORT`, health checks at `GET /` & `GET /v1/models`.

### Docker compose (VPS)

```bash
docker compose up --build -d
```

### Fly.io

```bash
fly launch --no-deploy
fly deploy
```

### Any VPS (systemd)

```bash
sudo install -m 755 gemini-api /opt/gemini-api/gemini-api
# /etc/systemd/system/gemini-api.service
ExecStart=/opt/gemini-api/gemini-api --config /opt/gemini-api/config.json
```

> ⚠️ Vercel **won't work** — long-lived process required. Use Render, Docker, or a VPS.

---

## ⚙️ Configuration

The most important env vars (all pluggable in `start.sh`):

| Variable | Default | Purpose |
|---|---|---|
| `PORT`                  | `8080` (Docker) | listen port (Render injects) |
| `HOST`                  | `0.0.0.0`       | bind address |
| `API_KEYS`              | *(empty)*       | comma-separated bearer tokens, empty = public |
| `DEFAULT_MODEL`         | `gemini-3.6-flash` | used when `model` missing |
| `IMPERSONATE`           | *(empty)*       | TLS fingerprint like `chrome_120` |
| `COOKIE_FILE`           | *(empty)*       | path to `gemini.google.com` cookies |
| `PROXY` / `HTTP_PROXY`  | *(empty)*       | HTTP(S) proxy (config + std env) |
| `IMAGE_MODEL`           | `gemini-image`  | exposed name for `/v1/images/generations` |
| `IMAGE_MODEL_ID`        | `7`             | upstream model id for image model |
| `IMAGE_BACKEND`         | `gemini`        | `gemini` web or `local` image API |
| `LOCAL_IMAGE_URL`       | *(empty)*       | base url of `local` image backend |
| `LOCAL_IMAGE_PROVIDER`  | `alpha`         | provider key for local backend |
| `LOCAL_IMAGE_MODEL`     | `flare`         | model for local backend |
| `LOCAL_IMAGE_KEY`       | *(empty)*       | bearer token for local backend |

A `config.json` (JSON with keys matching flags) is the recommended production path. `start.sh` always regenerates a file from env for Docker run.

---

## 🖼️ Images

`POST /v1/images/generations` follows OpenAI `images` shape and supports edits:

```bash
curl -X POST https://yourapp.onrender.com/v1/images/generations \
  -H 'content-type: application/json' \
  -d '{"model":"gemini-image","prompt":"a red bicycle on a white wall","n":1}'
```

```json
{
  "created": 1700000000,
  "data": [{ "url": "https://yourapp.onrender.com/v1/images/img_9f2c….png" }]
}
```

- `response_format: "b64_json"` → inline bytes
- `image` (http(s) URL, data URL, base64) → edit an existing image
- Gemini web upstream requires a **signed session**; otherwise 502
- Or use `IMAGE_BACKEND=local` with any backend that answers
  `{ "success":true, "imageUrl":"…" }` plus download

---

## 🍪 Pro models

Anonymous calls cover the **Flash tiers** (`0`-cool chain). Real `gemini-3.1-pro` routing requires a signed session:

1. Export cookies from `gemini.google.com` while signed in (any cookie extension → JSON)
2. Pass `--cookie-file /path/to/cookies.json` or `COOKIE_FILE`
3. If your account demands, set `auth_user` / `xsrf_token` in `config.json`

Without a cookie, Pro model names *still respond* but downgrade to a Flash tier.

---

## 🧠 Reasoning levels

All models accept the **`@think=N`** suffix:

```
gemini-3.6-flash@think=0   deepest reasoning (slow, largest output)
gemini-3.6-flash@think=1   …
gemini-3.6-flash@think=3   …
gemini-3.6-flash@think=4   fastest (minimal reasoning)
```

- `think=0` and `think=1` unlock the `* -thinking` variants benefit (deep chain)
- On-the-fly hints override without re-deploying

---

## 📊 Observability / logs

- Structured startup banner (`gemini-web2api` version, host, models, cookie, proxy, impersonate…)
- Per-request logging toggled by `log_requests` in `config.json`
- Health check: `GET /` returns status, `GET /v1/models` probes end to end
- Docker `HEALTHCHECK` polls `GET /v1/models`
- Auto-refresh log: `BL auto-updated: …` when build id changed

---

## 🔒 Security & hardening

- `go build -ldflags="-s -w"` — static, stripped binary
- `chmod 600` cookie/config files (sticky sticky umask)
- Bearer tokens (`API_KEYS`) → enforce with conventional OpenAI `Authorization`
- Zero `CGO`, multi-stage Docker image (Alpine runtime)
- Never leaks internal error details unless debug explicitly enabled
- Upstream responses never proxy cookies back

---

## 🏗️ How it works

```
┌────────────┐  OpenAI body  ┌──────────────┐   web protocol  ┌─────────┐
│  your app  │ ─────JSON───▶ │  gemini-api  │ ──────────────▶ │ Gemini  │
│OpenAI/Gem  │ ◀────JSON───── │   gateway    │ ◀────────────── │  web    │
│  client    │                │ (Go, 1 bin)  │  stream tokens  └─────────┘
└────────────┘                └──────┬───────┘
                                      │  auto-refreshes upstream build id
                                      ▼  on startup
                                keeps working across Google releases
```

- **Dual protocol** – one server speaks both OpenAI and native Gemini wire
- **Self-healing** – build id refreshed automatically; no redeploy on upstream shuffles
- **Lightweight** – single static binary, browser session provisioned lazily

---

## ❓ Troubleshooting

| Problem | Likely cause | Fix |
|---|---|---|
| `502` on image gen | no signed session | Set `COOKIE_FILE` or `IMAGE_BACKEND=local` |
| Pro model is "Flash-like" | cookie not synced | Re-export cookies (expired); enable `auth_user` |
| Cookie XSRF 400 | session rotation | Rotate cookie each several hours |
| `429` from upstream | anon rate limiting | reuse `IMPERSONATE=chrome_120` or proxy |
| Build id mismatch | Google released new | wait 30s for auto-refresh log |
| SDK connection refused | wrong host/port, cert | verify `/v1/models`, env `OPENAI_BASE_URL` |
| SSE truncation | proxy buffering | disable proxy buffer / use `X-Accel-Buffering: no` |
| Permission denied | umask / chmod | ensure config/cookie mode 600 |

---

## 🧑‍💻 Development

```bash
go mod download
go run . --port 8080        # dev (with full output)
CGO_ENABLED=0 go build -ldflags="-s -w" -o gemini-api .    # release build
go test ./...              # if tests exist
```

- Modules in `go.mod` (bogdanfinn/tls-client for fingerprint)
- HTTP core in `internal/server`, models in `internal/models`, config loader in `internal/config`
- Contribute: PRs welcome; keep docs + API surface in sync

### Repo layout

```
gemini-api/
├── main.go              # entry point & flags
├── go.mod / go.sum
├── Dockerfile / docker-compose.yml
├── build.sh / start.sh
├── internal/
│   ├── config/    # load + defaults + env
│   ├── models/    # registry & constants
│   └── server/    # dual-protocol http router & handlers
├── README.md
└── LICENSE
```

---

## 📝 License

MIT (per badging, check `LICENSE`). Not affiliated with Google.

<br/>

<div align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:D96570,50:9B72CB,100:4285F4&height=120&section=footer" width="100%" />
  <sub>Built with Go · MIT Licensed · Not affiliated with Google</sub>
</div>