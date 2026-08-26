# SPEC: nuvei-dmn-simulator

## Overview

Build a standalone Go application named `nuvei-dmn-simulator` that generates and sends signed Nuvei Direct Merchant Notification (DMN) payloads.

The simulator must not depend on Back Office or any merchant application. It should understand Nuvei DMN formats, Nuvei checksum rules, target safety rules, and credential verification.

The first supported DMN type is Payment DMN for APM transactions, beginning with Pix. The architecture must allow Boleto and other APMs to be added without rewriting the CLI or web server.

Module path: `github.com/parkerscobey/nuvei-dmn-simulator`.

## Stack

- Language: Go 1.26.
- CLI: Cobra.
- HTTP client/server: Go standard library.
- Web UI: Go `html/template` plus HTMX.
- Config format: TOML.
- Tests: Go standard testing package.
- License: MIT.

Keep third-party dependencies minimal. The tool should be easy to audit and easy to distribute as a single binary.

## Nuvei Documentation References

- Webhooks / DMNs: https://docs.nuvei.com/documentation/integration/webhooks/
- Payment DMNs: https://docs.nuvei.com/documentation/integration/webhooks/payment-dmns/
- getSessionToken API: https://docs.nuvei.com/api/main/indexMain_v1_0.html?json#getSessionToken
- openOrder API: https://docs.nuvei.com/api/main/indexMain_v1_0.html?json#openOrder

## Credential Verification Requirement

The simulator must prove merchant credentials are valid with Nuvei before sending simulated DMNs.

Required credential fields:

- `merchant-id`
- `merchant-site-id`
- `merchant-secret-key`

Credential verification must be a real Nuvei API call. Local checksum generation alone is not sufficient.

### Preferred Verification

Use Nuvei `/getSessionToken` as the preferred credential verification endpoint. Nuvei documents this method as receiving merchant authentication details and returning a `sessionToken`; it is recommended for pure API flows and does not open an order or send a payment request.

Known side effect: a successful verification issues a Nuvei session token for the supplied merchant/site credentials. The simulator stores this token only in memory for the current process and does not print it.

### Initial Acceptable Verification

If no read-only endpoint is available, use Nuvei `openOrder` with a minimal non-money-moving request to prove credentials can authenticate with Nuvei and produce a session token.

Initial `openOrder` verification constraints:

- Use the configured Nuvei environment (`test` / `prod`) explicitly.
- Use a generated `clientRequestId` and `clientUniqueId`.
- Use a synthetic `userTokenId` such as `nuvei-dmn-simulator-verify`.
- Prefer zero amount if Nuvei supports it for the merchant/site.
- Include a clearly synthetic item name if items are required.
- Do not perform `/payment` as credential verification.
- Do not create Pix, Boleto, payout, refund, void, or any money-moving transaction as credential verification.
- Document any side effects of `openOrder` verification.

The simulator should cache successful credential verification in memory for the current process so users are not forced to call Nuvei before every single payload preview or send inside the same CLI/server session. Sending should still require that credentials have been verified in the current process or that the send flow verifies them before sending.

## Credential Storage and Configuration

Do not require secrets on the command line.

Commands should be shaped like this:

```sh
nuvei-dmn-simulator config set-merchant local-demo
nuvei-dmn-simulator config verify local-demo
nuvei-dmn-simulator send payment pix --profile local-demo --status APPROVED --target http://localhost:3000/nuvei_direct_merchant_notifications
```

`config set-merchant` should prompt interactively for:

- Nuvei environment: `test` or `prod`
- Merchant ID
- Merchant Site ID
- Merchant Secret Key

Local config should be stored outside the repository by default, for example:

- macOS/Linux: `~/.config/nuvei-dmn-simulator/config.toml`
- Windows: `%AppData%\\nuvei-dmn-simulator\\config.toml`

Secret storage options, in order of preference:

1. OS keychain integration if implemented.
2. Local config file with restrictive file permissions.
3. Environment variables for automation only.

The repository may include example config files, but they must contain placeholders only.

## Target Safety

The simulator should protect against accidental misuse but still allow legitimate sandbox/demo gateways that happen to run behind production application hosts.

Targets should be classified as:

- Allowed by default: `localhost`, `127.0.0.1`, `.test`, `.local`.
- Allowed by profile: hosts explicitly added by the user as staging, sandbox, demo, or trusted testing targets.
- Blocked by default: unknown public hosts.
- Denied always: hosts on an explicit denylist.

Configuration should support:

```toml
[targets.local]
url = "http://localhost:3000/nuvei_direct_merchant_notifications"
kind = "local"

[targets.staging]
url = "https://staging.example.com/nuvei_direct_merchant_notifications"
kind = "staging"

[targets.prod_sandbox_gateway]
url = "https://app.example.com/nuvei_direct_merchant_notifications"
kind = "production-hosted-sandbox"
requires_confirm = true
```

Unknown public targets should require an explicit one-time confirmation or profile entry. A command-line escape hatch may exist, but it must be loud:

```sh
nuvei-dmn-simulator send payment pix --profile local-demo --target https://example.com/dmn --allow-untrusted-target
```

## Payment DMN Checksum

For payment DMNs, compute `advanceResponseChecksum` using Nuvei’s documented fields:

1. `totalAmount`
2. `currency`
3. `responseTimeStamp`
4. `PPP_TransactionID`
5. `Status`
6. `productId`
7. `merchantSecretKey`

The Back Office receiver currently expects this source string shape:

```text
totalAmount + currency + responseTimeStamp + PPP_TransactionID + Status + productId + merchantSecretKey
```

The simulator should support SHA-256 first. MD5 can be added only if needed for legacy merchant site configurations.

## Core Payload Model

Payment DMN generation should produce URL-encoded form payloads and support both preview and send.

