#!/bin/sh
set -eu

PORT="${PORT:-80}"
HOST="${HOST:-0.0.0.0}"
export PORT HOST
umask 077

app_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
config_path=$(mktemp "${TMPDIR:?TMPDIR must point to writable scratch storage}/gemini-api-config.XXXXXX")

if ! jq -n '
  def setting($name; $default):
    (env[$name] // "") | if . == "" then $default else . end;
  def integer($name; $default; $min; $max):
    setting($name; $default) | tonumber |
    if floor == . and . >= $min and . <= $max then .
    else error($name + " must be an integer in the supported range") end;
  {
    port: integer("PORT"; "80"; 1; 65535),
    host: setting("HOST"; "0.0.0.0"),
    retry_attempts: 3,
    retry_delay_sec: 2,
    request_timeout_sec: 180,
    default_model: setting("DEFAULT_MODEL"; "gemini-3.5-flash-lite"),
    api_keys: (setting("API_KEYS"; "") | split(",") | map(select(length > 0))),
    cookie_file: setting("COOKIE_FILE"; null),
    image_model: setting("IMAGE_MODEL"; "gemini-image"),
    image_model_id: integer("IMAGE_MODEL_ID"; "7"; 0; 2147483647),
    image_backend: setting("IMAGE_BACKEND"; "local"),
    local_image_url: setting("LOCAL_IMAGE_URL"; "https://image-gen-ebon.vercel.app"),
    local_image_provider: setting("LOCAL_IMAGE_PROVIDER"; "alpha"),
    local_image_model: setting("LOCAL_IMAGE_MODEL"; "flare"),
    local_image_key: setting("LOCAL_IMAGE_KEY"; ""),
    proxy: null,
    impersonate: setting("IMPERSONATE"; null),
    log_requests: true
  }
' > "$config_path"; then
  rm -f "$config_path"
  exit 1
fi

exec "$app_dir/gemini-api" -config "$config_path"
