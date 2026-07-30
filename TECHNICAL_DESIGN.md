# Technical Design Doc: PipeDebug

**Status:** Draft (section-by-section)  
**Related:** [PRD](PRD.md) · [User Stories](USER_STORIES.md)

---

## 1. Overview

### Purpose

PipeDebug (`pipedebug`) is a CLI-first tool that executes CI/CD jobs locally inside Docker containers that mirror remote runner environments (GitHub Actions first; GitLab CI and CircleCI later). On failure, an optional LLM auto-debug loop packages the failure context, proposes scoped patches for minor code/CI errors, applies them within an allowlist, and re-runs the job until it passes or the failure is escalated to the user.

This document describes how the system will be built: components, data, APIs, tests, and operational constraints. Product intent and user-facing requirements live in the PRD; this TDD is the engineering contract for implementation.

### Scope

**In scope (MVP / P0)**
- CLI entrypoint (`pipedebug run`, flags for job selection, `--no-ai`, `--max-iterations`)
- GitHub Actions workflow parsing sufficient to run jobs/`run` steps locally
- Docker-based executor with repo mount, env/secrets injection, streamed logs
- Failure packaging → LLM scoped patch → apply → re-run loop
- Scope gate so architectural / ambiguous failures escalate instead of auto-editing
- No automatic git commits

**In scope (later / P1–P2)**
- Step-through mode, `pipedebug doctor`, image override
- Patch preview/rollback
- GitLab CI / CircleCI parsers
- Optional `--keep-logs` (retain last N full log files with rotation)

**Out of scope (for this project)**
- Replacing hosted CI as the merge/deploy source of truth
- Full emulation of every marketplace action, OIDC, or proprietary runner feature
- LLM-driven architecture or product redesigns
- Interactive web dashboard / hosted multi-tenant UI (see §2 — demo via CLI + README instead)
- Persisting full step logs by default

### Key requirements from PRD

| ID | Requirement | Design implication |
|----|-------------|--------------------|
| G1 / FR-PAR-* | Local environment parity via Docker | Image resolver + executor are core; parity gaps must fail loudly |
| G2 / FR-CLI-3–4 | Clear step logs and exit codes | Structured step runner with streamed stdout/stderr |
| G3 / FR-CLI-5 | Optional step-through | Executor must support pause/resume in one container session |
| G4 / FR-LLM-* | Auto-debug minor failures and re-run | Loop controller + failure packager + patch applier + LLM client |
| G5 | Human owns architecture | Classifier/scope gate before any write; escalate path required |
| FR-LLM-10 | Never auto-commit | Patch applier edits working tree only |
| FR-CLI-7 / FR-PAR-6 | Image override + secrets | Config layer for env files/flags; never persist secrets into repo |
| FR-UI-* (P2) | Dashboard (PRD later) | **Deferred:** no web UI in this build; history via append-only run journal + strong CLI/README demo |

### Design decisions (locked)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | Go | Strong CLI + Docker ecosystem fit; single static binary is easy to install and demo |
| UI | No dashboard | Core product is terminal feedback; a SPA would dilute scope. Hiring managers rarely run full local stacks—README + short demo recording + clean CLI matters more |
| Persistence | Append-only run journal (Markdown table) | One summary row per run; no full logs by default → tiny disk footprint, human-readable, `git`-friendly to ignore |
| Log UX | Live progress in terminal; full step output buffered (not dumped); on failure show **tail** of failing step; optional `--verbose` / `--keep-logs` | Avoid flooding the terminal or disk; still enough context for humans + LLM |

---

## 2. Tech Stack

### Frontend

**None for MVP.** The product surface is the CLI.

Demo / hiring visibility (without a dashboard):
- High-quality README with install, quickstart, and a short screen recording or GIF of `pipedebug run` + AI loop
- Terminal UX as the “UI” (clear step boundaries, patch summaries, escalation messages)
- Run journal (below) as a lightweight artifact reviewers can open in the repo after a demo

