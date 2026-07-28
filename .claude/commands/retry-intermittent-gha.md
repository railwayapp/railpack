Retry intermittent GitHub Actions failures from the last week. Re-run CI only — do not change code.

## Steps

### 1. List failures

```bash
SINCE=$(date -u -v-7d +%Y-%m-%d 2>/dev/null || date -u -d '7 days ago' +%Y-%m-%d)
gh run list --status failure --created "${SINCE}.." --limit 200 \
  --json databaseId,displayTitle,workflowName,headBranch,event,url,createdAt,attempt,headSha
gh pr list --state open --json headRefName --jq '.[].headRefName'
```

### 2. Filter candidates

| Keep | Drop |
| --- | --- |
| `main` (`push` / `schedule`) | Closed/merged PR branches |
| Open PR head branches | Older than latest failed run for same `(workflowName, headBranch)` |
| | `attempt >= 3` (already flaky enough; report only) |

### 3. Classify (per run)

**Fast path** — failed step names only:

```bash
gh run view "$RUN_ID" --json jobs \
  --jq '[.jobs[]|select(.conclusion=="failure")|{name,steps:[.steps[]|select(.conclusion=="failure")|.name]}]'
```

| Failed steps | Likely | Next |
| --- | --- | --- |
| only `Start BuildKit` / `Stop BuildKit` | intermittent (Docker pull flake) | confirm with logs |
| unit/lint job (`Check and Test`, etc.) | real | confirm with logs |
| mix of setup + test | unknown | logs required |

**Confirm with logs** (always before retry; stop early):

```bash
gh run view "$RUN_ID" --log-failed 2>&1 \
  | rg -i -m 40 'i/o timeout|dial tcp|ECONNRESET|TLS handshake|toomanyrequests|429|502|503|504|rate.?limit|failed to resolve reference|temporary failure|--- FAIL:|snapshot failed|FAIL\s+github\.com|expected .+ got |error=.+timeout'
```

Batch ~3–5 log fetches at a time. Ignore cascade noise: `Stop BuildKit` / `No such container: buildkit` after a failed start is not a separate root cause.

**Decision rule** (check real signals first):

1. Any **real** signal → do **not** retry  
   `--- FAIL:`, `snapshot failed`, compile/lint errors, deterministic `expected … got …`, version/API break (not transport)
2. Else any **intermittent** signal → retry  
   registry/network timeouts, `dial tcp`, 429/5xx from registries, runner lost connection, Docker Hub pull flakes (`moby/buildkit`, ghcr, npm, PyPI)
3. Else → **needs review** (do not retry)

All failed jobs in a run must be intermittent (or pure cascades) to retry the run.

### 4. Rerun

```bash
gh run rerun "$RUN_ID" --failed
```

Continue on individual errors. Do not watch runs unless asked.

### 5. Report

| Category | run | workflow | branch | jobs | signal/reason | url |
| --- | --- | --- | --- | --- | --- | --- |
| retried | | | | | e.g. docker.io i/o timeout | |
| real (skipped) | | | | | e.g. snapshot failed | |
| filtered (skipped) | | | | | closed PR / not latest / attempt≥3 | |
| needs review | | | | | why unclear | |

Counts at the top. One line per run.
