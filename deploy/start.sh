#!/bin/sh
set -eu

HERMES_WEBUI_HOST=127.0.0.1 /usr/local/bin/hermes-web-studio &
backend_pid=$!
trap 'kill "$backend_pid" 2>/dev/null || true; wait "$backend_pid" 2>/dev/null || true' INT TERM EXIT

nginx -g 'daemon off;'
