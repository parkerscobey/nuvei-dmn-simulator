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

This repository has completed the **Phase 1 skeleton** and **Phase 2 payment DMN core**.

## 📋 Planned Usage

```sh
nuvei-dmn-simulator config set-merchant local-demo
nuvei-dmn-simulator config verify local-demo
nuvei-dmn-simulator send payment pix --profile local-demo --status APPROVED --target local
```

## 💻 Development

### Requirements:

- Go 1.26 or newer

### Run tests:

```sh
go test ./...
```

## 📄 License

MIT