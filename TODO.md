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

- [ ] Implement default target classification.
- [ ] Allow localhost/test/local targets by default.
- [ ] Block unknown public targets by default.
- [ ] Support trusted target profiles.
- [ ] Support explicit `--allow-untrusted-target` escape hatch.
- [ ] Add warnings for production-hosted sandbox targets that require confirmation.
- [ ] Add tests for target safety behavior.

## Phase 6: CLI Preview and Send

- [ ] Implement `preview payment pix`.
- [ ] Implement `send payment pix`.
- [ ] Support non-secret override flags.
- [ ] Print payload preview in readable table and raw URL-encoded form.
- [ ] Send as `application/x-www-form-urlencoded`.
- [ ] Show HTTP status and response body.
- [ ] Add tests with mocked target endpoint.

## Phase 7: Raw Payload Mode

- [ ] Implement URL-encoded raw payload parser.
- [ ] Implement status and field overrides.
- [ ] Replace merchant fields from selected verified profile by default.
- [ ] Recompute checksum after overrides.
- [ ] Support preview and send from raw payloads.
- [ ] Add tests with sanitized Nuvei sample payloads.

## Phase 8: Web Server UI

- [ ] Implement `server` command.
- [ ] Bind to `127.0.0.1` by default.
- [ ] Add Go templates for profile/target/APM/status form.
- [ ] Add HTMX preview endpoint.
- [ ] Add HTMX send endpoint.
- [ ] Add credential verification action.
- [ ] Reuse core builders and sender.
- [ ] Add tests for HTTP handlers with mocked sender.

## Phase 9: More APMs

- [ ] Add Boleto payment DMN defaults.
- [ ] Add generic APM builder options.
- [ ] Add documentation for APM-specific fields.
- [ ] Add fixture payloads for Pix and Boleto.
- [ ] Add tests for each APM default set.

## Phase 10: Distribution

- [ ] Add release build scripts for macOS, Linux, and Windows.
- [ ] Add checksums for released binaries.
- [ ] Add installation instructions.
- [ ] Add Homebrew tap plan if useful.
- [ ] Add license.
