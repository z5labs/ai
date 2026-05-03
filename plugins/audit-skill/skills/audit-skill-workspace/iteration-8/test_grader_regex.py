#!/usr/bin/env python3
"""Regression tests for the severity-tier check in grade.py.

The check used to fire on any occurrence of Critical/High/Medium/Low/P0/P1,
which caused false positives whenever the audit report quoted target text
containing one of those words as a fragment (e.g. `High-level workflow`).
The fix strips backtick-quoted spans before searching, so quotes from the
target no longer count.

Run: python iteration-8/test_grader_regex.py
"""
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from grade import strip_quoted

SEVERITY_RE = re.compile(r"\b(Critical|High|Medium|Low|P0|P1)\b")


def has_severity(report: str) -> bool:
    return bool(SEVERITY_RE.search(strip_quoted(report)))


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
    ("hyphenated 'high-level' bare in prose (lowercase, regex is case-sensitive)",
     "The skill provides a high-level overview.",
     False),
    # --- true positives the fix must still catch ---
    ("explicit severity label",
     "## Findings\n- Critical: gh pr create has no precondition.\n- High: description omits side effects.",
     True),
    ("P0 label outside quotes",
     "Triage: P0 — fix immediately.",
     True),
    ("Medium severity in prose",
     "This is a Medium severity issue worth fixing soon.",
     True),
    ("severity word inside backticks but also bare elsewhere",
     "We avoid the `Low` keyword in code, but this finding is High.",
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
