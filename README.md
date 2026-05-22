# nuvei-dmn-simulator

[![CI](https://github.com/parkerscobey/nuvei-dmn-simulator/actions/workflows/ci.yml/badge.svg)](https://github.com/parkerscobey/nuvei-dmn-simulator/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26-00ADD8?logo=go&logoColor=white)](go.mod)

`nuvei-dmn-simulator` is a safe-by-default Go CLI for previewing and sending signed Nuvei Direct Merchant Notification (DMN) webhook payloads.

## Why this exists

Teams integrating Nuvei DMNs often need deterministic webhook testing without waiting on real asynchronous payment events. This tool provides a repeatable local workflow for validating payload shape, signature behavior, target safety rules, and send handling.

## Safety model

- Never send unsigned DMNs.
- Verify merchant credentials with Nuvei before real send operations.
- Never accept merchant secrets as standard CLI flags in the primary UX.
- Never commit or document real merchant credentials.
- Never use `/payment` or another money-moving endpoint for credential verification.
- Block unknown public targets by default unless explicitly overridden.

## What works today

- Merchant profile configuration stored outside the repository.
- Target profile configuration for local, staging, sandbox, demo, and other trusted test endpoints.
- Nuvei credential verification through `/getSessionToken`.
- Pix payment DMN preview with signed payload generation.
- Pix payment DMN send flow with target safety checks.
- Strict correlation mode via `--require-correlation-fields`.
- Raw URL-encoded payload import with safe override + checksum recompute.
- Local web UI with HTMX actions for verify, preview, and send.

## Install

```sh
go install github.com/parkerscobey/nuvei-dmn-simulator/cmd/nuvei-dmn-simulator@latest
```

Or run from source:

```sh
git clone https://github.com/parkerscobey/nuvei-dmn-simulator.git
cd nuvei-dmn-simulator
go run ./cmd/nuvei-dmn-simulator --help
```

## Example workflow

```sh
nuvei-dmn-simulator config set-merchant local-demo
nuvei-dmn-simulator config set-target local
nuvei-dmn-simulator config verify local-demo
nuvei-dmn-simulator preview payment pix --profile local-demo --status APPROVED --target local
nuvei-dmn-simulator send payment pix --profile local-demo --status APPROVED --target local
nuvei-dmn-simulator preview payment from-raw --profile local-demo --file payload.txt --status APPROVED --target local
nuvei-dmn-simulator send payment from-raw --profile local-demo --file payload.txt --status DECLINED --target local
```

Config is stored outside the repository by default at your OS user config path, for example `~/.config/nuvei-dmn-simulator/config.toml` on macOS/Linux. `config set-merchant` prompts for the merchant secret without echoing it when run in a terminal, and `config list` always redacts stored secrets.

`config verify <profile>` calls Nuvei `/getSessionToken` for the profile's configured `test` or `prod` environment. This verifies credentials without using `/payment`, without opening an order, and without printing merchant secrets or returned session tokens.

See `examples/config.example.toml` for a placeholder-only example.

## Quick start

Use a temporary config path so you can test without touching your normal user config.

1. Create a temporary config path.

```sh
mkdir -p /tmp/nuvei-dmn-e2e
export NUVEI_DMN_CONFIG=/tmp/nuvei-dmn-e2e/config.toml
```

2. Add placeholder merchant and local target profiles.

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-merchant local-demo
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-target local
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" list
```

Expected: `merchant_secret_key = "********"` appears instead of the raw secret value.

3. Preview a payload (network-free).

```sh
go run ./cmd/nuvei-dmn-simulator preview payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --target local \
  --status APPROVED
```

Expected: target summary, payload field table, and raw URL-encoded payload. No Nuvei call and no target POST.

4. Exercise strict correlation mode.

```sh
go run ./cmd/nuvei-dmn-simulator preview payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --target local \
  --require-correlation-fields \
  --status APPROVED \
  --total-amount 42.10 \
  --currency BRL \
  --client-request-id req-123 \
  --client-unique-id uniq-123 \
  --user-payment-option-id upo-123
```

Strict mode fails fast unless all required fields are explicitly passed.

5. Confirm target safety behavior (no real credentials required).

```sh
go run ./cmd/nuvei-dmn-simulator send payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --status APPROVED \
  --target https://example.com/nuvei_direct_merchant_notifications
```

Expected: blocked as untrusted by default.

6. Run a full send test (real Nuvei credentials required).

- Update `local-demo` with valid Nuvei credentials.
- Run a local webhook receiver.

```sh
go run ./cmd/nuvei-dmn-simulator send payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --status APPROVED \
  --target local
```

Expected: command prints receiver HTTP status and response body.

## Raw payload mode

Use raw mode when you want to start from an existing URL-encoded Nuvei payment DMN payload (for example, docs samples or sanitized real payloads).

```sh
go run ./cmd/nuvei-dmn-simulator preview payment from-raw \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --file ./payload.txt \
  --target local \
  --status APPROVED
```

Behavior:

- Parses the URL-encoded payload from `--file`.
- Replaces `merchant_id` and `merchant_site_id` from the selected profile by default.
- Applies optional field overrides (`--status`, amount/currency, correlation IDs, reason fields).
- Recomputes `advanceResponseChecksum` after overrides.

If you need to preserve merchant identifiers from the file, pass `--keep-raw-merchant-fields`.

## Web UI

Prerequisite: the web UI reads merchant and target profiles from the same local config file used by CLI commands. Set up profiles first:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-merchant local-demo
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-target local
```

Run the local UI server:

```sh
go run ./cmd/nuvei-dmn-simulator server --config "$NUVEI_DMN_CONFIG"
```

Defaults: binds to `127.0.0.1:8080`. Override with `--host` and `--port` if needed.

The web UI uses the same core logic as CLI send/preview:

- Merchant profile and target resolution from config.
- Pix payload building and checksum generation.
- Target safety classification and allow/untrusted controls.
- Nuvei `/getSessionToken` verification before send.
- Form-encoded DMN sending with response status/body display.

Phase 9 media artifacts:

- Screenshot: `docs/media/phase-9-web-ui.png`
- GIF: `docs/media/phase-9-verify-preview-send.gif`

7. Clean up.

```sh
rm -rf /tmp/nuvei-dmn-e2e
```

## Target safety

Target safety blocks unknown public URLs by default. Local development URLs are allowed when they use `localhost`, a loopback IP, `.test`, `.local`, or `.localhost`.

Public staging, sandbox, demo, or trusted testing hosts must be saved as target profiles with `config set-target <name>`. Production-hosted sandbox profiles, or any profile marked `requires_confirm = true`, require explicit confirmation before send proceeds. The `--allow-untrusted-target` escape hatch is intended only for unknown public hosts; denied schemes and malformed URLs stay blocked.

## Reference documentation

- Nuvei Webhooks / DMNs: https://docs.nuvei.com/documentation/integration/webhooks/
- Nuvei Payment DMNs: https://docs.nuvei.com/documentation/integration/webhooks/payment-dmns/
- Nuvei `getSessionToken`: https://docs.nuvei.com/api/main/indexMain_v1_0.html?json#getSessionToken
- Nuvei `openOrder`: https://docs.nuvei.com/api/main/indexMain_v1_0.html?json#openOrder

## Development

Requirements:

- Go 1.26 or newer

Run tests:

```sh
go test ./...
```

Run local smoke:

```sh
./scripts/local_smoke.sh
```

The smoke script is fully local and uses sanitized placeholder values. It validates preview output and exercises the local receiver send flow through a targeted CLI test.

Optional credentialed integration:

- GitHub Actions includes a separate manual workflow at `.github/workflows/integration.yml`.
- It runs only when repository secrets are present and verifies real Nuvei credentials before sending to a local receiver.

See `VISION.md`, `SPEC.md`, and `TODO.md` for broader direction.

## License

MIT
