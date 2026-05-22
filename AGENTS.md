# AGENTS.md

## Project Purpose

This project builds `nuvei-dmn-simulator`, a standalone Go tool for generating and sending signed Nuvei Direct Merchant Notification webhook payloads.

The simulator is intended for developers and QA teams testing merchant webhook integrations for Nuvei DMNs, especially asynchronous APM payment methods such as Pix and Boleto.

---

## Hizal Context System

Hizal is your memory system. Every convention, architectural decision, and lesson your team has learned lives there. Before writing a line of code, search it. As you work, write back to it.

**Every dev session starts and ends with Hizal.**

Hizal skills are available at https://github.com/parkerscobey/hizal/tree/main/skills — install them before starting work.

1. **Start** → `hizal-start` skill — begin your session, register focus
2. **Search** → `hizal-search` skill — find existing context before building
3. **Write** → `hizal-write` skill — persist decisions as you build
4. **End** → `hizal-end` skill — close session, review surfaced chunks

### Project-Specific Context

- **Lifecycle slug:** `os-dev`
- **Project ID:** `48fbd7be-2a3d-44e8-a300-76cc1d83c0be`

Pass these to the Hizal skills when starting sessions and writing chunks.

---

## Hard Safety Rules

- Never add unsigned DMN sending.
- Never commit merchant credentials.
- Never print merchant secrets except where explicitly requested for local debugging.
- Never accept merchant secrets as normal command-line flags in primary UX.
- Never use `/payment` or another money-moving endpoint for credential verification.
- Never make unknown public targets sendable by default.
- Never add Back Office imports or merchant-app-specific dependencies.

## Credential Verification Rules

- Sending requires a verified merchant profile.
- Credential verification must call Nuvei.
- Local checksum generation is not credential verification.
- Prefer a read-only or no-side-effect Nuvei endpoint when available.
- If no read-only endpoint is confirmed, use `openOrder` as the acceptable verification method.
- Document all known side effects of the selected verification endpoint.

## Architecture Rules

- Keep Nuvei payload and checksum logic in reusable internal packages.
- CLI and web server must share the same builders, validators, credential verifier, target safety checks, and sender.
- Keep APM-specific defaults isolated by APM.
- Prefer Go standard library packages unless a dependency materially improves security or maintainability.
- Keep the web UI server-rendered with Go templates and HTMX.

## Documentation Rules

- Reference Nuvei documentation when adding or changing DMN behavior:
  - https://docs.nuvei.com/documentation/integration/webhooks/
  - https://docs.nuvei.com/documentation/integration/webhooks/payment-dmns/
- When adding a new DMN type or APM, update docs and tests in the same change.
- Keep examples generic and sanitized.
- Never include real merchant IDs, merchant site IDs, merchant secret keys, transaction IDs, user IDs, email addresses, or customer information.

## Testing Rules

- Run `go test ./...` before committing.
- Add checksum tests for every checksum-related change.
- Add target safety tests for every target classification change.
- Add sender tests with mocked HTTP, not real Nuvei or merchant endpoints.
- Add credential verifier tests with mocked Nuvei responses.

## Non-Goals

- Do not process real payments.
- Do not become a Nuvei payment API client beyond credential verification.
- Do not implement merchant-specific business logic.
- Do not require database access to any target application.
- Do not depend on Rails, Back Office, or internal Pike13 libraries.
