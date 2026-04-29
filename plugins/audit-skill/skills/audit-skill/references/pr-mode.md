# PR-mode mechanics

This file specifies the exact `gh` invocations the skill uses when a PR is open. The SKILL.md workflow points here when output mode is PR. Read this only after the four objective checks have produced a list of findings.

## Inputs you derive once, up front

| Variable | How to get it |
|---|---|
| `PR_NUMBER` | `gh pr view --json number -q .number` |
| `OWNER`, `REPO` | `gh repo view --json owner,name -q '.owner.login + "/" + .name'`, then split on `/` |
| `HEAD_SHA` | `gh pr view --json headRefOid -q .headRefOid` |
| `MARKER` | the literal string `<!-- audit-skill: $HEAD_SHA -->` |

`HEAD_SHA` must come from the PR's tip commit (`headRefOid`), not from `git rev-parse HEAD` — those diverge if the working tree has uncommitted changes or if the user is reviewing a different branch.

## Step 1: deduplicate against prior runs (idempotency in PR mode)

Re-running the audit on the same head SHA must be a no-op. Without this guard, every re-run posts a duplicate review.

1. List existing reviews: `gh api repos/$OWNER/$REPO/pulls/$PR_NUMBER/reviews --paginate -q '.[] | {id, body, commit_id, html_url}'`.
2. Filter to reviews whose body starts with `<!-- audit-skill: ` — those are prior audit-skill runs.
3. Decide:
   - **Marker matches `HEAD_SHA`**: an audit was already posted for this exact commit. Skip the rest of PR mode entirely. Tell the user the existing review URL and stop.
   - **Marker is for an older SHA**: dismiss it before posting fresh. `gh api -X PUT repos/$OWNER/$REPO/pulls/$PR_NUMBER/reviews/$ID/dismissals -f message="superseded by audit on $HEAD_SHA"`. Continue to step 2.
   - **No marker found**: continue to step 2.

Dismissed reviews stay in the PR's history (collapsed), so the user can still see what changed between audits — that's the point of dismissing rather than deleting.

## Step 2: build the modified-line index

GitHub's review-comments API only accepts inline comments anchored to lines that appear in the PR diff. Findings on unmodified lines have to fall through to the summary review body, where their `file:line` references stay textual.

```bash
gh pr diff --patch > /tmp/audit-skill.diff
```

Parse the diff to build a `path → set-of-RIGHT-side-line-numbers` map:

- File boundary: lines starting with `+++ b/`. The path that follows is `$PATH` for the next set of hunks.
- Hunk header: `@@ -A,B +C,D @@` or `@@ -A +C @@` (when `B` or `D` is 1, the comma+number is omitted). For each hunk, the lines `[C, C + D)` on the RIGHT side are reachable by an inline comment. Lines that appear only as `-` deletions are LEFT-side and not eligible.
- A line in the new file is eligible iff its line number falls in some `[C, C+D)` range for that path AND the corresponding patch line begins with `+` or ` ` (space). Pure-context lines (` `) are eligible too — GitHub anchors comments on them.

Skip path entries for binary files (`Binary files ... differ`).

## Step 3: post inline comments for anchored findings

For each finding whose `path:line` is in the modified-line index:

```bash
gh api -X POST repos/$OWNER/$REPO/pulls/$PR_NUMBER/comments \
  -f commit_id="$HEAD_SHA" \
  -f path="$PATH" \
  -F line=$LINE \
  -f side=RIGHT \
  -f body="$BODY"
```

`-F` (capital) is for numeric/bool values that must be sent unquoted; `-f` (lowercase) is for strings. `LINE` is numeric, `commit_id` / `path` / `side` / `body` are strings.

For multi-line findings (a finding that spans `start_line` to `line`), add:

```bash
  -F start_line=$START_LINE -f start_side=RIGHT
```

The `body` follows the inline-comment template in `report-template.md`: a bold `**<Objective>** —` lead, then a one-line description and an optional one-sentence suggestion.

If a POST returns 422, the most common cause is `LINE` not being on the diff. Double-check the modified-line index; if it's correct, the line was likely deleted in a force-push between when you grabbed `HEAD_SHA` and when you posted. Re-fetch `headRefOid` and retry; if the SHA changed, restart the deduplication step.

## Step 4: post the summary review

A single review with the full audit summary. This is the only place the marker line appears.

The body:

```
<!-- audit-skill: $HEAD_SHA -->
audit-skill: <total> findings across <N> objectives

idempotency: <n>, reproducibility: <n>, context-management: <n>, strict-definitions: <n>

<for each finding NOT posted inline (path:line not in the modified-line index):>
- `<file>:<line>` — **<objective>** — <description>

<if total == 0:>
audit clean — <total-checks-run> checks passed across all four objectives.
```

Post via:

```bash
gh api -X POST repos/$OWNER/$REPO/pulls/$PR_NUMBER/reviews \
  -f commit_id="$HEAD_SHA" \
  -f event=COMMENT \
  -f body=@/tmp/audit-skill-summary.md
```

`event=COMMENT` posts a non-blocking review (vs `APPROVE` or `REQUEST_CHANGES` which would gate the merge). The audit doesn't have an opinion on whether to merge — it surfaces findings; the human decides.

## Step 5: report back to the user

Print a one-block summary: the PR number, the number of inline comments posted, the number of findings rolled up into the summary, and the URL of the summary review (`html_url` from the POST response).

If step 1 short-circuited (audit already posted for this SHA), the summary message is just: "audit already posted at <existing-review-url>; re-running on the same head commit is a no-op".
