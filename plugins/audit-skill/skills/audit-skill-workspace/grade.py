"""Grade audit-skill eval runs against assertions."""
import json
import os
import re
import sys
from pathlib import Path

WS = Path(os.environ.get("AUDIT_SKILL_WS", "/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1"))
FIXTURE_BAD = Path("/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill/evals/fixtures/bad-skill/SKILL.md")
FIXTURE_GOOD = Path("/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill/evals/fixtures/good-skill/SKILL.md")


def read_first_md(d: Path) -> tuple[Path, str]:
    """Return the first .md file in d (recursively) that's not a SKILL.md fixture."""
    candidates = sorted(d.rglob("*.md"))
    for c in candidates:
        # Skip the planted skill fixtures
        if "/skills/" in str(c) and c.name == "SKILL.md":
            continue
        if "/.claude/skills/" in str(c) and c.name == "SKILL.md":
            continue
        return c, c.read_text()
    return None, ""


def file_unchanged(planted: Path, fixture: Path) -> bool:
    if not planted.exists() or not fixture.exists():
        return False
    return planted.read_bytes() == fixture.read_bytes()


def severity_tiers_present(text: str) -> bool:
    """Check for severity-tier labels."""
    pattern = re.compile(r"\b(critical|high[-\s]severity|medium[-\s]severity|low[-\s]severity|severity:\s*(critical|high|medium|low)|P[0-3]|priority:\s*(critical|high|medium|low))\b", re.I)
    return bool(pattern.search(text))


def has_section(text: str, names: list[str]) -> bool:
    """Check if any of the names appears as a markdown heading or list label."""
    for name in names:
        n = re.escape(name)
        # match as heading (any level)
        if re.search(r"(^|\n)#{1,6}\s+[^\n]*" + n, text, re.I):
            return True
        if re.search(r"\*\*\s*" + n + r"\s*\*\*", text, re.I):
            return True
        if re.search(r"^\s*[-*]\s*" + n, text, re.I | re.M):
            return True
    return False


def count_file_line_cites(text: str) -> tuple[int, int]:
    """Count bullets *inside Findings sub-sections* and how many cite file:line."""
    findings = []
    in_findings = False
    for line in text.splitlines():
        h2 = re.match(r"^##\s+(.+)$", line)
        if h2:
            heading = h2.group(1).strip().lower()
            in_findings = "finding" in heading
            continue
        if not in_findings:
            continue
        s = line.strip()
        if s.startswith("-") or s.startswith("*"):
            # Skip "no findings" placeholders
            if re.search(r"^[-*]\s*no findings", s, re.I):
                continue
            findings.append(s)
    cited = sum(1 for f in findings if re.search(r"`?[\w./-]+\.\w+:\d+", f) or re.search(r"line\s+\d+", f, re.I))
    return cited, len(findings)


def grade_eval0(report_path: Path, planted_skill: Path) -> list[dict]:
    """Grade eval-0 (bad-skill, expects many findings)."""
    text = report_path.read_text() if report_path else ""
    size = len(text.encode())
    cited, total = count_file_line_cites(text)
    return [
        {"text": "a markdown report file is written somewhere the user can read (the assistant's final message names the path)",
         "passed": report_path is not None and report_path.exists(),
         "evidence": f"report at {report_path}" if report_path else "no report found"},
        {"text": "the report file is non-empty (at least 800 bytes)",
         "passed": size >= 800,
         "evidence": f"{size} bytes"},
        {"text": "./skills/bad-skill/SKILL.md is byte-identical to the input fixture (read-only audit)",
         "passed": file_unchanged(planted_skill, FIXTURE_BAD),
         "evidence": f"unchanged={file_unchanged(planted_skill, FIXTURE_BAD)}"},
        {"text": "the report has an Idempotency section (heading or labeled findings list)",
         "passed": has_section(text, ["Idempotency", "idempotency"]),
         "evidence": "found Idempotency heading/label" if has_section(text, ["Idempotency"]) else "no Idempotency section"},
        {"text": "the report has a Reproducibility section",
         "passed": has_section(text, ["Reproducibility"]),
         "evidence": ""},
        {"text": "the report has a Context management section",
         "passed": has_section(text, ["Context management", "Context-management", "Context"]),
         "evidence": ""},
        {"text": "the report has a Strict definitions section",
         "passed": has_section(text, ["Strict definitions", "Strict-definitions", "Strict"]),
         "evidence": ""},
        {"text": "a finding mentions that the description has no 'when to skip' / negative case (under strict-definitions)",
         "passed": bool(re.search(r"when[\s-]to[\s-]skip|negative case|over[\s-]trigger|when not to|when NOT to", text, re.I)),
         "evidence": ""},
        {"text": "at least one finding cites a specific vague phrase from the bad skill",
         "passed": any(re.search(rf"\b{p}\b", text, re.I) for p in ["as needed", "appropriately", "reasonably", "use your judgment", "use judgment", "as appropriate", "if relevant", "as relevant"]),
         "evidence": ""},
        {"text": "a finding mentions that `date` is used without being declared as an input (under reproducibility)",
         "passed": bool(re.search(r"\bdate\b.*\b(input|declared|undeclared|implicit)", text, re.I) or re.search(r"(input|declared|undeclared|implicit).*\bdate\b", text, re.I)),
         "evidence": ""},
        {"text": "a finding flags `gh pr create` for missing a precondition or idempotency declaration",
         "passed": bool(re.search(r"gh pr create", text)) and bool(re.search(r"precondition|idempoten|state[\s-]check|exists|already", text, re.I)),
         "evidence": ""},
        {"text": "a finding flags the long inline output template as content that should move to references/",
         "passed": bool(re.search(r"inline (template|block|code|content)|references/|move (it )?to references|assets/", text, re.I)),
         "evidence": ""},
        {"text": "a finding notes the description doesn't mention the workflow's side effects (tag push, PR creation)",
         "passed": any(
             bool(re.search(r"description.{0,150}(but|workflow|also|tag|pr|push|side[\s-]effect|changelog|gh pr create|surface)", finding, re.I | re.S))
             and bool(re.search(r"(tag|gh pr create|push|changelog|side[\s-]effect)", finding, re.I))
             for finding in re.findall(r"(?ms)^[-*]\s+.+?(?=\n[-*]|\n#|\Z)", text)
         ),
         "evidence": ""},
        {"text": "at least 80% of findings cite a `file:line` reference",
         "passed": (total > 0 and cited / total >= 0.8),
         "evidence": f"{cited}/{total} bullets cite file:line"},
        {"text": "the report does not use severity tiers (no Critical/High/Medium/Low/P0/P1 labels)",
         "passed": not severity_tiers_present(text),
         "evidence": ""},
        {"text": "findings are grouped by objective (the four objective names appear as headings or list-section labels)",
         "passed": all(has_section(text, [n]) for n in ["Idempotency", "Reproducibility", "Context", "Strict"]),
         "evidence": ""},
    ]


