#!/bin/sh
PORT="${PORT:-8080}"
HOST="${HOST:-0.0.0.0}"

if [ -n "$API_KEYS" ]; then
  KEYS_JSON=$(echo "$API_KEYS" | tr ',' '\n' | sed 's/^/"/;s/$/"/' | paste -sd, -)
  KEYS_JSON="[$KEYS_JSON]"
else
  KEYS_JSON="[]"
fi

cat > /app/config.json <<EOC
{
  "port": ${PORT},
  "host": "${HOST}",
  "retry_attempts": 3,
  "retry_delay_sec": 2,
  "request_timeout_sec": 180,
  "default_model": "${DEFAULT_MODEL:-gemini-3.6-flash}",
  "api_keys": ${KEYS_JSON},
  "cookie_file": ${COOKIE_FILE:+"$COOKIE_FILE"}${COOKIE_FILE:-null},
  "image_model": "${IMAGE_MODEL:-gemini-image}",
  "image_model_id": ${IMAGE_MODEL_ID:-7},
  "image_backend": "${IMAGE_BACKEND:-gemini}",
  "local_image_url": "${LOCAL_IMAGE_URL:-}",
  "local_image_provider": "${LOCAL_IMAGE_PROVIDER:-alpha}",
  "local_image_model": "${LOCAL_IMAGE_MODEL:-flare}",
  "local_image_key": "${LOCAL_IMAGE_KEY:-}",
  "proxy": null,
  "impersonate": ${IMPERSONATE:+"$IMPERSONATE"}${IMPERSONATE:-null},
  "log_requests": true
}
EOC

exec /app/gemini-api -config /app/config.json
