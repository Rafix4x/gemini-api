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
  "cookie_file": null,
  "proxy": null,
  "impersonate": ${IMPERSONATE:+"$IMPERSONATE"}${IMPERSONATE:-null},
  "log_requests": true
}
EOC

exec /app/gemini-api -config /app/config.json