def count_findings(text: str) -> int:
    """Count findings under the four objectives."""
    # Heuristic: bullets within sections labeled by objective names
    sections = re.split(r"(?im)^#{1,6}\s+(idempotency|reproducibility|context[\s-]management|strict[\s-]definitions)\b.*$", text)
    # The split returns: [pre, name1, body1, name2, body2, ...]
    total = 0
    for i in range(1, len(sections), 2):
        body = sections[i+1] if i+1 < len(sections) else ""
        # Stop at next ## heading
        body = re.split(r"\n#{1,3}\s", body)[0]
        for line in body.splitlines():
            s = line.strip()
            if s.startswith("-") or s.startswith("*"):
                if "no findings" in s.lower() or "no issues" in s.lower():
                    continue
                total += 1
    return total


def grade_eval1(report_path: Path, planted_skill: Path) -> list[dict]:
    """Grade eval-1 (good-skill, expects few findings)."""
    text = report_path.read_text() if report_path else ""
    findings_count = count_findings(text)
    return [
        {"text": "a markdown report file is written and the path is named in the assistant's final message",
         "passed": report_path is not None and report_path.exists(),
         "evidence": f"report at {report_path}" if report_path else "no report found"},
        {"text": "./skills/word-count/SKILL.md is byte-identical to the input fixture",
         "passed": file_unchanged(planted_skill, FIXTURE_GOOD),
         "evidence": ""},
        {"text": "total finding count across all four objectives is ≤ 4",
         "passed": findings_count <= 4,
         "evidence": f"{findings_count} findings"},
        {"text": "a finding catches the description ↔ workflow scope mismatch (set of files vs single-path)",
         "passed": bool(re.search(r"set of files|or set of|files\s*\(plural\)|plurality|description.*single[\s-]?file|single[\s-]?file.*description", text, re.I)) and
                   bool(re.search(r"description|argument[-\s]hint|inputs?", text, re.I)),
         "evidence": ""},
        {"text": "the report has a 'Passing checks' / 'Passed' section that lists at least 3 things the skill got right",
         "passed": bool(re.search(r"#{1,6}\s+passing\s+checks", text, re.I) or re.search(r"#{1,6}\s+passed", text, re.I) or re.search(r"#{1,6}\s+strengths?", text, re.I)),
         "evidence": ""},
        {"text": "the report does NOT raise a finding for missing idempotency declaration",
         "passed": not bool(re.search(r"missing idempot|no idempot|does not declare idempot|idempoten.*declaration\s*[—-]\s*missing|absence of.*idempot", text, re.I)),
         "evidence": ""},
        {"text": "the report does NOT raise a finding for missing 'when to skip'",
         "passed": not bool(re.search(r"missing\s+(when\s+to\s+skip|when-to-skip)|no\s+when\s+to\s+skip|description\s+lacks?\s+when\s+to\s+skip", text, re.I)),
         "evidence": ""},
        {"text": "the report does not use severity tiers",
         "passed": not severity_tiers_present(text),
         "evidence": ""},
    ]


