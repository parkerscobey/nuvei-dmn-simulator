#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d)"
CONFIG_PATH="$TMP_DIR/config.toml"

cd "$ROOT_DIR"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

cat >"$CONFIG_PATH" <<'EOF'
[merchants.local_demo]
environment = "test"
merchant_id = "merchant-id-placeholder"
merchant_site_id = "merchant-site-id-placeholder"
merchant_secret_key = "merchant-secret-key-placeholder"

[targets.local]
url = "http://localhost:3000/nuvei_direct_merchant_notifications"
kind = "local"
requires_confirm = false
EOF

echo "==> Preview smoke"
PREVIEW_OUTPUT="$(go run ./cmd/nuvei-dmn-simulator preview payment pix --config "$CONFIG_PATH" --profile local_demo --target local --status APPROVED)"

if [[ "$PREVIEW_OUTPUT" != *"Payload fields:"* ]]; then
  echo "preview output missing payload table"
  exit 1
fi
if [[ "$PREVIEW_OUTPUT" != *"Raw URL-encoded payload:"* ]]; then
  echo "preview output missing raw payload"
  exit 1
fi

echo "==> Local receiver send smoke"
go test ./cmd/nuvei-dmn-simulator -run '^TestSendPaymentPixPostsToTargetAndPrintsStatusAndBody$' -count=1

echo "local smoke completed"
