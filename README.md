<!-- ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ -->

<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:4285F4,45:9B72CB,100:D96570&height=210&section=header&text=gemini-api&fontSize=74&fontColor=ffffff&fontAlignY=38&animation=fadeIn&desc=Gemini%20%E2%86%92%20OpenAI-%20%26%20Gemini-compatible%20gateway&descSize=17&descAlignY=60" width="100%" />

<a href="#-quick-start">
  <img src="https://readme-typing-svg.demolab.com?font=Fira+Code&weight=600&size=22&duration=2800&pause=700&color=9B72CB&center=true&vCenter=true&width=720&lines=Two+APIs%2C+one+binary%3A+OpenAI+%2B+native+Gemini;Streaming+%E2%80%A2+Vision+%E2%80%A2+Tools+%E2%80%A2+Reasoning+control;No+account%2C+no+token+%E2%80%94+deploy+and+go." alt="Typing SVG" />
</a>

<br/>

<p>
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/OpenAI_API-compatible-10a37f?style=for-the-badge&logo=openai&logoColor=white" />
  <img src="https://img.shields.io/badge/Gemini_API-compatible-4285F4?style=for-the-badge&logo=googlegemini&logoColor=white" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" />
</p>

<p>
  <img src="https://img.shields.io/badge/streaming-SSE-1f6feb?style=flat-square" />
  <img src="https://img.shields.io/badge/function_calling-yes-6E56CF?style=flat-square" />
  <img src="https://img.shields.io/badge/vision-yes-D96570?style=flat-square" />
  <img src="https://img.shields.io/badge/single_binary-static-brightgreen?style=flat-square" />
  <img src="https://img.shields.io/badge/Render-deploy-46E3B7?style=flat-square&logo=render&logoColor=white" />
</p>

</div>

<br/>

> **gemini-api** exposes Google Gemini's conversation endpoint as a standard
> **OpenAI-compatible** *and* **native Gemini (`v1beta`)** REST API — both served
> side by side from a single static Go binary. Works anonymously out of the box:
> no account, no token, nothing to configure. Add a cookie only for real Pro routing.

<div align="center">

`https://your-app.onrender.com/v1`  →  any OpenAI client &nbsp;•&nbsp; `…/v1beta`  →  any Gemini client

</div>

<br/>

<!-- ─────────────────────────────  CONTENTS  ───────────────────────────── -->

## 🧭 Contents

- [✨ Features](#-features)
- [⚡ Quick start](#-quick-start)
- [📡 API reference](#-api-reference)
- [🧪 Examples](#-examples)
- [🚀 Deploy](#-deploy)
- [⚙️ Configuration](#-configuration)
- [🍪 Pro models](#-pro-models)
- [🏗️ How it works](#-how-it-works)

<br/>

<!-- ─────────────────────────────  FEATURES  ───────────────────────────── -->

## ✨ Features

<table>
<tr>
<td width="50%" valign="top">

### 🔌 Two APIs, one server
OpenAI shape (`/v1/…`) **and** native Gemini shape (`/v1beta/…`) side by side.

### 🌊 Streaming
Token-by-token **SSE** (`"stream": true`) and `:streamGenerateContent`.

### 🛠️ Function calling
OpenAI `tools` and Gemini `functionDeclarations`, both supported.

### 👁️ Vision
Base64 data URLs, `image_url`, and inline Gemini `inlineData`.

</td>
<td width="50%" valign="top">

### 🧠 Reasoning control
`@think=N` suffix — `0` (deepest) → `4` (fastest).

### 📡 Responses API
`/v1/responses` — native support for Codex CLI.

### 🔁 Auto build-id refresh
Self-updates Gemini's upstream build id on startup; keeps working as Google ships.

### 🔐 Optional auth · 📦 Zero config
Bearer keys via `API_KEYS`; one static binary, no upstream account.

</td>
</tr>
</table>

<br/>

<!-- ────────────────────────────  QUICK START  ─────────────────────────── -->

## ⚡ Quick start

```bash
# build (Go 1.26+)
./build.sh
./gemini-api --port 8080

# or with Docker
docker compose up --build
```

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"ping"}]}'
```

<br/>

<!-- ───────────────────────────  API REFERENCE  ────────────────────────── -->

## 📡 API reference

<div align="center">

**OpenAI-compatible**

| Method | Endpoint | Description |
|:------:|:---------|:------------|
| `GET`  | `/` | Health / status |
| `GET`  | `/v1/models` | List models |
| `POST` | `/v1/chat/completions` | Chat completion + streaming |
| `POST` | `/v1/responses` | Responses API (Codex CLI native) |

**Native Gemini**

| Method | Endpoint | Description |
|:------:|:---------|:------------|
| `GET`  | `/v1beta/models` | List models (Gemini shape) |
| `POST` | `/v1beta/models/{model}:generateContent` | Generate |
| `POST` | `/v1beta/models/{model}:streamGenerateContent` | Stream |

</div>

### Models

<div align="center">

| Model | Notes |
|:------|:------|
| **`gemini-3.6-flash`** | All-around model *(default)* |
| `gemini-3.7-flash` | Latest all-around model |
| `gemini-3.5-flash` | Alias for `gemini-3.6-flash` |
| `gemini-3.5-flash-thinking` | Deep thinking, longest output |
| `gemini-3.5-flash-thinking-lite` | Dynamic, adaptive-depth thinking |
| `gemini-3.1-pro` | Pro model *(cookie needed for real routing)* |
| `gemini-3.1-pro-enhanced` | Pro, enhanced output *(experimental)* |
| `gemini-auto` | Automatic model selection |
| `gemini-flash-lite` | Lightweight, fastest |

</div>

> 💡 Append **`@think=N`** to any model to override reasoning depth —
> `gemini-3.6-flash@think=0` (deepest) … `@think=4` (fastest).

<br/>

<!-- ──────────────────────────────  EXAMPLES  ──────────────────────────── -->

## 🧪 Examples

<details open>
<summary><b>💬 Chat completion</b></summary>

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":"Capital of Bangladesh?"}]}'
```
</details>

<details>
<summary><b>🌊 Streaming</b></summary>

```bash
curl -N http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","stream":true,"messages":[{"role":"user","content":"Count to 5"}]}'
```
</details>

<details>
<summary><b>👁️ Vision</b></summary>

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash","messages":[{"role":"user","content":[
        {"type":"text","text":"Describe this image"},
        {"type":"image_url","image_url":{"url":"https://example.com/photo.jpg"}}
      ]}]}'
```
</details>

<details>
<summary><b>🛠️ Function calling</b></summary>

```bash
curl http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"gemini-3.6-flash",
       "messages":[{"role":"user","content":"Weather in Dhaka?"}],
       "tools":[{"type":"function","function":{"name":"get_weather",
         "parameters":{"type":"object","properties":{"city":{"type":"string"}}}}}]}'