A full dashboard is only useful later if users need to browse many historical runs or compare local vs remote CI at scale. That is not the hiring-critical path for this tool.

### Backend

**Go CLI application** (single module, library packages + `cmd/pipedebug`).

| Layer | Role |
|-------|------|
| `cmd/pipedebug` | Cobra (or stdlib `flag`) CLI: `run`, later `doctor` |
| Config / CI parsers | YAML → normalized job/step model (GitHub Actions first) |
| Executor | Docker Engine API or `docker` CLI wrapper: create container, mount repo, run steps; capture step logs with controlled terminal output (see Log UX) |
| Auto-debug loop | Failure packager, scope gate, patch apply/rollback, re-run orchestration |
| LLM client | HTTP client to a chat/completions API (provider configurable via env) |
| Journal | Append one summary row per completed run |

Pipeline execution always happens **inside Docker**; the Go process is the orchestrator on the host.

### Database

**No database.** Persistence is file-based:

| Artifact | Location (default) | Contents |
|----------|--------------------|----------|
| Run journal | `.pipedebug/history.md` in the target repo (gitignored) | Markdown table; **one new row appended per run** |
| Optional logs | `.pipedebug/logs/` (only with `--keep-logs`, P1) | Full step logs; retain last N runs, delete older |
| In-memory / temp buffer | Per-step during a run | Used for failure packaging to the LLM and for printing a failure tail; discarded when the run ends unless `--keep-logs` |

**Default terminal output (not a raw dump of every line):**
- Step start/end lines and pass/fail status as the job runs (live progress)
- Successful steps: summary only (no full stdout), unless `--verbose`
- Failed step: print the **last N lines** (e.g. 50–100) of that step’s output, plus exit code
- AI loop: short rationale + diff summary, not a second copy of the full log

**Journal row fields (relevant, minimal):**

| Column | Purpose |
|--------|---------|
| `timestamp` | When the run finished (UTC ISO-8601) |
| `workflow` / `job` | What was executed |
| `status` | `passed` / `failed` / `escalated` / `max_iterations` |
| `failed_step` | Step name where it last failed (empty if passed) |
| `iterations` | Auto-debug attempts used |
| `files_changed` | Short list or count of patched paths (e.g. `ci.yml,Makefile` or `2`) |
| `duration` | Wall time (e.g. `47s`) |
| `ai` | `on` / `off` |

No full log bodies in the journal. Example row:

```md
| 2026-07-30T21:04:12Z | ci.yml / build | passed | | 2 | package.json | 47s | on |
```

### Infrastructure

| Piece | Choice |
|-------|--------|
| Runtime host | Developer laptop/CI agent with Docker available |
| Job isolation | Docker containers (runner images mapped from `runs-on` / overrides) |
| Distribution | Go binary via `go install` or GitHub Releases |
| Hosted cloud app | None required |

### Third-party services

| Service | Use |
|---------|-----|
| Docker Engine | Local container runtime for parity execution |
| LLM API (e.g. OpenAI-compatible) | Scoped fix proposals; API key via env (`PIPEDEBUG_LLM_API_KEY` or provider-specific) |
| (Optional later) GitHub/GitLab API | Fetch remote run results for parity comparison—not MVP |

---

## 3. Architecture

*TBD*

### System context diagram

### Component diagram

### Service boundaries

---

## 4. Data Design

*TBD*

### Entities

### Database schema

### Relationships

---

## 5. API Design

*TBD*

### Endpoints

### Request/response formats

### Authentication

---

## 6. Component Design

*TBD*

### Class diagrams (UML)

### Module diagrams

---

## 7. Unit Tests

*TBD*

### Normal case

### Boundary / corner cases

### Exception cases

---

## 8. Non-Functional Requirements

*TBD*

### Performance

### Scalability

### Security

### Reliability

---

## 9. Deployment Architecture

*TBD*

### Environments

### CI/CD

### Cloud resources

---

## 10. Risks & Tradeoffs

*TBD*

### Alternative approaches considered

### Known limitations
