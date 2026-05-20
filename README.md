# nuvei-dmn-simulator

[![CI](https://github.com/parkerscobey/nuvei-dmn-simulator/actions/workflows/ci.yml/badge.svg)](https://github.com/parkerscobey/nuvei-dmn-simulator/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26-00ADD8?logo=go&logoColor=white)](go.mod)

Safe-by-default Go CLI for previewing and sending signed Nuvei Direct Merchant Notification (DMN) webhooks, with a local HTMX UI planned next for QA and less technical testers.

Nuvei exposes DMNs, but does not provide a Stripe CLI-style developer tool for triggering them on demand. That makes local development, staging verification, and regression testing harder than it should be. `nuvei-dmn-simulator` exists to close that gap.

Inspired in spirit by Stripe CLI, but focused on Nuvei DMN workflows, merchant credential safety, and a path to a simple QA-friendly local UI.

## Why this exists

- Trigger realistic signed Nuvei payment DMNs without waiting on real asynchronous payment events.
- Give developers a fast CLI loop for previewing and sending webhook payloads.
- Give QA a path to a simple local UI built on the same core primitives.
- Make safe behavior the default, not an afterthought.
- Create a foundation that can expand from Pix into Boleto and other Nuvei DMN/APM flows.

## What works today

Current implemented capabilities:

- Merchant profile configuration stored outside the repository.
- Target profile configuration for local, staging, sandbox, demo, and other trusted test endpoints.
- Real Nuvei credential verification through `/getSessionToken`.
- Pix payment DMN preview with signed payload generation.
- Pix payment DMN send flow with target safety checks.
- Strict correlation mode via `--require-correlation-fields`.
- Readable payload preview plus raw `application/x-www-form-urlencoded` output.

## Product shape

The long-term shape is one shared Go core with two interfaces:

- **CLI** for developers and automation.
- **Local web UI** for QA and less technical team members.

The web UI is planned as server-rendered Go templates plus HTMX, reusing the same payload builders, safety rules, and send flow as the CLI.

## Install

Install the CLI with Go:

```sh
go install github.com/parkerscobey/nuvei-dmn-simulator/cmd/nuvei-dmn-simulator@latest
```

Or clone the repo and run it directly:

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
```

That flow gives you:

1. a saved merchant profile
2. a saved target profile
3. proof that the configured Nuvei credentials are valid
4. a preview of the exact signed payload that would be sent
5. a real POST to a trusted target when you are ready

## Quick start

Use a temporary config path so you can test without touching your normal user config.

1. Create a temporary config location:

```sh
mkdir -p /tmp/nuvei-dmn-e2e
export NUVEI_DMN_CONFIG=/tmp/nuvei-dmn-e2e/config.toml
```

2. Add a placeholder merchant profile:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-merchant local-demo
```

Use placeholder values when prompted:

```text
Nuvei environment (test/prod): test
Merchant ID: merchant-id-placeholder
Merchant Site ID: merchant-site-id-placeholder
Merchant Secret Key: merchant-secret-key-placeholder
```

3. Add a local target profile:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-target local
```

Example values:

```text
Target URL: http://localhost:3000/nuvei_direct_merchant_notifications
Target kind (local/staging/sandbox/demo/trusted/production-hosted-sandbox): local
Requires confirmation before send (true/false): false
```

4. Confirm secrets are redacted:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" list
```

Expected output includes:

```toml
merchant_secret_key = "********"
```

5. Preview a signed Pix DMN without making a network call:

```sh
go run ./cmd/nuvei-dmn-simulator preview payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --target local \
  --status APPROVED
```

Expected result: target summary, payload field table, and raw URL-encoded payload. No Nuvei verification call and no POST to the target.

6. Exercise strict correlation mode:

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

Strict mode fails fast unless the required correlation fields are explicitly provided.

7. Confirm untrusted public targets are blocked by default:

```sh
go run ./cmd/nuvei-dmn-simulator send payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --status APPROVED \
  --target https://example.com/nuvei_direct_merchant_notifications
```

Expected result: blocked as untrusted.

8. Run a real send test with valid Nuvei credentials and a local receiver:

```sh
go run ./cmd/nuvei-dmn-simulator send payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --status APPROVED \
  --target local
```

Expected result: the command prints receiver HTTP status and response body.

9. Clean up:

```sh
rm -rf /tmp/nuvei-dmn-e2e
```

## Safety model

This project is intentionally strict.

- It never sends unsigned DMNs.
- It verifies merchant credentials with Nuvei before real send operations.
- It does not accept merchant secrets as normal CLI flags in the primary UX.
- It does not use `/payment` or another money-moving endpoint for credential verification.
- It blocks unknown public targets by default unless explicitly overridden.
- It keeps example config files placeholder-only and redacts stored secrets in command output.

The goal is not just convenience. The goal is safe, repeatable test tooling for a messy integration surface.

## Roadmap

Near-term roadmap:

- Raw payload mode for replaying and editing sanitized DMN samples.
- HTMX local web UI for QA workflows.
- Boleto support and additional Nuvei APM / DMN types.
- Release packaging for easier installation outside a Go dev environment.
- Demo assets such as screenshots, GIFs, and sanitized walkthroughs once the UI is ready.
- A local end-to-end smoke path, including a tiny example receiver and CI-friendly sanitized demo flow.

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

See `VISION.md`, `SPEC.md`, and `TODO.md` for the broader product direction.

## License

MIT
