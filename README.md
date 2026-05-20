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

This repository has completed the **Phase 1 skeleton**, **Phase 2 payment DMN core**, **Phase 3 config/secret handling**, **Phase 4 credential verification**, and **Phase 5 target safety**.

## 📋 Planned Usage

```sh
nuvei-dmn-simulator config set-merchant local-demo
nuvei-dmn-simulator config set-target local
nuvei-dmn-simulator config list
nuvei-dmn-simulator config verify local-demo
nuvei-dmn-simulator send payment pix --profile local-demo --status APPROVED --target local
```

Config is stored outside the repository by default at your OS user config path, for example `~/.config/nuvei-dmn-simulator/config.toml` on macOS/Linux. `config set-merchant` prompts for the merchant secret without echoing it when run in a terminal, and `config list` always redacts stored secrets.

`config verify <profile>` contacts Nuvei `/getSessionToken` using the selected profile's configured `test` or `prod` environment. This proves the stored merchant credentials can authenticate with Nuvei without using `/payment`, without opening an order, and without printing the merchant secret or returned session token.

See `examples/config.example.toml` for a placeholder-only example.

## Quick Start: Config Smoke Test

Run these steps from the repository root to exercise the Phase 3 config commands without touching your real user config.

1. Create a temporary config path:

```sh
mkdir -p /tmp/nuvei-dmn-e2e
export NUVEI_DMN_CONFIG=/tmp/nuvei-dmn-e2e/config.toml
```

2. Add a merchant profile:

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

When run in a terminal, the merchant secret input is hidden while you type.

3. Add a local target profile:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" set-target local
```

Use these values when prompted:

```text
Target URL: http://localhost:3000/nuvei_direct_merchant_notifications
Target kind (local/staging/sandbox/demo/trusted/production-hosted-sandbox): local
Requires confirmation before send (true/false): false
```

4. Confirm secrets are redacted in command output:

```sh
go run ./cmd/nuvei-dmn-simulator config --config "$NUVEI_DMN_CONFIG" list
```

Expected output includes:

```toml
merchant_secret_key = "********"
```

It should not print `merchant-secret-key-placeholder`.

5. Clean up the temporary config when finished:

```sh
rm -rf /tmp/nuvei-dmn-e2e
```

To use your normal user config instead, omit `--config "$NUVEI_DMN_CONFIG"`. On macOS/Linux, the default path is typically `~/.config/nuvei-dmn-simulator/config.toml`.

`config verify` requires real Nuvei test or production credentials and will call Nuvei. `preview` and `send` are planned for later phases and are not expected to work yet.

## Target Safety

Target safety blocks unknown public URLs by default. Local development URLs are allowed when they use `localhost`, a loopback IP, `.test`, `.local`, or `.localhost`.

Public staging, sandbox, demo, or trusted testing hosts must be saved as target profiles with `config set-target <name>`. Production-hosted sandbox profiles, or any profile marked `requires_confirm = true`, require explicit confirmation before a future send operation proceeds. The future `--allow-untrusted-target` escape hatch is intended only for unknown public hosts; denied schemes and malformed URLs stay blocked.

## 💻 Development

### Requirements:

- Go 1.26 or newer

### Run tests:

```sh
go test ./...
```

## 📄 License

MIT
