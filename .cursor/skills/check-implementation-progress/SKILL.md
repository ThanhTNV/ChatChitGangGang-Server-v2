---
name: check-implementation-progress
description: Reads plans and checklists under output/, reconciles them with the repository, and writes a dated folder under output/YYYY-MM-DD/ containing a progress summary and a detailed next-step implementation spec for the coding agent. Use when tracking Internal Comm backend progress, planning the next implementation slice, or before starting a new coding task.
---

# Check implementation progress & next-step spec

## When to use

Apply this skill when the user asks to **check progress**, **what to build next**, **sync checklist with code**, or **generate a dated implementation brief** for the Golang backend.

## Inputs (always read)

1. **`output/backend-golang-implementation-plan.md`** — canonical roadmap, sections **A–J** checklists (`- [ ]` / `- [x]`).
2. **Other `output/**/*.md`** — prior dated summaries or notes (if present); use for continuity, not as source of truth over the main plan.
3. **Prior dated dirs** — `output/YYYY-MM-DD/*.md` for last known state; mention deltas if useful.

## Repository verification (cross-check)

Inspect the repo (not only markdown) and record **evidence** in the progress summary:

| Look for | Suggests checklist progress |
|----------|-----------------------------|
| `go.mod`, `go.sum` | A (module) |
| `cmd/`, `internal/` layout | A |
| `Dockerfile` | A, H |
| `Makefile` / `Taskfile` | A |
| `migrations/` or `atlas.hcl` / `sqlc.yaml` | B |
| JWT/JWKS code under `internal/auth` (or similar) | C |
| HTTP routes for channels/messages | D |
| WebSocket + Redis pub/sub | E |
| MinIO / presign handlers | F |
| LiveKit token endpoint | G |
| `/health`, `/ready`, `/docs`, `/openapi.yaml`, metrics | H |
| `internal/apiembed/openapi.yaml` | J (API docs) |

If checkboxes say done but code is missing (or vice versa), **flag the mismatch** explicitly.

## Ordering rule for “next step”

1. Use checklist order **A → J** from `backend-golang-implementation-plan.md`.
2. Pick the **first section** that has any **unchecked** item **or** any mismatch with the repo.
3. The **next step** is the **smallest coherent slice** that completes the next unchecked item(s) in that section (prefer 1–3 days of work, one PR-sized scope).
4. If multiple unchecked items in the same section are tightly coupled (e.g. migrations + first query), **group them** in one spec.

## Output location and date

- Create **`output/YYYY-MM-DD/`** using the **authoritative “Today’s date” from user_info** when available; otherwise use the actual system date for the run.
- Do **not** overwrite an existing dated folder without the user asking; if the folder already exists for that date, append a suffix only if the user requests multiple runs the same day (e.g. `output/2026-04-06-run2/`) — default is **one folder per calendar day**, merge into the same folder’s files if re-run same day **only when the user says to update today’s brief**.

## Files to write (required)

Inside `output/YYYY-MM-DD/`:

| File | Purpose |
|------|---------|
| `README.md` | One-screen index: date, link to the two files below, one-line current focus |
| `progress-summary.md` | Table of sections A–J: each checklist line with status **Done / Open / Mismatch**, plus repo evidence |
| `next-step-spec.md` | Executable brief for the coding agent (see [reference.md](reference.md) for section template) |

## next-step-spec.md content requirements

The spec must be **actionable without guessing architecture**:

- **Goal** — one sentence tied to checklist ids (e.g. “A.1–A.3”).
- **Prerequisites** — env vars, running services (Postgres/Redis), Keycloak URL, etc.
- **Files to create/change** — concrete paths under this repo.
- **Data model / API** — structs, routes, WS message types if relevant; align with the main plan.
- **Step-by-step implementation** — ordered steps the agent can execute in sequence.
- **Tests** — what to add; commands to run (`go test ./...`, `-race`, lint).
- **Definition of done** — bullet list including **updating** `output/backend-golang-implementation-plan.md` checkboxes to `[x]` for completed items (when the user wants the plan kept in sync), and **updating `internal/apiembed/openapi.yaml`** whenever the next slice will touch HTTP routes or JSON contracts (call this out in the spec if the slice is API work).

## Quality bar

- Prefer **concrete** names (`internal/chat`, `GET /v1/channels`) over vague advice.
- Respect **`.cursor/rules/golang-production.mdc`** for this repo’s Go standards.
- Keep `next-step-spec.md` **under ~250 lines**; split extra detail into `next-step-spec-appendix.md` in the same dated folder only if necessary.

## After writing

Tell the user the **dated folder path** and the **single next focus** (e.g. “Section A: scaffold module + config + health”).

## Companion skill

To **implement** the generated `next-step-spec.md` as code, use the **implement-from-next-step-spec** project skill (same `output/YYYY-MM-DD/` folder).
