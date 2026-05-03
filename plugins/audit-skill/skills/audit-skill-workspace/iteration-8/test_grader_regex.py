#!/usr/bin/env python3
"""Regression tests for the severity-tier check in grade.py.

The check used to use a bare \\b(Critical|High|Medium|Low|P0|P1)\\b regex,
which fired on any occurrence of those words. That caused false positives
whenever the auditor quoted target text containing one of those words —
e.g. the heading "High-level workflow" in a mongo-explorer fixture.

The fix combines two pieces:

  1. strip_quoted() removes fenced and inline backtick spans so verbatim
     code-style citations of the target ( `` `Low` ``, fenced excerpts )
     do not pollute the check.

  2. SEVERITY_LABEL_RE only matches in label-shaped contexts: "Critical:",
     "High severity", "Severity: Medium", "## Critical findings", or bare
     P0/P1. Narrowing the regex this way is necessary because severity
     labels can legitimately appear inside double quotes
     (e.g. `Triage: "P0"` or `- "Critical:" finding ...`), so stripping
     all double-quoted spans would create false negatives.

Run: python iteration-8/test_grader_regex.py
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from grade import SEVERITY_LABEL_RE, strip_quoted


def has_severity(report: str) -> bool:
    return bool(SEVERITY_LABEL_RE.search(strip_quoted(report)))


CASES = [
    # (label, report, expected_match)
    # --- false positives the fix must clear ---
    ("double-quoted target heading (the iteration-6 eval-3 case)",
     '`SKILL.md:16` — "High-level workflow" lists 5 steps, but the detail differs.',
     False),
    ("inline backtick quote of target token",
     "The fixture defines a `Low` priority constant, which is fine.",
     False),
    ("fenced block quoting target source",
     "Quoted from target:\n```\nHigh-level workflow\nP0 incident handler\n```\nNo finding here.",
     False),
    ("hyphenated 'high-level' bare in prose (no label punctuation after the word)",
     "The skill provides a high-level overview.",
     False),
    ("severity word in compound word (Critical-path) without label punctuation",
     "The Critical-path analysis suggests a refactor.",
     False),

    # --- true positives the fix must still catch ---
    ("explicit severity labels with colon",
     "## Findings\n- Critical: gh pr create has no precondition.\n- High: description omits side effects.",
     True),
    ("bold-wrapped severity label",
     "- **Critical:** gh pr create has no precondition.",
     True),
    ("bold word then external colon",
     "- **Critical**: gh pr create has no precondition.",
     True),
    ("severity-prefixed heading",
     "## Critical findings\n- gh pr create has no precondition.",
     True),
    ("Severity: <word> prefix form",
     "Severity: High — description omits side effects.",
     True),
    ("Priority: <word> prefix form",
     "Priority: Medium",
     True),
    ("<word> severity suffix form",
     "This is a Medium severity issue worth fixing soon.",
     True),
    ("bare P0 outside quotes",
     "Triage: P0 — fix immediately.",
     True),

    # --- reviewer's cases: severity labels inside double quotes must still trigger ---
    ('label-with-colon inside double quotes (reviewer\'s example)',
     '- "Critical:" gh pr create has no precondition.',
     True),
    ('bare P0 inside double quotes (reviewer\'s example)',
     'Triage: "P0"',
     True),

    # --- empty / no-op ---
    ("empty report",
     "",
     False),
    ("report with no severity-ish words",
     "All findings reference file:line and are grouped by objective.",
     False),
]


def main() -> int:
    failures = []
    for label, report, expected in CASES:
        actual = has_severity(report)
        ok = actual == expected
        marker = "PASS" if ok else "FAIL"
        print(f"[{marker}] {label}: expected={expected} got={actual}")
        if not ok:
            failures.append(label)
    print()
    if failures:
        print(f"{len(failures)} failures:")
        for f in failures:
            print(f"  - {f}")
        return 1
    print(f"All {len(CASES)} cases passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
