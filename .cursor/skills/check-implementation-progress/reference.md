# Templates for dated output files

Use these structures in `output/YYYY-MM-DD/`. Replace placeholders.

---

## README.md

```markdown
# Implementation checkpoint — YYYY-MM-DD

- [Progress summary](./progress-summary.md)
- [Next step spec](./next-step-spec.md)

**Current focus:** <one line>
```

---

## progress-summary.md

```markdown
# Progress summary — YYYY-MM-DD

## Source of truth

- Plan: `output/backend-golang-implementation-plan.md`

## Checklist status

| Section | Item | Status | Evidence |
|---------|------|--------|----------|
| A | ... | Done / Open / Mismatch | e.g. `go.mod` present |

## Mismatches

- <checkbox says X but repo shows Y, or none>

## Recently completed (since last dated folder)

- <bullets or "none">
```

---

## next-step-spec.md

```markdown
# Next step implementation spec — YYYY-MM-DD

## Metadata

- **Targets checklist:** <e.g. A. Foundation — items 1–3>
- **Estimated scope:** <e.g. 1–2 days>
- **Out of scope:** <explicit>

## Goal

<Single sentence>

## Prerequisites

- [ ] <tool / service / secret>

## Design notes

- <2–5 bullets tying to architecture: single binary, JWT, etc.>

## Files to add or modify

| Path | Action |
|------|--------|
| `cmd/server/main.go` | Create |
| ... | Modify |

## API / types (if applicable)

- REST: ...
- WebSocket: ...
- OpenAPI: extend `internal/apiembed/openapi.yaml` for any new/changed routes or schemas (served at `/openapi.yaml`, UI at `/docs`).
- Env vars: ...

## Implementation steps

1. ...
2. ...
3. ...

## Tests

- Unit: ...
- Integration: ...
- Commands: `go test ./...`, `go test -race ./...`, `golangci-lint run`

## Definition of done

- [ ] <behavior works locally>
- [ ] Tests pass
- [ ] If HTTP API changed: `internal/apiembed/openapi.yaml` updated (`/docs` and `/openapi.yaml` stay accurate)
- [ ] Checkboxes updated in `output/backend-golang-implementation-plan.md` for: <list ids>

## Risks / follow-ups

- <optional>
```
