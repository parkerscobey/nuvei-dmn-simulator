# Nuvei DMN Simulator Vision

`nuvei-dmn-simulator` is an open source development and QA tool for generating and sending signed Nuvei Direct Merchant Notification (DMN) webhook payloads.

Nuvei DMNs are server-to-server webhook callbacks used by asynchronous payment methods such as Pix, Boleto, and other APMs. These callbacks are hard to trigger on demand in local development and staging. The simulator exists to make those flows testable without relying on Nuvei to emit real asynchronous events.

## Goals

- Generate realistic Nuvei payment DMN payloads.
- Support asynchronous APM status flows: `PENDING`, `APPROVED`, and `DECLINED` first.
- Compute valid Nuvei `advanceResponseChecksum` values.
- Prove merchant credentials are valid with Nuvei before sending simulated DMNs.
- Provide both a CLI and a local web server UI.
- Keep the simulator independent from any one merchant application.
- Make Pix the first supported APM while keeping the design expandable to Boleto and other APMs.
- Be safe by default and make accidental production misuse difficult.

## Non-Goals

- The simulator does not process real payments.
- The simulator does not replace Nuvei.
- The simulator does not bypass merchant webhook authentication.
- The simulator does not send unsigned DMNs.
- The simulator does not require knowledge of Back Office internals or any other merchant database.
- The simulator does not store production merchant credentials in committed files.

## Safety Model

The simulator requires real Nuvei merchant credentials before it can send any simulated DMN:

- `merchant-id`
- `merchant-site-id`
- `merchant-secret-key`

The simulator must verify those credentials against Nuvei before any send operation. A user who can verify all three values with Nuvei already has the credential material required to generate valid signed DMNs manually. The simulator should still make unsafe usage difficult, but credential verification is the primary boundary.

Target URL safety remains important. The simulator should refuse or warn on unknown production-looking targets by default, while still allowing explicit configuration for legitimate staging, sandbox, demo, or production-hosted sandbox gateway environments.

## Credential Handling

Merchant credentials should not be passed as command-line flags because command invocations are commonly persisted in shell history.

The CLI should provide an interactive `config set-merchant` command that reads secret values from prompts. The web UI should accept credentials through form fields or selected profiles. Local config files may be supported, but examples must never contain real secrets.

## Product Shape

The project should have one shared Go core library and two interfaces:

- CLI for developers and automation.
- Local web server UI for QA and less technical testers.

The web UI should be server-rendered Go templates enhanced with HTMX. The first UI can be simple: one form, payload preview, and send buttons.

## Reference Documentation

- Nuvei Webhooks / DMNs: https://docs.nuvei.com/documentation/integration/webhooks/
- Nuvei Payment DMNs: https://docs.nuvei.com/documentation/integration/webhooks/payment-dmns/
- Nuvei Web SDK openOrder API: https://docs.nuvei.com/api/main/indexMain_v1_0.html?json#openOrder
