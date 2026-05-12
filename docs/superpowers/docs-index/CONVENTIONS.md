# Superpowers Documentation Conventions

This file defines how to place, name, status, and index new documents under `docs/superpowers/`. The goal is that a new doc can be classified in 30 seconds without creating another parallel source of truth.

## Directory Boundaries

| Directory | Put documents here when... | Do not put here when... | Source-of-truth role |
|---|---|---|---|
| `audits/` | You are collecting observed evidence, inventories, gaps, parity matrices, code locations, or risk lists without changing behavior. | The document chooses an implementation route or claims final API behavior. | Evidence base for plans and decisions. |
| `contracts/` | You are defining stable API/interface facts, compatibility rules, request/response shape, test contract coverage, or external behavior guarantees. | The document is merely a gap list or implementation checklist. | Final authority for HTTP/interface behavior. |
| `plans/` | You are sequencing implementation phases, ownership, gates, rollback, or task breakdown from approved inputs. | The document is only evidence or a single architectural choice. | Authority for work order and phase gates. |
| `decisions/` | You are recording an explicit choice, rejected alternatives, deprecation/sealing rule, or user-review decision. | The document is a broad execution checklist or raw audit. | Authority for selected direction after approval. |
| `docs-index/` | You are maintaining the documentation map and writing conventions. | The document is about product/code behavior. | Authority for doc governance only. |

If a document appears to belong in multiple directories, choose by its primary verb:

- "Observed / found / counted" -> `audits/`
- "The API must / response is / compatibility is" -> `contracts/`
- "Implement in phases / sequence / rollback" -> `plans/`
- "We choose / deprecate / reject / approve" -> `decisions/`
- "Where docs live / how docs are named" -> `docs-index/`

## Naming Rules

Default filename pattern:

```text
YYYY-MM-DD-<topic>.md
```

Category-specific guidance:

| Directory | Preferred pattern | Examples |
|---|---|---|
| `audits/` | `YYYY-MM-DD-<scope>-<audit-kind>.md` when the topic is broad; `YYYY-MM-DD-<topic>.md` is acceptable for compact inventories. | `2026-05-12-frontend-v2-parity-matrix.md`, `2026-05-12-test-coverage-inventory.md` |
| `contracts/` | Use stable uppercase names only for canonical long-lived contracts; otherwise use dated names. | `API-CONTRACT.md`, `CONTRACT-TEST-PLAN.md` |
| `plans/` | `YYYY-MM-DD-<implementation-scope>.md` | `2026-05-12-frontend-v2-systemic-alignment.md` |
| `decisions/` | `YYYY-MM-DD-<decision-slug>.md` | `2026-05-12-deprecate-traditional-invite.md` |
| `docs-index/` | Stable names for governance docs. | `README.md`, `CONVENTIONS.md` |

Slug rules:

- Use lowercase ASCII words separated by hyphens.
- Prefer product/system nouns over task labels: `frontend-v2-systemic-alignment`, not `t009-plan`.
- Do not reuse a filename for a materially different scope.
- If a document supersedes another, create a new dated file and mark the old one `superseded`; do not rewrite history unless the original was never shared.

## Minimum Header

Every new document should start with a short metadata block before the first section. Existing documents can be left as-is until touched by an explicit cleanup task.

Recommended format:

```markdown
# <Title>

- Date: YYYY-MM-DD
- Task: T00X
- Status: active | draft | superseded | archived
- Owner: <actor or team>
- Scope: <one sentence>
- Supersedes: <path or N/A>
```

Status meanings:

| Status | Meaning |
|---|---|
| `draft` | Written for review; not yet a binding input for implementation. |
| `active` | Current input for implementation, testing, or decision-making. |
| `superseded` | Replaced by a newer named document but kept for traceability. |
| `archived` | Historical record; do not use as an implementation input unless explicitly revived. |

## Minimum Sections

Every new document should include at least these three sections, adapted to the category:

```markdown
## Background

Why this document exists, what inputs it used, and what it intentionally excludes.

## Conclusion

The main finding, contract, plan, or decision. For audits, this can be the most important evidence summary.

## Red Lines

Constraints that downstream work must not violate.
```

Category-specific optional sections:

| Directory | Useful additions |
|---|---|
| `audits/` | Evidence table, file:line references, open questions, risk ranking. |
| `contracts/` | Endpoint matrix, request/response schema, compatibility notes, golden-test mapping. |
| `plans/` | Phases, entry/exit gates, rollback, owners, verification commands. |
| `decisions/` | Options considered, chosen option, rejected options, consequences, seal/rollback rules. |

## Lifecycle Rules

When a document changes state, update two places in the same change:

1. The document's top status line, if present.
2. The row in [`README.md`](./README.md).

Lifecycle transitions:

| From | To | When |
|---|---|---|
| `draft` | `active` | User/foreman accepts it as an implementation or decision input. |
| `active` | `superseded` | A newer document explicitly replaces its scope. |
| `active` | `archived` | The work is complete, abandoned, or no longer relevant, and no newer doc replaces it. |
| `superseded` | `archived` | The superseded doc is no longer useful for traceability beyond history. |

Never delete old planning or decision documents just because they are obsolete. Mark them and preserve the trail unless the user explicitly asks for removal.

## Single Source Of Truth

Use this precedence when documents disagree:

| Fact type | Final authority |
|---|---|
| HTTP method/path/status/envelope/field contract | `contracts/API-CONTRACT.md` |
| Contract-test list and golden-test organization | `contracts/CONTRACT-TEST-PLAN.md` |
| Current frontend-v2 systemic execution order | `plans/2026-05-12-frontend-v2-systemic-alignment.md` |
| Raw evidence for 2026-05-12 frontend/backend gaps | The matching `audits/2026-05-12-*.md` file |
| Traditional invite deprecation direction | `decisions/2026-05-12-deprecate-traditional-invite.md` after it is user-approved |
| Documentation placement and lifecycle | `docs-index/CONVENTIONS.md` |

If a plan conflicts with a contract, fix the plan. If an audit conflicts with a later contract or decision, keep the audit as historical evidence and add a note in the index if needed.

## Task Linking

Every new document should include a task ID near the top:

```markdown
- Task: T00X
```

Rules:

- Use the CCCC task ID that caused the document to be created.
- If multiple tasks contributed, list the primary task first and mention secondary tasks in `Background`.
- The index row must include the same task ID.
- If no task exists, use `N/A` rather than inventing one.

## Index Update Checklist

When adding a new document:

1. Pick the target directory using the boundary table above.
2. Name the file using the category pattern.
3. Add the minimum header with `Date`, `Task`, `Status`, `Owner`, `Scope`, and `Supersedes`.
4. Include `Background`, `Conclusion`, and `Red Lines` sections.
5. Add or update the row in [`README.md`](./README.md) with title/path, one-line summary, landed date, task ID, and status.
6. If it supersedes another document, update the old row and status to `superseded` in the same change.
