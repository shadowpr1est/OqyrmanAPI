#!/bin/sh
set -eu

AUTH_DIR=/etc/nginx/auth
AUTH_FILE="$AUTH_DIR/monitoring.htpasswd"
AUTH_USER="${MONITORING_BASIC_AUTH_USER:-monitoring}"
AUTH_PASSWORD="${MONITORING_BASIC_AUTH_PASSWORD:-}"

if [ -z "$AUTH_PASSWORD" ]; then
  AUTH_PASSWORD="$(head -c 48 /dev/urandom | base64 | tr -dc 'A-Za-z0-9' | head -c 20)"
  echo "MONITORING_BASIC_AUTH_PASSWORD is empty. Temporary password generated for user '$AUTH_USER'." >&2
  echo "Set MONITORING_BASIC_AUTH_PASSWORD in .env to a permanent value." >&2
  echo "Temporary monitoring password: $AUTH_PASSWORD" >&2
fi

mkdir -p "$AUTH_DIR"
rm -f "$AUTH_FILE"
htpasswd -cbB "$AUTH_FILE" "$AUTH_USER" "$AUTH_PASSWORD" >/dev/null
chmod 644 "$AUTH_FILE"
