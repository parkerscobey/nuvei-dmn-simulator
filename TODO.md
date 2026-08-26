# TODO: nuvei-dmn-simulator

## Phase 1: Repository Skeleton

- [x] Initialize a Go module named `github.com/parkerscobey/nuvei-dmn-simulator`.
- [x] Add `cmd/nuvei-dmn-simulator/main.go`.
- [x] Add package folders under `internal/`.
- [x] Add `README.md` with safety warning and basic usage.
- [x] Add `VISION.md`, `SPEC.md`, `TODO.md`, and `AGENTS.md` to the new repo.
- [x] Add `.gitignore` entries for local config and generated test payloads.
- [x] Add basic `go test ./...` CI.
- [x] Add MIT license.

## Phase 2: Nuvei Payment DMN Core

- [x] Implement payment DMN SHA-256 `advanceResponseChecksum` generation.
- [x] Implement URL-encoded payload serialization.
- [x] Implement payment DMN base payload builder.
- [x] Implement Pix payment DMN defaults.
- [x] Add status mapping for `PENDING`, `APPROVED`, and `DECLINED`.
- [x] Add unit tests using Nuvei documentation examples where possible.

## Phase 3: Config and Secret Handling

- [x] Implement config file loading from user config directory.
- [x] Implement `config set-merchant <profile>` with interactive prompts.
- [x] Ensure merchant secret input is not echoed.
- [x] Implement `config list` with secrets redacted.
- [x] Implement `config set-target <name>`.
- [x] Add config examples with placeholder values only.
- [x] Add tests for config parsing and redaction.

## Phase 4: Credential Verification

- [x] Research and confirm the safest Nuvei credential verification endpoint.
- [x] Implement preferred non-money-moving `/getSessionToken` verification.
- [x] Document `openOrder` as fallback only when `/getSessionToken` is unavailable.
- [x] Ensure `/payment` is never used for credential verification.
- [x] Cache successful verification in memory for the current CLI/server session.
- [x] Add `config verify <profile>` command.
- [x] Add reusable verification gate for future send operations.
- [x] Add tests with mocked Nuvei API responses.

## Phase 5: Target Safety

- [x] Implement default target classification.
- [x] Allow localhost/test/local targets by default.
- [x] Block unknown public targets by default.
- [x] Support trusted target profiles.
- [x] Support explicit `--allow-untrusted-target` escape hatch.
- [x] Add warnings for production-hosted sandbox targets that require confirmation.
- [x] Add tests for target safety behavior.

## Phase 6: CLI Preview and Send

- [x] Implement `preview payment pix`.
- [x] Implement `send payment pix`.
- [x] Support non-secret override flags.
- [x] Print payload preview in readable table and raw URL-encoded form.
- [x] Send as `application/x-www-form-urlencoded`.
- [x] Show HTTP status and response body.
- [x] Add tests with mocked target endpoint.
- [x] Add strict correlation mode via `--require-correlation-fields`.

## Phase 7: Quality Gates and Local Smoke

- [x] Add a local smoke script for preview plus local receiver send flow.
- [x] Add a CI smoke job that runs the local smoke script with sanitized placeholders.
- [x] Keep local smoke fully local with no external network dependency.
- [x] Add a separate optional integration job for credential verification and send using repository secrets.
- [x] Keep `go test ./...` required on every pull request.

## Phase 8: Raw Payload Mode

- [x] Implement URL-encoded raw payload parser.
- [x] Implement status and field overrides.
- [x] Replace merchant fields from selected verified profile by default.
- [x] Recompute checksum after overrides.
- [x] Support preview and send from raw payloads.
- [x] Add tests with sanitized Nuvei sample payloads.

## Phase 9: Web Server UI

- [x] Implement `server` command.
- [x] Bind to `127.0.0.1` by default.
- [x] Add Go templates for profile/target/APM/status form.
- [x] Add HTMX preview endpoint.
- [x] Add HTMX send endpoint.
- [x] Add credential verification action.
- [x] Reuse core builders and sender.
- [x] Add tests for HTTP handlers with mocked sender.
- [x] Capture first usable UI screenshot for README.
- [x] Record short GIF: verify -> preview -> send.

## Phase 10: More APMs

- [x] Add Boleto payment DMN defaults.
- [x] Add generic APM builder options.
- [x] Add documentation for APM-specific fields.
- [x] Add fixture payloads for Pix, Boleto, card, and Local Payments Africa.
- [x] Add tests for each APM default set.

## Phase 11: Open Source Polish

- [ ] Add sanitized README example output for both `preview` and `send`.
- [ ] Add `CONTRIBUTING.md` with local dev and test flow.
- [ ] Add a short walkthrough script for a five-minute local evaluation path.

## Phase 12: Demo Harness

- [ ] Add a tiny example webhook receiver for local end-to-end testing.
- [ ] Add a one-command or short scripted demo flow for verify -> preview -> send.
- [ ] Add sanitized request and response fixtures for docs and demos.

## Phase 13: Distribution

- [ ] Add release build scripts for macOS, Linux, and Windows.
- [ ] Add checksums for released binaries.
- [ ] Add installation instructions for released binaries.
- [ ] Add Homebrew tap plan if useful.

## Phase 14: Web UI Config Management

- [ ] Add web UI form to create/update merchant profiles.
- [ ] Add web UI form to create/update target profiles.
- [ ] Ensure merchant secret input is never echoed or rendered back in responses.
- [ ] Add server-side validation and redacted success/error messaging for config updates.
- [ ] Add handler tests for config-in-UI create/update flows.
