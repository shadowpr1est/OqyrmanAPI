#!/bin/sh
set -eu

AUTH_DIR=/etc/nginx/auth
AUTH_FILE="$AUTH_DIR/monitoring.htpasswd"

if [ -z "${MONITORING_BASIC_AUTH_USER:-}" ] || [ -z "${MONITORING_BASIC_AUTH_PASSWORD:-}" ]; then
  echo "MONITORING_BASIC_AUTH_USER and MONITORING_BASIC_AUTH_PASSWORD must be set" >&2
  exit 1
fi

mkdir -p "$AUTH_DIR"
rm -f "$AUTH_FILE"
htpasswd -cbB "$AUTH_FILE" "$MONITORING_BASIC_AUTH_USER" "$MONITORING_BASIC_AUTH_PASSWORD" >/dev/null
chmod 640 "$AUTH_FILE"
