# User Stories: PipeDebug

Stories are written for the developer end user. Priority aligns with the [PRD](PRD.md) (P0 = MVP, P1 = should ship, P2 = nice to have).

---

## Personas

| Persona | Description |
|---------|-------------|
| **Developer** | Writes application code and CI config; wants fast local feedback before pushing |
| **DevOps engineer** | Owns pipeline reliability; tunes YAML, images, and secrets for the team |

---

## Epic A — Local CI execution

### US-A1 — Run a GitHub Actions job locally
**Priority:** P0  
**Related:** FR-CLI-1, FR-CLI-3, FR-CLI-4, FR-PAR-1, FR-PAR-4

> As a developer,  
> I want to run a GitHub Actions job from my repo locally with `pipedebug run`,  
> so that I can see pass/fail without committing and waiting for a remote runner.

**Acceptance criteria**
- [ ] From a repo with `.github/workflows`, `pipedebug run` executes the selected job via **nektos/act** (+ Docker)
- [ ] Job logs stream to the terminal with clear progress / failure output
- [ ] The process exits `0` on success and non-zero on failure
- [ ] The working tree used by the job is the local repo (mounted), not a remote checkout

---

### US-A2 — Select workflow and job
**Priority:** P0  
**Related:** FR-CLI-2

> As a developer,  
> I want to choose which workflow and job to run,  
> so that I can focus on the failing part of the pipeline instead of the whole workflow.

**Acceptance criteria**
- [ ] User can specify workflow and/or job via CLI flags
- [ ] If multiple workflows/jobs exist and none is selected, the tool prompts or errors with a clear message
- [ ] Step-range selection is optional/P1 and only if cleanly supported on top of act

---

### US-A3 — Inject secrets and environment variables
**Priority:** P0  
**Related:** FR-PAR-6

> As a developer,  
> I want to pass secrets and env vars into the local run from a local file or flags,  
> so that jobs that need credentials can run without putting secrets in git.

**Acceptance criteria**
- [ ] Env/secrets can be supplied via a local env file and/or CLI flags
- [ ] Values are available to steps inside the container
- [ ] The tool never writes those secrets into the repo or commit history

---

### US-A4 — Override the runner image
**Priority:** P1  
**Related:** FR-CLI-7, FR-PAR-5

> As a DevOps engineer,  
> I want to override the Docker image / platform used for a local act run,  
> so that I can match a custom or mis-detected runner environment.

**Acceptance criteria**
- [ ] CLI accepts an image/platform override
- [ ] When set, the override is passed through to act
- [ ] Override is reported in the run output so the user knows which image was used

---

### US-A5 — Check local setup health
**Priority:** P1  
**Related:** FR-CLI-6

> As a developer,  
> I want a `pipedebug doctor` command,  
> so that I can quickly verify Docker, act, image access, and workflow selection before debugging a failure.

**Acceptance criteria**
- [ ] Checks Docker availability
- [ ] Checks act availability / basic invocation health
- [ ] Checks ability to pull/use a runner image (when practical)
- [ ] Detects workflows and reports basic health / selection errors
- [ ] Prints clear pass/fail results for each check

---

### US-A6 — Step through a job interactively
**Priority:** P1  
**Related:** FR-CLI-5

> As a developer,  
> I want to pause between steps, inspect the container, optionally change a command, and continue,  
> so that I can prototype fixes interactively without restarting from scratch each time.

**Acceptance criteria**
- [ ] `--step-through` pauses between steps (and/or on failure)
- [ ] User can inspect, edit the next command, continue, or abort
- [ ] Remaining steps resume in the same container session when possible

---

### US-A7 — Run GitLab CI jobs locally
**Priority:** Dropped (out of scope)  
**Related:** FR-PAR-2

> Dropped for this project. Keep the `Executor` interface so a future GitLab backend could plug in without rewriting the AI loop—do not implement or advertise GitLab support now.

### US-A8 — Run CircleCI jobs locally
**Priority:** Dropped (out of scope)  
**Related:** FR-PAR-3

> Dropped for this project.

---

## Epic B — LLM auto-debug loop

### US-B1 — Auto-fix minor failures and re-run
**Priority:** P0  
**Related:** FR-LLM-1, FR-LLM-2, FR-LLM-3, FR-LLM-4, FR-LLM-5, FR-LLM-8

> As a developer,  
> I want failed local runs to be analyzed by an LLM that applies minor fixes and re-runs the job,  
> so that obvious CI/script errors are resolved without a push-and-wait cycle.

