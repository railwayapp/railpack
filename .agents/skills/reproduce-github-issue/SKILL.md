---
name: reproduce-github-issue
description: Investigate a bug from a single GitHub issue URL, attempt a local reproduction with a minimal example project, and draft an evidence-backed issue response. Use when asked to reproduce or verify a reported issue.
---

# Reproduce a GitHub issue

Input: one GitHub issue URL. Infer the repository and issue number from it. The default task is to investigate locally, leave a reviewable reproducer in the repository's example or fixture structure, and report the result. Fixing the product, opening a PR, and posting comments require scope or authorization from the user.

## Read the complete report

Fetch the issue body and **all comments**, following pagination. Read them before choosing reproduction cases. With GitHub CLI, use the issue endpoint and the paginated comments endpoint:

```sh
gh api repos/OWNER/REPO/issues/NUMBER
gh api --paginate repos/OWNER/REPO/issues/NUMBER/comments
```

An authenticated GitHub connector is also suitable if it returns the complete thread. Follow linked repro projects, logs, or attachments when they materially define the failure.

Extract the failure stage, observed error, reported releases, dependency/tool versions and ranges, platform, configuration, and attempted workarounds. Later comments can correct the original report: a release upgrade plus a PATH workaround is not evidence that the upgrade alone fixed the bug. Treat suggested causes as hypotheses.

## Inspect the checkout and test path

Identify the matching local repository from the current workspace and remotes. Read its applicable instructions and task definitions; inspect the branch, HEAD, and dirty/staged/untracked state. Preserve existing work. If the repository is unavailable, report that limitation and request only the missing checkout information needed to proceed.

Trace the relevant implementation and nearest existing examples before editing. Establish whether existing coverage actually executes the reported failure: configuration, version selection, fixture workarounds, or a build-only shortcut can mask it. A generated plan or snapshot cannot establish that a command received the intended environment at runtime.

Use repository-supported development and test commands. Inspect task bodies before running them, especially snapshot updaters that stage files or clean snapshots for tests excluded by a filter. Avoid unrelated snapshot changes and repository mutations prohibited by local instructions.

## Create and exercise the reproducer

Add the smallest example that retains the reported trigger. Match version declarations, dependency type, configuration, and failing operation before simplifying. Keep the initial attempt free of the reporter's workaround; use a workaround only as an explicitly identified comparison.

Cover material clarifications from the comments, such as a second affected major version. Prefer a few relevant cases over a broad compatibility matrix. Test current code and, when feasible and useful, the reported release in an isolated location without disturbing the active checkout. Distinguish them in the findings; success on current HEAD says nothing conclusive about an older release.

Follow the repository's integration-test format and exercise real behavior. For a native dependency, load and use the native module. For a server, use its HTTP check convention. Ensure the build actually reaches the operation named in the report, then verify the resulting application where appropriate.

Record the **effective tool version at the failing stage**, not just the requested range or provisioned package. Package managers can select another version during installation or startup. If that happens, explain which versions executed which operations. Do not silently relax an assertion in a way that stops exercising the reported trigger.

Generate lockfiles with their package managers. If lockfile validation or another setup error blocks the experiment, distinguish it from the reported failure. Explain any fixture adjustment or lockfile omission and the resulting difference from the reporter's setup; do not present a modified setup as an exact reproduction.

Run focused integration checks and the repository's required validation. Keep temporary plans and logs in its designated temporary directory. Review generated snapshots and the final diff, including untracked files. Leave the requested example and necessary test artifacts, with comments explaining non-obvious test conditions. Do not expand a reproduction task into an unsolicited product fix.

## Report the result precisely

State one of: reproduced, could not reproduce in the tested setup, or blocked before reaching the reported operation. Include:

- The example path and rerun command.
- The checkout SHA or release, platform, and effective versions tested.
- Whether the reported operation ran, which checks passed, and any different failures encountered.
- Material differences from the report and remaining limits, including untested releases or unavailable project configuration.

Use “could not reproduce on this checkout” rather than claiming the issue is fixed. A tooling update, including a mise update, is only a possible explanation unless a comparison establishes it as the cause. Do not attribute success to an update merely because current code passes.

When reproduction fails, include a concise **draft** issue comment summarizing the verified scope and inviting the reporter to open a PR with a minimal reproducing example project. Adapt the destination to the repository, such as `examples/` when that is its convention. Mention a possible recent tooling fix only when the context supports that hypothesis, and label it as uncertain.

Show the draft before posting and wait for explicit user approval of it. Once approved, post that text without asking again, verify success, and return the comment URL. If the thread changed materially since drafting, reconcile the new information before publishing. Do not close the issue or open a PR as part of posting the comment.