Required generated fields for the first version:

- `merchant_id`
- `merchant_site_id`
- `totalAmount`
- `currency`
- `responseTimeStamp`
- `PPP_TransactionID`
- `Status`
- `productId`
- `payment_method`
- `ppp_status`
- `message`
- `transactionType`
- `type`
- `clientRequestId`
- `clientUniqueId`
- `userPaymentOptionId`
- `advanceResponseChecksum`

Include enough optional fields to resemble Nuvei examples without pretending every Nuvei field is mandatory.

## Payment Method Defaults

For Pix payment DMNs:

- `payment_method=apmgw_PIX`
- `transactionType=Sale`
- `type=DEPOSIT`
- `currency=BRL` by default
- `totalAmount=30.00` by default
- `productId=` by default
- `message=<Status>`

Status mapping:

- `PENDING` -> `ppp_status=PENDING`
- `APPROVED` -> `ppp_status=OK`
- `DECLINED` -> `ppp_status=FAIL`

Declined defaults:

- `Reason=Rejected by simulator.` unless provided
- `ReasonCode=9999` unless provided
- `ErrCode=9` unless provided

For Boleto payment DMNs:

- `payment_method=apmgw_BOLETO`
- `transactionType=Sale`
- `type=DEPOSIT`
- `currency=BRL` by default
- `totalAmount=30.00` by default
- `productId=` by default
- `message=<Status>`

For card payment DMNs:

- `payment_method=cc_card`
- `transactionType=Sale`
- `type=DEPOSIT`
- `currency=USD` by default
- `totalAmount=30.00` by default
- `productId=` by default
- `message=<Status>`
- sanitized `nameOnCard`, masked `cardNumber`, `expMonth`, `expYear`, and `cardCompany` defaults

The simulator must not accept or emit real PANs or customer cardholder data in its default card flow.

For Local Payments Africa payment DMNs:

- `payment_method=apmgw_Local_payments_Africa`
- `transactionType=Sale`
- `type=DEPOSIT`
- `currency=USD` by default
- `totalAmount=30.00` by default
- `productId=` by default
- `message=<Status>`

## CLI Commands

Initial command set:

```sh
nuvei-dmn-simulator config set-merchant <profile>
nuvei-dmn-simulator config verify <profile>
nuvei-dmn-simulator config set-target <target-name>
nuvei-dmn-simulator preview payment pix --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator send payment pix --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator preview payment boleto --profile <profile> --status PENDING --target <target-name-or-url>
nuvei-dmn-simulator send payment boleto --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator preview payment card --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator send payment card --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator preview payment local-payments-africa --profile <profile> --status PENDING --target <target-name-or-url>
nuvei-dmn-simulator send payment local-payments-africa --profile <profile> --status APPROVED --target <target-name-or-url>
nuvei-dmn-simulator server --profile <profile> --target <target-name-or-url> --port 4545
```

Common non-secret flags:

- `--status`
- `--target`
- `--total-amount`
- `--currency`
- `--user-payment-option-id`
- `--ppp-transaction-id`
- `--transaction-id`
- `--client-unique-id`
- `--client-request-id`
- `--reason`
- `--reason-code`
- `--allow-untrusted-target`
- `--require-correlation-fields`

`--require-correlation-fields` should enforce explicit correlation-focused inputs for deterministic webhook matching tests. In strict mode, require explicit values for:

- `--status`
- `--total-amount`
- `--currency`
- `--user-payment-option-id`

Secret values should not be accepted as normal flags in the primary UX. If environment-variable support is added for CI, it must be documented as automation-only.

## Web Server UI

Initial command:

```sh
nuvei-dmn-simulator server --profile local-demo --target local --port 4545
```

Server defaults:

- Bind to `127.0.0.1` only.
- Do not bind publicly unless explicitly configured.
- Reuse the same core payload builder and sender as the CLI.
- Use HTMX for preview/send interactions.

Initial UI fields:

- Profile selector.
- Target selector or URL field.
- APM selector, initially Pix.
- Status selector: `PENDING`, `APPROVED`, `DECLINED`.
- User Payment Option ID field.
- Total amount and currency fields.
- Optional reason and reason code fields.

Initial UI actions:

- Verify credentials.
- Preview payload.
- Send DMN.

## Raw Payload Mode

Support importing a real Nuvei DMN payload:

```sh
nuvei-dmn-simulator preview payment from-raw --profile local-demo --file payload.txt --status APPROVED
nuvei-dmn-simulator send payment from-raw --profile local-demo --file payload.txt --status DECLINED --target local
```

Behavior:

- Parse URL-encoded payload.
- Apply field overrides.
- Replace merchant fields from the selected verified profile unless explicitly disabled.
- Recompute `advanceResponseChecksum`.
- Preview or send.

This lets developers start from Nuvei documentation examples or real sanitized payloads.

## Extensibility

The internal package structure should separate DMN types and APM-specific defaults.

Suggested package layout:

```text
cmd/nuvei-dmn-simulator
internal/config
internal/nuvei/checksum
internal/nuvei/credentials
internal/nuvei/dmn/payment
internal/nuvei/dmn/payment/apm
internal/sender
internal/targetsafe
internal/server
```

Adding Boleto should require adding an APM default builder, tests, and docs, not changing transport, checksum, config, or UI architecture.

## Tests

Add tests for:

- SHA-256 payment checksum generation.
- Pix status mapping.
- Required field validation.
- URL-encoded body output.
- Raw payload parse and override.
- Target safety allowlist/blocklist/profile behavior.
- Config file permission checks where practical.
- Credential verification request construction.
- CLI preview without sending.
- CLI send with mocked HTTP.
- Web preview/send handlers with mocked HTTP.
