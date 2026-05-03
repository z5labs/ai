#!/usr/bin/env bash
# Regression check for issue #60: audit-skill must not instruct callers to
# dismiss its own PR reviews. The skill posts with event=COMMENT (pr-mode.md
# step 4), and GitHub's API rejects dismissals on COMMENT-event reviews with
# HTTP 422 "Can not dismiss a commented pull request review". The marker
# line is the dedup mechanism; this script makes sure no future edit
# silently reintroduces the inert dismissal call.

set -euo pipefail

SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PR_MODE="$SKILL_DIR/references/pr-mode.md"
SKILL_MD="$SKILL_DIR/SKILL.md"

fail=0

if grep -nE '/reviews/[^[:space:]]*/dismissals' "$PR_MODE" >/dev/null; then
  echo "FAIL: $PR_MODE references /reviews/.../dismissals" >&2
  grep -nE '/reviews/[^[:space:]]*/dismissals' "$PR_MODE" >&2
  fail=1
fi

if grep -nE 'gh api[^|]*-X[[:space:]]+PUT[^|]*dismissals' "$PR_MODE" >/dev/null; then
  echo "FAIL: $PR_MODE invokes \`gh api -X PUT ... dismissals\`" >&2
  fail=1
fi

if grep -nE 'dismiss it before posting fresh' "$SKILL_MD" >/dev/null; then
  echo "FAIL: $SKILL_MD still narrates the removed dismissal step" >&2
  grep -nE 'dismiss it before posting fresh' "$SKILL_MD" >&2
  fail=1
fi

if [[ $fail -eq 0 ]]; then
  echo "OK: pr-mode dismissal step is absent (issue #60 regression guard)"
fi

exit "$fail"
