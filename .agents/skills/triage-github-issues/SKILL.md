---
name: triage-github-issues
description: Review open GitHub issues on the upstream repository and apply `bug`, `feature`, and `waiting-on-response` labels. Use when asked to triage, review, or label open GitHub issues.
---

# Triage and Label GitHub Issues

Review all open issues on `railwayapp/railpack` and apply `bug`, `feature`, and `waiting-on-response` labels based on issue descriptions and comment threads.

## Instructions

1. **Ensure Required Labels Exist**:
   - Verify or create `bug`, `feature`, and `waiting-on-response` on `railwayapp/railpack`:
     ```bash
     gh label create bug --repo railwayapp/railpack --description "Something isn't working" --color d73a4a 2>/dev/null || true
     gh label create feature --repo railwayapp/railpack --description "New feature or request" --color a2eeef 2>/dev/null || true
     gh label create waiting-on-response --repo railwayapp/railpack --description "Waiting on feedback/information from a user" --color fbca04 2>/dev/null || true
     ```

2. **Fetch Open Issues**:
   - Query all open issues with their bodies, authors, existing labels, and full comment threads:
     ```bash
     gh issue list --repo railwayapp/railpack --state open --limit 200 \
       --json number,title,body,author,comments,labels,url
     ```

3. **Classify Each Issue**:
   - **Type (`bug` vs `feature`)**:
     - `bug`: Broken builds or runtime behavior, regressions, crashes, path/platform defects, inaccurate tool/version resolution, or unintended side effects in existing providers.
     - `feature`: Requests for new capabilities, support for additional package managers/languages/frameworks, new `railpack.json` schema options, or CLI/output enhancements.
   - **Response Status (`waiting-on-response`)**:
     - Add `waiting-on-response` when a maintainer or collaborator has asked the issue author/user for clarifying information, their specific use case, a reproduction repository, or confirmation on whether a fix works, and the user has not yet replied with that information.
     - Do **not** add `waiting-on-response` if the user is the one asking a question waiting on a maintainer response, or if a maintainer merely invited a pull request without blocking on missing information.
     - If an issue already has `waiting-on-response` and the user has since replied with the requested details, remove `waiting-on-response`.

4. **Apply Label Updates**:
   - Add labels to each issue:
     ```bash
     gh issue edit <number> --repo railwayapp/railpack --add-label "<label1>,<label2>"
     ```
   - Remove stale `waiting-on-response` labels when answered:
     ```bash
     gh issue edit <number> --repo railwayapp/railpack --remove-label "waiting-on-response"
     ```

5. **Report Summary**:
   - Output a markdown table listing:
     - Issue number (linked) and title
     - Applied type (`bug` / `feature`)
     - `waiting-on-response` status
     - Concise 1-sentence rationale