def grade_eval2(report_path: Path, planted_skill: Path, final_message: str) -> list[dict]:
    """Grade eval-2 (name resolution)."""
    combined = (report_path.read_text() if report_path else "") + "\n" + final_message
    return [
        {"text": "the assistant's transcript / final message indicates the skill was resolved at ./.claude/skills/word-count/",
         "passed": bool(re.search(r"\.claude[/\\]skills[/\\]word-count", combined)),
         "evidence": ""},
        {"text": "a report file is written and its path is named in the final message",
         "passed": report_path is not None and report_path.exists(),
         "evidence": f"report at {report_path}" if report_path else "no report"},
        {"text": "./.claude/skills/word-count/SKILL.md is byte-identical to the input fixture",
         "passed": file_unchanged(planted_skill, FIXTURE_GOOD),
         "evidence": ""},
    ]


# Final messages from the agent runs (captured verbatim)
FINAL_MESSAGES = {
    (0, "with_skill"): """Audit complete. Target audited: bad-skill at .../eval-0/with_skill/work/skills/bad-skill/. Output mode: file mode. Findings: Idempotency 5, Reproducibility 9, Context management 1, Strict definitions 7, Total 22. Report: .../audit-bad-skill-2026-04-29.md. Did NOT modify input.""",
    (0, "without_skill"): """Audit complete. Report: .../AUDIT_REPORT.md. 11 issues ranked by severity (Critical/High/Medium/Low). gh pr create flagged as Critical, description as High, etc.""",
    (1, "with_skill"): """Audit complete. Target: word-count. File mode. 0 findings across all four objectives. Skill is well-formed. Report: .../audit-word-count-2026-04-29.md. Target not modified.""",
    (1, "without_skill"): """Audit complete. Report: .../AUDIT.md. Headline findings: wc -wlc column-order misconception, single vs multiple files, robustness gaps. Target not modified.""",
    (2, "with_skill"): """Audit complete. Resolver: ~/.claude/skills/word-count/ does not exist; ./.claude/skills/word-count/ match. Resolved path: .../.claude/skills/word-count/. File mode. Findings: 1 total (Strict definitions). Report: .../audit-word-count-2026-04-29.md. Target not modified.""",
    (2, "without_skill"): """Audit complete. Report: .../audit-report.md. Key findings: file --mime-type without -b prefixes path, wc column-order misconception, argument-hint singular contradicts description. Target not modified.""",
}


def find_report(out_dir: Path) -> Path | None:
    """Find the report file in outputs/."""
    if not out_dir.exists():
        return None
    md_files = sorted(out_dir.glob("*.md"))
    return md_files[0] if md_files else None


def main():
    grades_summary = []
    for eval_id in (0, 1, 2):
        for kind in ("with_skill", "without_skill"):
            run_dir = WS / f"eval-{eval_id}" / kind
            outputs_dir = run_dir / "outputs"
            report_path = find_report(outputs_dir)

            # Find the planted target skill
            if eval_id == 0:
                planted = run_dir / "work" / "skills" / "bad-skill" / "SKILL.md"
            elif eval_id == 1:
                planted = run_dir / "work" / "skills" / "word-count" / "SKILL.md"
            else:
                planted = run_dir / "work" / ".claude" / "skills" / "word-count" / "SKILL.md"

            if eval_id == 0:
                expectations = grade_eval0(report_path, planted)
            elif eval_id == 1:
                expectations = grade_eval1(report_path, planted)
            else:
                expectations = grade_eval2(report_path, planted, FINAL_MESSAGES.get((eval_id, kind), ""))

            passed = sum(1 for e in expectations if e["passed"])
            total = len(expectations)
            failed = total - passed
            timing_path = run_dir / "timing.json"
            timing = json.load(open(timing_path)) if timing_path.exists() else {}
            grading = {
                "eval_id": eval_id,
                "kind": kind,
                "report_path": str(report_path) if report_path else None,
                "summary": {
                    "passed": passed,
                    "failed": failed,
                    "total": total,
                    "pass_rate": round(passed / total, 4) if total else 0.0,
                },
                "timing": timing,
                "expectations": expectations,
            }
            run1_dir = run_dir / "run-1"
            run1_dir.mkdir(exist_ok=True)
            out_path = run1_dir / "grading.json"
            # Also copy timing into run-1 for the aggregator's discovery
            if timing:
                json.dump(timing, open(run1_dir / "timing.json", "w"), indent=2)
            json.dump(grading, open(out_path, "w"), indent=2)
            print(f"eval-{eval_id} {kind}: {passed}/{total} passed")
            grades_summary.append((eval_id, kind, passed, total))
    print()
    return grades_summary


if __name__ == "__main__":
    main()
