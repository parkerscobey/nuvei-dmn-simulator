# AGENTS.md

## Project Purpose

This project builds `nuvei-dmn-simulator`, a standalone Go tool for generating and sending signed Nuvei Direct Merchant Notification webhook payloads.

The simulator is intended for developers and QA teams testing merchant webhook integrations for Nuvei DMNs, especially asynchronous APM payment methods such as Pix and Boleto.

---

Hizal is your memory system. Every convention, architectural decision, and lesson your team has learned lives there. Before writing a line of code, you search it. As you work, you write back to it.

## Your First Two Steps (every session, no exceptions)

1. **Start a Hizal session**
2. **Search Hizal for existing context on the task**

Only then start writing code.

## 1. Start a Hizal Session

Every dev session starts and ends with Hizal.

```
start_session(lifecycle_slug="os-dev")
```

This returns a `session_id`. Keep it visible — you'll need it for `register_focus` and `end_session`.

Then register what you're working on:

```
register_focus(
  session_id="<session-id>",
  task="<task description>",
  project_id="48fbd7be-2a3d-44e8-a300-76cc1d83c0be"
)
```

### Session Recovery

If you lose your `session_id` (context reset, compaction):

```
get_active_session()
```

- `status="active"` → use the returned `session_id`, call `resume_session` to extend TTL
- `status="none"` → call `start_session` to begin fresh

## 2. Search Hizal for Existing Context

Now that you know what you're building, search Hizal broadly first, then narrow if needed.

`search_context` can search across all accessible scopes by default:

- `AGENT` — your personal memory and prior investigations
- `PROJECT` — Back Office knowledge and conventions
- `ORG` — org-wide standards and principles

Start with 2-3 broad searches using different phrasings:

```
search_context(query="<key concept from the spec>")
search_context(query="<ticket id or feature name>")
search_context(query="<related subsystem or endpoint>")
```

Then narrow when you need a specific layer of context:

```
# Project-specific knowledge and conventions
search_context(
  query="<key concept from the spec>",
  project_id="48fbd7be-2a3d-44e8-a300-76cc1d83c0be",
  scope="PROJECT"
)

# Prior agent memory / investigation notes
search_context(
  query="<key concept from the spec>",
  scope="AGENT",
  chunk_type="MEMORY"
)

# Org-wide principles and standards
search_context(
  query="<key concept from the spec>",
  scope="ORG"
)
```

If you know the exact saved item you're looking for, search by `query_key`.

Examples:

```
search_context(query="<key concept from the spec>", project_id="48fbd7be-2a3d-44e8-a300-76cc1d83c0be")
search_context(query_key="<exact-query-key>", project_id="48fbd7be-2a3d-44e8-a300-76cc1d83c0be")
```

Run 2-3 searches with different phrasings. Read the returned chunks — they contain
architecture decisions, conventions, and prior work that must inform your implementation.

If an `AGENT` memory chunk turns out to be broadly useful for the team, promote it later by
writing it back as `write_knowledge` or `write_convention`.

Don't rediscover what the team already decided.

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

---

## Write to Hizal As You Build

This is not optional. Write chunks as you make decisions — not just at the end.

| What you're writing | Tool | Scope |
|---------------------|------|-------|
| Architecture or design decision | `write_knowledge` | PROJECT |
| Convention this codebase follows | `write_convention` | PROJECT (always_inject) |
| Something personal you learned | `write_memory` | AGENT |

**Do not use `write_context`** — it's deprecated. Use the purpose-built tools above.

Write one chunk per meaningful decision. Don't batch everything into one chunk at the end.

## End Your Session

```
end_session(session_id="<session-id>")
```

Review the returned MEMORY chunks. For each one, decide:
- **Keep** — useful personal observation, leave as AGENT memory
- **Promote** — valuable for the team, call `write_knowledge` with the content
- **Discard** — noise, ignore it

This is how knowledge compounds across agents and sessions.
