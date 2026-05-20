# 🚀 nuvei-dmn-simulator

`nuvei-dmn-simulator` is an open source Go tool for generating and sending signed Nuvei Direct Merchant Notification (DMN) webhook payloads.

## ✨ Overview

Intended for developers and QA teams testing merchant webhook integrations for asynchronous payment methods such as **Pix** and **Boleto**.

## ⚠️ Safety Warning

This tool must be safe by default:

- 🔒 It must **never** send unsigned DMNs.
- 🔐 It must verify merchant credentials with Nuvei before sending simulated DMNs.
- 🚫 It must not accept merchant secrets as normal command-line flags in the primary UX.
- 📝 It must not commit or document real merchant credentials.
- 💳 It must not use `/payment` or another money-moving endpoint for credential verification.
- 🛡️ It must block unknown public targets by default unless explicitly trusted or loudly overridden.

## 📊 Status

This repository has completed the **Phase 1 skeleton**, **Phase 2 payment DMN core**, **Phase 3 config/secret handling**, **Phase 4 credential verification**, **Phase 5 target safety**, and **Phase 6 CLI preview/send**.

## 📋 Usage

```sh
nuvei-dmn-simulator config set-merchant local-demo
nuvei-dmn-simulator config set-target local
nuvei-dmn-simulator config list
nuvei-dmn-simulator config verify local-demo
nuvei-dmn-simulator preview payment pix --profile local-demo --status APPROVED --target local
nuvei-dmn-simulator send payment pix --profile local-demo --status APPROVED --target local
```

Config is stored outside the repository by default at your OS user config path, for example `~/.config/nuvei-dmn-simulator/config.toml` on macOS/Linux. `config set-merchant` prompts for the merchant secret without echoing it when run in a terminal, and `config list` always redacts stored secrets.

`config verify <profile>` contacts Nuvei `/getSessionToken` using the selected profile's configured `test` or `prod` environment. This proves the stored merchant credentials can authenticate with Nuvei without using `/payment`, without opening an order, and without printing the merchant secret or returned session token.

See `examples/config.example.toml` for a placeholder-only example.

## Quick Start

Use a temporary config so you can test without touching your real user config.

1. Create temp config path:

```sh
mkdir -p /tmp/nuvei-dmn-e2e
export NUVEI_DMN_CONFIG=/tmp/nuvei-dmn-e2e/config.toml
```

2. Add placeholder merchant + local target:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-merchant local-demo
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-target local
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" list
```

Expected: `merchant_secret_key = "********"` is shown, not the raw secret.

3. Preview payload (network-free):

```sh
go run ./cmd/nuvei-dmn-simulator preview payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --target local \
  --status APPROVED
```

Expected: target summary + payload field table + raw URL-encoded payload; no Nuvei call and no target POST.

4. Strict correlation mode (required matching fields):

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

5. Target safety check (no real creds required):

```sh
go run ./cmd/nuvei-dmn-simulator send payment pix \
  --config "$NUVEI_DMN_CONFIG" \
  --profile local-demo \
  --status APPROVED \
  --target https://example.com/nuvei_direct_merchant_notifications
```

Expected: blocked as untrusted by default.

6. Full send test (real Nuvei credentials required):

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

7. Clean up:

```sh
rm -rf /tmp/nuvei-dmn-e2e
```

## Target Safety

Target safety blocks unknown public URLs by default. Local development URLs are allowed when they use `localhost`, a loopback IP, `.test`, `.local`, or `.localhost`.

Public staging, sandbox, demo, or trusted testing hosts must be saved as target profiles with `config set-target <name>`. Production-hosted sandbox profiles, or any profile marked `requires_confirm = true`, require explicit confirmation before send proceeds. The `--allow-untrusted-target` escape hatch is intended only for unknown public hosts; denied schemes and malformed URLs stay blocked.

## 💻 Development

### Requirements:

- Go 1.26 or newer

### Run tests:

```sh
go test ./...
```

## 📄 License

MIT
