---
name: retry-flakey-jobs
description: Identify and retry flakey failing CI workflow jobs associated with active pull requests.
---

# Retry Flakey Failing Jobs

Find and retry flakey/transient failing GitHub Actions workflow jobs across all active open Pull Requests.

## Instructions

1. **List Open Pull Requests**:
   - Query all open PRs with their branch names and status checks:
     ```bash
     gh pr list --state open --limit 50 --json number,title,headRefName,url,updatedAt,statusCheckRollup
     ```

2. **Inspect Failing Jobs**:
   - For any PR with failing status checks (`FAILURE`, `TIMED_OUT`, `FAILED`), retrieve the failure logs for each failed job:
     ```bash
     gh run view --job <job-id> --log-failed
     ```
     Or search recent runs for the branch:
     ```bash
     gh run list --branch <branch-name> --limit 5 --json databaseId,workflowName,conclusion,status
     ```

3. **Identify Flakey vs Deterministic Failures**:
   - **Flakey / Infrastructure Failures (Retry)**:
     - Network timeouts / connection reset during package or binary downloads (e.g., `mise`, `node`, `ruby`, `pip`, `bun`).
     - HTTP 503, 502, 504, 429 errors from asset mirrors, GitHub releases, or upstream registries.
     - Docker daemon / container socket transient connection drops or runner-level timeouts.
     - Test cases failing due to external network latency or timing race conditions unrelated to code changes.
   - **Deterministic Failures (Do Not Retry)**:
     - Unit test or snapshot test assertion errors (`FAIL: Test...`).
     - Linter / type-check / formatting errors.
     - Build or compilation failures caused by code changes.
     - Missing files or import errors directly related to PR changes.

4. **Retry Flakey Failed Jobs**:
   - For workflow runs with flakey failures, trigger a rerun of only the failed jobs:
     ```bash
     gh run rerun <run-id> --failed
     ```

5. **Report Summary**:
   - Output a clear table detailing:
     - PR number and title
     - Workflow and job name
     - Run ID with link
     - Root cause of the flakey failure
     - Rerun status (`in_progress` / `queued`)
   - List any non-flakey failures that were intentionally skipped with reasons.
