# Product Requirements Document: PipeDebug

**Product:** CI/CD Pipeline Debugger (`pipedebug`)  
**Status:** Draft

---

## 1. Overview

PipeDebug is a CLI-first developer tool that runs **GitHub Actions** jobs locally via [nektos/act](https://github.com/nektos/act) (Docker-backed), then—when a local run fails—uses an LLM to propose and apply fixes for **minor** code and CI config errors and re-runs until the job passes or the failure is judged out of scope for automated repair.

---

## 2. Problem Statement

Developers do not have easy access to the same environment their CI provider uses. The common workflow is:

1. Edit CI YAML or a script
2. Commit and push
3. Wait 2–10 minutes for a remote runner
4. Read the failure
5. Repeat

That loop produces:

- **Slow feedback** — minutes per attempt instead of seconds
- **Wasted time** — 5 debug commits can mean 10–50 minutes of waiting
- **Noisy git history** — “fix CI”, “try again”, “please work” commits

Developers want to prototype and debug CI jobs on their machine with environment parity, then commit once when it works.

---

## 3. Goals

| Goal | Description |
|------|-------------|
| G1 | Run GitHub Actions jobs locally (via act + Docker) with parity close to the remote runner |
| G2 | Surface failures with clear step context and logs (same failure mode as remote when parity holds) |
| G3 | Optionally step through jobs: pause, inspect/modify, continue |
| G4 | On failure, use an LLM to fix **minor** errors and automatically re-run until success or escalation |
| G5 | Keep architectural / product decisions in human hands—LLM must not rewrite system design |

### Non-goals

- Replacing hosted GitHub Actions as the source of truth for merges and deployments
- Reimplementing a GitHub Actions runner (parity comes from act)
- Shipping GitLab CI / CircleCI support in this project (architecture may stay expandable; not a product commitment)
- Guaranteeing 100% parity with every GitHub-hosted runner feature (secrets vaults, OIDC, hosted services)—bounded by what act can emulate
- Letting the LLM redesign architecture, choose frameworks, or make product trade-offs

---

## 4. Users

- Solo developers and small teams (roughly 2–10 people)
- Backend / DevOps-leaning engineers who edit GitHub Actions YAML often
- Developers who want a faster local feedback loop before pushing to GitHub Actions

---

## 5. Product Principles

1. **Parity first** — Local runs should fail for the same reasons remote runs would, when environment and config allow.
2. **CLI as primary UX** — Fast path is the terminal; dashboard is secondary.
3. **Human owns architecture** — Automation fixes typos, missing deps, bad paths, and obvious CI misconfig—not design.
4. **Fail safe, not silent** — If the LLM cannot confidently fix a failure, stop and show the developer what to decide.
5. **Minimal surface area** — Prefer a small, solid MVP over a large half-built platform.

---

## 6. Core Concepts

| Concept | Meaning |
|---------|---------|
| **Pipeline config** | GitHub Actions workflow(s) under `.github/workflows/*.yml` |
| **Runner image** | Docker image act uses to approximate the remote job’s OS / toolchain |
| **Local run** | One execution of a selected workflow job via act (+ Docker) |
| **Failure artifact** | Exit code, step name, stdout/stderr, relevant env, and config snippets |
| **Auto-debug loop** | Fail → LLM propose patch → apply → re-run → repeat until pass or stop |

---

## 7. Functional Requirements

### 7.1 CLI (`pipedebug`)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-CLI-1 | `pipedebug run` detects GitHub Actions workflows in the current repo and runs the selected workflow/job locally via act (+ Docker) | P0 |
| FR-CLI-2 | Support selecting workflow, job, and optionally step range (e.g. one job, or from step N) | P0 |
| FR-CLI-3 | Stream step logs to the terminal with clear step boundaries and final pass/fail status | P0 |
| FR-CLI-4 | Exit non-zero when the local run fails (usable in scripts) | P0 |
| FR-CLI-5 | `pipedebug run --step-through` pauses between steps; user can inspect, edit a command, or continue | P1 |
| FR-CLI-6 | `pipedebug doctor` checks Docker availability, act availability, image pull ability, and basic workflow parse health | P1 |
| FR-CLI-7 | Configurable runner image override when auto-detection is wrong | P1 |

### 7.2 Local environment parity

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-PAR-1 | Detect/select GitHub Actions workflows and jobs; run them locally via nektos/act (including `run:`, `uses:`, `if`, etc. as supported by act) | P0 |
| FR-PAR-2 | ~~GitLab CI~~ | **Dropped** — out of project scope; keep `Executor` seam only |
| FR-PAR-3 | ~~CircleCI~~ | **Dropped** — out of project scope |
| FR-PAR-4 | Execute against the mounted working tree (act/Docker workspace mount) | P0 |
| FR-PAR-5 | Support runner image/platform override passed through to act when auto-detection is wrong | P1 |
| FR-PAR-6 | Inject user-provided secrets/env via local env file or flags (never commit secrets) | P0 |
| FR-PAR-7 | Document known parity gaps relative to GitHub-hosted runners / act limitations | P1 |

### 7.3 LLM auto-debug loop

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-LLM-1 | On failed local run, package failure artifact (logs, step, exit code, relevant YAML/file snippets) for the LLM | P0 |
| FR-LLM-2 | LLM proposes a **scoped patch** (file edits and/or CI YAML edits) with a short rationale | P0 |
| FR-LLM-3 | Apply patch only within allowed paths (repo workspace; configurable allowlist) | P0 |
| FR-LLM-4 | Re-run the same job/container flow after each applied patch | P0 |
| FR-LLM-5 | Loop until: success, max iterations reached, or LLM/classifier marks failure as out-of-scope | P0 |
| FR-LLM-6 | Persist a run report: iterations, patches, final status | P1 |
| FR-LLM-7 | `--no-ai` / `--fix` flag to run without auto-debug | P0 |
| FR-LLM-8 | `--max-iterations N` (sensible default, e.g. 3–5) | P0 |

#### In-scope for LLM fixes (examples)

- Typos in shell commands or paths
- Missing packages / wrong package names in install steps
- Incorrect working directory or file path in YAML
- Obvious syntax errors in scripts or CI YAML
- Wrong test command flags that the logs clearly contradict
- Pinning a version when the log shows a clear version mismatch the project already intends

#### Out-of-scope for LLM fixes (escalate to user)

- Choosing architecture, frameworks, or major dependency upgrades
- Changing product behavior / business logic beyond a clear bug indicated by CI logs
- Broad refactors or multi-file redesigns
- Security policy or secrets-management redesign
- Flaky / infrastructure issues with no clear local fix (quota, network policy, private registry auth the user must supply)
- Ambiguous failures where multiple valid fixes exist and choice is a trade-off

When out-of-scope: stop the loop, print the failure summary and why automation stopped, leave the working tree unchanged (or roll back the last speculative patch—see FR-LLM-9).

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-LLM-9 | Prefer atomic apply + rollback: if a patch does not improve the run (or breaks parse), revert that iteration’s patch | P1 |
| FR-LLM-10 | Never auto-commit; user reviews and commits intentionally | P0 |
| FR-LLM-11 | Show a concise diff of each proposed change before apply (or `--yes` for fully automatic apply in local-only mode) | P1 |

### 7.4 Web dashboard (secondary)

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-UI-1 | View saved local run history (status, duration, iteration count) | P2 |
| FR-UI-2 | Inspect logs and applied patches per run | P2 |
| FR-UI-3 | Diff view: local run result vs a fetched remote CI run (same commit/job when available) | P2 |

Dashboard is not required for MVP; CLI + auto-debug loop is the core product.

---

## 8. User Flows

### 8.1 Happy path — local run passes

1. Developer clones/opens repo with CI config
2. Runs `pipedebug run` (optionally selects job)
3. Tool pulls/starts runner image, executes steps
4. Job passes; exits 0; no LLM involvement

### 8.2 Auto-debug until green

1. Developer runs `pipedebug run` (AI enabled by default or via flag)
2. A step fails; failure artifact is captured
3. LLM classifies: minor fix vs escalate
4. If minor: propose patch → show/apply → re-run
5. Repeat until pass or stop condition
6. Developer reviews diffs and commits once

### 8.3 Escalation

1. Failure is architectural / ambiguous / secrets-related
2. Loop stops with explanation and logs
3. Developer decides next change manually, then re-runs

### 8.4 Step-through (optional)

1. `pipedebug run --step-through`
2. After each step (or at failure), pause for inspect / edit / continue
3. Resume remaining steps in the same container session when possible

---

## 9. Technical Design (high level)

```
┌─────────────┐     parse      ┌──────────────────┐
│ CI YAML     │ ─────────────► │ Job / Step model │
└─────────────┘                └────────┬─────────┘
                                        │
                                        ▼
┌─────────────┐   mount repo   ┌──────────────────┐
│ Docker      │ ◄─────────────│ ActExecutor      │
│ (via act)   │   run job      └────────┬─────────┘
└─────────────┘                         │
                                   fail │ pass → done
                                        ▼
                               ┌──────────────────┐
                               │ Failure artifact │
                               └────────┬─────────┘
                                        ▼
                               ┌──────────────────┐
                               │ LLM agent        │
                               │ (scope gate +    │
                               │  patch propose)  │
                               └────────┬─────────┘
                                        │ apply patch
                                        ▼
                               re-enter executor (loop)
```

### Suggested components

| Component | Responsibility |
|-----------|----------------|
| Workflow select / validate | Detect `.github/workflows`; select workflow/job; optional actionlint; build invocation `Job` |
| ActExecutor | Run selected job via nektos/act; stream/capture logs into capped tail → `RunResult` |
| Executor (interface) | Abstraction over job execution so a future non-GHA backend could plug in without rewriting the AI loop |
| Failure packager | Build structured context for the LLM (bounded size) |
| Scope classifier | Minor vs out-of-scope (rules + LLM judgment) |
| Patch applier | Apply unified diffs; validate; rollback on bad apply |
| Loop controller | Iteration limits, stop conditions, run report |
| (Optional) Dashboard API | Persist and serve run history |

### Constraints

- Prefer a single language/runtime for the CLI unless there’s a clear reason to split
- Prefer act for Actions execution; document act/GHA parity gaps rather than silently inventing behavior
- Cap LLM context (tail of logs + relevant file snippets) to keep runs predictable
- All LLM calls require an API key supplied by the developer via env/config

---

## 10. Success Metrics

| Metric | Target |
|--------|--------|
| Environment parity | Local runs reproduce the same failing step as the remote CI job for supported configs |
| Time-to-feedback | Local re-run of a short job completes in seconds–low minutes without a push |
| Auto-debug effectiveness | Minor CI/script failures reach green within the configured max iterations |
| Safe escalation | Architectural or ambiguous failures stop the loop with a clear explanation |
| Git hygiene | Developers can fully debug locally and commit only when ready |

---

## 11. MVP Scope

### Must ship (P0)

- `pipedebug run` for GitHub Actions (one workflow / job) via **nektos/act**
- Streamed logs + capped failure tail for humans and the LLM
- Failure packaging + LLM minor-fix loop with max iterations and `--no-ai`
- No auto-commit; clear terminal report of patches applied

### Should ship (P1)

- `pipedebug doctor` (Docker + act + workflow health)
- Image/platform override + secrets/env injection (via act)
- Patch rollback per failed iteration
- Step-through mode only if cleanly feasible on top of act

### Nice to have (P2)

- Web dashboard + local-vs-remote diff
- Deeper act output parsing (richer failed-step attribution)

**Explicitly not in scope:** GitLab CI, CircleCI.

---

## 12. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Incomplete parity with hosted runners | Depend on act; document act/GHA gaps honestly |
| LLM over-edits architecture | Hard scope rules + escalate path + path allowlist + iteration cap |
| Unsafe patches | Diff preview, rollback, never auto-commit, refuse edits outside allowlist |
| Huge logs blow context | Truncate/summarize; keep failing step + nearby config |
| act integration complexity | Thin `ActExecutor` adapter; pin act version; fake `Executor` in loop tests |

---

## 13. Open Questions

1. ~~Which language/stack for the CLI?~~ → **Go** (locked in TDD)
2. act as Go library vs shell-out to `act` CLI for v1?
3. Is step-through feasible on top of act, or defer?
4. Local-only LLM option (e.g. Ollama) vs cloud API only for v1?
5. Should applied patches live in a git worktree / branch by default for safer review?

---

## 14. Appendix — Example CLI sessions

### Run with AI loop

```bash
pipedebug run --job build --max-iterations 5
# ... step logs ...
# ✗ step "Install deps" failed (exit 1)
# 🤖 Proposing minor fix: package name typo in package.json / install command
# Applied patch (1 file). Re-running…
# ✓ job build passed after 2 iteration(s)
# Review changes with: git diff
```

### Escalation

```bash
pipedebug run --job deploy
# ✗ failure classified as out-of-scope: requires choosing deployment strategy
# Stopped after 0 auto-fixes. See logs above.
```

### No AI

```bash
pipedebug run --no-ai --job test
```
