<div align="center">

# ⚡ gemini-api

**Gemini Web → OpenAI- & Gemini-compatible API gateway**

A single-binary Go server that turns Google Gemini's web endpoint into a drop-in
**OpenAI-compatible** *and* **native Gemini (`v1beta`)** REST API. Works anonymously
out of the box — no API key, no account, nothing to configure. Add a cookie only if
you want real Pro-model routing.

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com)
[![Render](https://img.shields.io/badge/Render-deploy%20ready-46E3B7?logo=render&logoColor=white)](https://render.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

</div>

---

## ✨ Features

| | |
|---|---|
| 🔌 **Two APIs, one server** | OpenAI shape (`/v1/...`) **and** native Gemini shape (`/v1beta/...`) side by side |
| 🌊 **Streaming** | Server-Sent Events with `"stream": true` (and `:streamGenerateContent`) |
| 🛠️ **Function calling** | OpenAI `tools` / Gemini `functionDeclarations` |
| 👁️ **Vision** | Base64 data URLs, `image_url`, and inline Gemini `inlineData` |
| 🧠 **Reasoning control** | `@think=N` suffix — `0` (deepest) → `4` (fastest) |
| 📡 **Responses API** | `/v1/responses` — native support for Codex CLI |
| 🔁 **Auto BL refresh** | Self-updates Gemini's `bl` build id on startup, so it keeps working as Google ships |
| 🔐 **Optional auth** | Bearer keys via `API_KEYS`; open by default |
| 🍪 **Pro routing** | Drop in a cookie to unlock real Pro-model responses |
| 📦 **Zero config** | One static binary, no upstream account or tokens required |

## 📡 API

### Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/` | Health / status — returns `version` and the model list |
| `GET` | `/v1/models` | List models (OpenAI shape) |
| `POST` | `/v1/chat/completions` | Chat completion (OpenAI-compatible) |
| `POST` | `/v1/responses` | Responses API (Codex CLI native) |
| `GET` | `/v1beta/models` | List models (native Gemini shape) |
| `POST` | `/v1beta/models/{model}:generateContent` | Native Gemini generate |
| `POST` | `/v1beta/models/{model}:streamGenerateContent` | Native Gemini streaming |

### Models

Names below are stable labels the gateway exposes; the server maps each to a Gemini
web routing mode, so the backend behind a label may be upgraded transparently by Google.

| Model | Notes |
|---|---|
| `gemini-3.7-flash` | Latest all-around model |
| `gemini-3.6-flash` | All-around model (**default**) |
| `gemini-3.5-flash` | Alias for `gemini-3.6-flash` |
| `gemini-3.5-flash-thinking` | Deep thinking mode, longest output (~20k chars) |
| `gemini-3.5-flash-thinking-lite` | Dynamic thinking with adaptive depth |
| `gemini-3.1-pro` | Pro model (needs a cookie for real Pro routing) |
| `gemini-3.1-pro-enhanced` | Pro with enhanced output (experimental) |
| `gemini-auto` | Automatic model selection |
| `gemini-flash-lite` | Lightweight, fastest |

**Reasoning depth** — append `@think=N` to any model to override its default effort:
`0` = deepest reasoning, `4` = fastest. Example: `"model": "gemini-3.6-flash@think=0"`.

### Quick example

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "gemini-3.6-flash",
    "messages": [{"role": "user", "content": "ping"}]
  }'
```

```json
{
  "id": "chatcmpl-c3e69f8e85ed",
  "object": "chat.completion",
  "model": "gemini-3.6-flash",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "pong"},
    "finish_reason": "stop"
  }],
  "usage": {"prompt_tokens": 5, "completion_tokens": 2, "total_tokens": 7}
}
```

Streaming works the same way — set `"stream": true` and consume the SSE chunks, which
end with `data: [DONE]`.

### Native Gemini shape

If your client already speaks the Google GenAI API, point it at `/v1beta` instead:

```bash
curl http://localhost:8080/v1beta/models/gemini-3.6-flash:generateContent \
  -H 'Content-Type: application/json' \
  -d '{"contents":[{"parts":[{"text":"ping"}]}]}'
```

## 🚀 Getting started

### Run locally (Go 1.26+)

```bash
./build.sh
./gemini-api --port 8080
```

### Run locally (Docker)

```bash
docker compose up --build
```

### Deploy on Render

Free tier works — deploy as a **Web Service** with the **Docker** runtime:

1. Push this repository to GitHub.
2. Render → **New +** → **Web Service**.
3. Pick the repo. Runtime: **Docker**, plan: **Free**.
4. Add env vars (below) if needed.

The container reads `$PORT` (Render injects it) and answers health checks at both
`GET /` and `GET /v1/models`.

### Command-line flags

```
--port N            Listen port (overrides config)
--config PATH       Path to a config.json
--cookie-file PATH  Cookie file for Pro routing
--proxy URL         HTTP proxy
--impersonate NAME  TLS fingerprint profile (e.g. chrome_120)
--version           Print version and exit
```

### Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` (Docker) / `8081` (raw binary) | Listen port; Render injects it |
| `HOST` | `0.0.0.0` | Bind address |
| `API_KEYS` | *(empty)* | Comma-separated bearer tokens; empty = public |
| `DEFAULT_MODEL` | `gemini-3.6-flash` | Model used when a request omits one |
| `IMPERSONATE` | *(empty)* | Chrome/Edge TLS fingerprint (`chrome_120`…) for WAF-protected networks |

### Config file (optional)

Anything the flags/env cover can also live in a `config.json` (auto-discovered at
`./config.json` or `~/.config/gemini-web2api/config.json`):

```json
{
  "port": 8080,
  "host": "0.0.0.0",
  "default_model": "gemini-3.6-flash",
  "request_timeout_sec": 180,
  "retry_attempts": 3,
  "api_keys": [],
  "cookie_file": "",
  "auth_user": "",
  "xsrf_token": "",
  "proxy": "",
  "impersonate": "",
  "temporary_chats": false
}
```

Set `"temporary_chats": true` to ask Gemini not to persist the conversation server-side.

## 🍪 Unlocking Pro models

Anonymous requests are routed through Gemini's public web endpoint, which is enough for
the Flash tiers. For real `gemini-3.1-pro` routing, supply a signed-in session:

- point `cookie_file` (or `--cookie-file`) at a file with your `gemini.google.com`
  cookies, and
- set `auth_user` / `xsrf_token` if your account needs them.

Without a cookie, Pro model names still respond, but Google may downgrade them to a
Flash tier.

## 🔗 Connecting clients

Point any OpenAI SDK or agent at `http://localhost:8080/v1` (or your Render URL):

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://your-render-url.onrender.com/v1",
    api_key="not-needed",   # or one of your API_KEYS if auth is enabled
)

resp = client.chat.completions.create(
    model="gemini-3.6-flash",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(resp.choices[0].message.content)
```

Works as a drop-in base URL for Cline, Codex CLI, OpenWebUI, and anything else that
speaks the OpenAI or Gemini API.

## 📝 Notes on hosting

Verified on the Render free tier. If you ever hit `403`/`429` from Google's endpoint —
usually from a heavily-flagged shared IP range — set `IMPERSONATE` to a browser
fingerprint (`chrome_120`…) to present a real TLS handshake, which clears most cases.
The server refreshes Gemini's `bl` build id automatically at startup, so it keeps
working as Google rolls out new web builds.

## 📄 License

MIT — see [LICENSE](LICENSE).
