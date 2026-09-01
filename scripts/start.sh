#!/bin/sh
set -eu

TAILSCALE_SOCKET="${TAILSCALE_SOCKET:-/var/run/tailscale/tailscaled.sock}"
TAILSCALE_STATE="${TAILSCALE_STATE:-/data/tailscaled.state}"
TAILSCALE_SERVE_CONFIG="${TAILSCALE_SERVE_CONFIG:-/app/config/tailscale-services.json}"
GATEWAY_PID=""
TAILSCALED_PID=""

if [ -z "${TAILSCALE_AUTHKEY:-}" ]; then
  echo "TAILSCALE_AUTHKEY is required. Set it with: fly secrets set TAILSCALE_AUTHKEY=tskey-..." >&2
  exit 1
fi

mkdir -p "$(dirname "$TAILSCALE_SOCKET")" "$(dirname "$TAILSCALE_STATE")"

/app/tailscaled \
  --state="$TAILSCALE_STATE" \
  --socket="$TAILSCALE_SOCKET" &
TAILSCALED_PID=$!

cleanup() {
  if [ -n "$GATEWAY_PID" ]; then
    kill "$GATEWAY_PID" 2>/dev/null || true
    wait "$GATEWAY_PID" 2>/dev/null || true
  fi
  if [ -n "$TAILSCALED_PID" ]; then
    kill "$TAILSCALED_PID" 2>/dev/null || true
    wait "$TAILSCALED_PID" 2>/dev/null || true
  fi
}
trap cleanup INT TERM EXIT

# Give tailscaled a moment to create its local socket.
i=0
while [ ! -S "$TAILSCALE_SOCKET" ]; do
  i=$((i + 1))
  if [ "$i" -gt 100 ]; then
    echo "tailscaled did not create its socket" >&2
    exit 1
  fi
  sleep 0.1
done

set -- /app/tailscale --socket="$TAILSCALE_SOCKET" up \
  --auth-key="$TAILSCALE_AUTHKEY" \
  --hostname="${TS_HOSTNAME:-shadow-collective}" \
  --accept-dns=false

if [ -n "${TS_TAGS:-}" ]; then
  set -- "$@" --advertise-tags="$TS_TAGS"
fi

"$@"

/app/shadow-collective -config "${GATEWAY_CONFIG:-/app/config/services.json}" &
GATEWAY_PID=$!

# Tailscale Services are optional. An empty services object is a no-op.
if [ -f "$TAILSCALE_SERVE_CONFIG" ]; then
  service_count="$(jq '.services | length' "$TAILSCALE_SERVE_CONFIG")"
  if [ "$service_count" -gt 0 ]; then
    /app/tailscale --socket="$TAILSCALE_SOCKET" serve set-config --all "$TAILSCALE_SERVE_CONFIG"
    jq -r '.services | keys[]' "$TAILSCALE_SERVE_CONFIG" | while IFS= read -r service; do
      if ! /app/tailscale --socket="$TAILSCALE_SOCKET" serve advertise "$service"; then
        echo "warning: could not advertise $service; create/approve the Service in Tailscale and verify tag permissions" >&2
      fi
    done
  fi
fi

wait "$GATEWAY_PID"
