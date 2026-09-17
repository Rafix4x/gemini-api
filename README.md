<div align="center">

# ⚡ gemini-api

**Gemini Web → OpenAI-compatible API gateway**

A lightweight Go server that exposes Google Gemini's web conversation endpoint
as a standard **OpenAI-compatible REST API**. Single static binary — no account,
no tokens, nothing to configure.

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com)
[![Render](https://img.shields.io/badge/Render-deploy%20ready-46E3B7?logo=render&logoColor=white)](https://render.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen)](#)

</div>

---

## ✨ Features

| | |
|---|---|
| 🔌 **OpenAI-compatible** | Drop-in base URL for any OpenAI SDK or agent (Cline, Codex CLI, OpenWebUI…) |
| 🌊 **Streaming** | Server-Sent Events with `stream: true` |
| 🛠️ **Function calling** | OpenAI tool-call schema |
| 👁️ **Vision** | Base64 and `image_url` image inputs |
| 🧠 **Reasoning control** | `@think=N` suffix — `0` (deep) → `4` (fast) |
| 📡 **Responses API** | Native support for Codex CLI |
| 🔐 **Optional auth** | Bearer keys via env; open by default |
| 📦 **Zero config** | Single static binary, no upstream account or tokens |

## 📡 API

### Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/v1/chat/completions` | Chat completion (OpenAI-compatible) |
| `POST` | `/v1/responses` | Responses API (Codex CLI native) |
| `GET` | `/v1/models` | List available models |

### Models

| Model | Notes |
|---|---|
| `gemini-3.7-flash` | Latest all-around model |
| `gemini-3.6-flash` | All-around model (default) |
| `gemini-3.5-flash-thinking` | Deep thinking mode, ~20k char output |
| `gemini-3.5-flash-thinking-lite` | Dynamic thinking, adaptive depth |
| `gemini-3.1-pro` | Pro model (requires cookie for real routing) |
| `gemini-3.1-pro-enhanced` | Pro with enhanced output (experimental) |
| `gemini-auto` | Automatic model selection |
| `gemini-flash-lite` | Lightweight, fastest |
| `gemini-3.5-flash` | Alias for `gemini-3.6-flash` |

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

Streaming works the same way — set `"stream": true` and consume SSE chunks
ending with `data: [DONE]`.

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

The container listens on `$PORT` automatically and answers health checks at `/v1/models`.

### Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port (Render injects it) |
| `HOST` | `0.0.0.0` | Bind address |
| `API_KEYS` | *(empty)* | Comma-separated bearer tokens; empty = public |
| `DEFAULT_MODEL` | `gemini-3.6-flash` | Model used when the request omits one |
| `IMPERSONATE` | *(empty)* | Chrome/Edge TLS fingerprint (`chrome_120`…) for WAF-protected networks |

## 🔗 Connecting clients

Point any OpenAI SDK or agent at `http://localhost:8080/v1` (or your Render URL):

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://your-render-url.onrender.com/v1",
    api_key="not-needed",
)

resp = client.chat.completions.create(
    model="gemini-3.6-flash",
    messages=[{"role": "user", "content": "Hello!"}],
)
print(resp.choices[0].message.content)
```

## 📝 Notes on hosting

Google applies risk-based filtering on its web endpoint. Requests from
residential networks pass through routinely (verified on Android/Termux).
Traffic from cloud/datacenter ranges may occasionally trigger `403`/`429`;
setting `IMPERSONATE` to a browser fingerprint resolves most cases.

## 📄 License

MIT — see [LICENSE](LICENSE).