```
</details>

<details>
<summary><b>🧩 Native Gemini shape</b></summary>

```bash
curl http://localhost:8080/v1beta/models/gemini-3.6-flash:generateContent \
  -H 'Content-Type: application/json' \
  -d '{"contents":[{"parts":[{"text":"ping"}]}]}'
```
</details>

<br/>

<!-- ──────────────────────────────  DEPLOY  ────────────────────────────── -->

## 🚀 Deploy

<div align="center">

**Render** · **Railway** · **Fly.io** · **Docker** · **any VPS**

</div>

Deploy on Render's free tier as a **Web Service** with the **Docker** runtime:

1. Push the repo to GitHub.
2. Render → **New +** → **Web Service** → pick the repo.
3. Runtime **Docker**, plan **Free**.
4. Add env vars below if needed → **Create**.

The container reads `$PORT` automatically and health-checks at `GET /` and `GET /v1/models`.

> ⚠️ **Vercel won't work** — this needs a long-lived process. Use the included Dockerfile
> on Render / Railway / a VPS.

<br/>

<!-- ──────────────────────────  CONFIGURATION  ─────────────────────────── -->

## ⚙️ Configuration

<div align="center">

| Variable | Default | Purpose |
|:---------|:-------:|:--------|
| `PORT` | `8080` (Docker) | Listen port (Render injects it) |
| `HOST` | `0.0.0.0` | Bind address |
| `API_KEYS` | *(empty)* | Comma-separated bearer tokens; empty = public |
| `DEFAULT_MODEL` | `gemini-3.6-flash` | Model used when a request omits one |
| `IMPERSONATE` | *(empty)* | TLS fingerprint (`chrome_120`…) for WAF-protected networks |

</div>

<br/>

<!-- ────────────────────────────  PRO MODELS  ──────────────────────────── -->

## 🍪 Pro models

Anonymous requests cover the Flash tiers. For real `gemini-3.1-pro` routing, supply a
signed-in session:

- point `--cookie-file` (or `COOKIE_FILE`) at your `gemini.google.com` cookies, and
- set `auth_user` / `xsrf_token` in `config.json` if your account needs them.

Without a cookie, Pro model names still respond but may be downgraded to a Flash tier.

<br/>

<!-- ────────────────────────────  HOW IT WORKS  ────────────────────────── -->

## 🏗️ How it works

```
┌────────────┐  OpenAI / Gemini  ┌──────────────┐   web protocol    ┌─────────┐
│  your app  │ ────────JSON─────▶ │  gemini-api  │ ────────────────▶ │ Gemini  │
│  OpenAI /  │ ◀───────JSON────── │   gateway    │ ◀──────────────── │  web    │
│  Gemini SDK│                    │  (Go, 1 bin) │  streamed tokens  └─────────┘
└────────────┘                    └──────┬───────┘
                                         │  auto-refreshes the upstream
                                         ▼  build id on startup
                                   keeps working across Google's releases
```

- **Dual protocol** — one server answers both the OpenAI and native Gemini API shapes.
- **Self-healing** — the upstream build id is refreshed automatically; no redeploys.
- **Lightweight** — a single static Go binary; talks to Gemini with a browser-grade TLS fingerprint, no headless browser or Chromium.

<br/>

<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=0:D96570,50:9B72CB,100:4285F4&height=120&section=footer" width="100%" />

<sub>Built with Go · MIT Licensed · Not affiliated with Google</sub>

</div>