**Acceptance criteria**
- [ ] On failure, the tool packages logs, step name, exit code, and relevant file/YAML snippets
- [ ] LLM proposes a scoped patch with a short rationale
- [ ] Patch is applied only within the workspace allowlist
- [ ] The same job is re-run after each applied patch
- [ ] Loop stops on success, max iterations, or out-of-scope classification
- [ ] `--max-iterations N` is supported with a sensible default

---

### US-B2 — Keep architectural decisions with the developer
**Priority:** P0  
**Related:** FR-LLM-5, out-of-scope rules in PRD §7.3

> As a developer,  
> I want the auto-debug loop to stop when a failure needs an architectural or product decision,  
> so that the tool does not silently change design choices I should make myself.

**Acceptance criteria**
- [ ] Failures involving architecture, major dependency/framework changes, broad refactors, or ambiguous trade-offs are classified as out-of-scope
- [ ] Loop stops and prints why automation will not fix the issue
- [ ] Working tree is left unchanged, or the last speculative patch is rolled back

---

### US-B3 — Run without AI
**Priority:** P0  
**Related:** FR-LLM-7

> As a developer,  
> I want to disable auto-debug for a run (`--no-ai` / `--fix`),  
> so that I can inspect a pure local CI failure without any file changes.

**Acceptance criteria**
- [ ] With AI disabled, a failed run does not call the LLM or modify files
- [ ] Logs and exit status still reflect the local job result

---

### US-B4 — Review changes before committing
**Priority:** P0  
**Related:** FR-LLM-10

> As a developer,  
> I want PipeDebug never to create git commits automatically,  
> so that I remain in control of what lands in history.

**Acceptance criteria**
- [ ] No `git commit` (or equivalent) is performed by the tool
- [ ] After a successful auto-debug loop, the user is reminded to review with `git diff` (or similar)

---

### US-B5 — Preview or auto-apply patches
**Priority:** P1  
**Related:** FR-LLM-11

> As a developer,  
> I want to see a concise diff of each proposed fix before it is applied, with an option to auto-apply,  
> so that I can supervise changes or run unattended locally when I choose.

**Acceptance criteria**
- [ ] By default, proposed diffs are shown before apply
- [ ] `--yes` (or equivalent) applies patches without interactive confirmation
- [ ] User can reject a patch and stop the loop

---

### US-B6 — Roll back a bad auto-fix
**Priority:** P1  
**Related:** FR-LLM-9

> As a developer,  
> I want a patch that does not improve the run (or breaks config parsing) to be reverted,  
> so that a bad suggestion does not leave my workspace worse than before.

**Acceptance criteria**
- [ ] Each iteration’s patch is applied atomically enough to revert
- [ ] If re-run is worse or config becomes invalid, that iteration’s patch is rolled back
- [ ] Rollback is reported in the terminal output

---

### US-B7 — Persist a run report
**Priority:** P1  
**Related:** FR-LLM-6

> As a developer,  
> I want a saved report of iterations, patches, and final status,  
> so that I can review what the auto-debug loop changed after the fact.

**Acceptance criteria**
- [ ] A run report is written after an AI-enabled run (path documented)
- [ ] Report includes iteration count, patch summaries, and final pass/fail

---

## Epic C — Visibility & comparison

### US-C1 — Browse local run history
**Priority:** P2  
**Related:** FR-UI-1, FR-UI-2

> As a developer,  
> I want a dashboard of past local runs with logs and applied patches,  
> so that I can revisit how a pipeline failure was diagnosed.

**Acceptance criteria**
- [ ] User can list past runs (status, duration, iteration count)
- [ ] User can open a run to view logs and patches

---

### US-C2 — Diff local vs remote CI results
**Priority:** P2  
**Related:** FR-UI-3

> As a DevOps engineer,  
> I want to compare a local run to a remote CI run for the same commit/job,  
> so that I can spot environment parity gaps.

**Acceptance criteria**
- [ ] User can select a local run and a remote run (same commit/job when available)
- [ ] Diff view highlights result/log differences relevant to parity debugging

---

## Story map (priority)

| Priority | Stories |
|----------|---------|
| P0 | US-A1, US-A2, US-A3, US-B1, US-B2, US-B3, US-B4 |
| P1 | US-A4, US-A5, US-A6, US-B5, US-B6, US-B7 |
| P2 | US-C1, US-C2 |
| Dropped | US-A7 (GitLab), US-A8 (CircleCI) |
