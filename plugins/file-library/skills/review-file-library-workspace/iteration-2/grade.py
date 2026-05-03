"""Grade iteration-2 runs.

Same checks as iteration-1 plus two new assertions per text/binary eval:
the "test gap" findings should now appear as [blocker] (not [warning]).

Variants are with_skill (new) and old_skill (= reused iteration-1 with-skill outputs).
"""

import json
import re
from pathlib import Path

WORKSPACE = Path(__file__).parent
SKILL = WORKSPACE.parent.parent / "review-file-library"
EVALS = SKILL / "evals" / "evals.json"
FIXTURES = SKILL / "evals" / "fixtures"


def read_audit(run_dir, fixture_subdir):
    p = run_dir / "outputs" / fixture_subdir / "AUDIT.md"
    return p.read_text() if p.exists() else None


def diff_against_fixture(run_dir, fixture_name, paths):
    fixture = FIXTURES / fixture_name
    target = run_dir / "outputs" / fixture_name
    differing = []
    for rel in paths:
        a = fixture / rel
        b = target / rel
        if not a.exists() and not b.exists():
            continue
        if not a.exists() or not b.exists():
            differing.append(rel)
            continue
        if a.read_bytes() != b.read_bytes():
            differing.append(rel)
    return differing


def has_section(text, heading):
    pattern = r"^" + re.escape(heading) + r"\s*$"
    return bool(re.search(pattern, text, re.MULTILINE))


def line_contains_blocker_about(text, *keywords):
    """True if any line in text has both [blocker] AND all keywords."""
    for line in text.splitlines():
        low = line.lower()
        if "[blocker]" in low and all(k.lower() in low for k in keywords):
            return True
    return False


def grade_eval0(run_dir, audit, agent_response=None):
    target_dir = run_dir / "outputs" / "kvr-text"
    leftover_scratch = list(target_dir.glob("_audit_*.md"))
    edits_to_source = diff_against_fixture(
        run_dir, "kvr-text",
        ["tokenizer.go", "parser.go", "printer.go"]
    )
    edits_to_tests = diff_against_fixture(
        run_dir, "kvr-text",
        ["tokenizer_test.go", "parser_test.go", "printer_test.go"]
    )
    edits_to_spec = diff_against_fixture(run_dir, "kvr-text", ["SPEC.md"])

    if audit is None:
        return [{"id": "audit-md-exists", "passed": False,
                 "evidence": "AUDIT.md does not exist"}]

    text = audit.lower()
    return [
        {"id": "audit-md-exists", "passed": True, "evidence": "AUDIT.md exists"},
        {"id": "header-says-text", "passed": "text file library" in text,
         "evidence": "header text"},
        {"id": "header-has-test-status", "passed": "**tests:**" in text or "tests:" in text,
         "evidence": "tests: line"},
        {"id": "header-has-spec-line-count",
         "passed": bool(re.search(r"SPEC\.md\s*\(\s*\d+\s*lines", audit, re.IGNORECASE)),
         "evidence": "SPEC.md (N lines)"},
        {"id": "tokenizer-section-present",
         "passed": has_section(audit, "## Tokenizer findings"), "evidence": ""},
        {"id": "parser-section-present",
         "passed": has_section(audit, "## Parser findings"), "evidence": ""},
        {"id": "printer-section-present",
         "passed": has_section(audit, "## Printer findings"), "evidence": ""},
        {"id": "missing-token-types-category",
         "passed": has_section(audit, "### Missing token types"), "evidence": ""},
        {"id": "grammar-gaps-category",
         "passed": has_section(audit, "### Grammar gaps"), "evidence": ""},
        {"id": "printer-rules-category",
         "passed": has_section(audit, "### Missing/incomplete printer rules"), "evidence": ""},
        {"id": "round-trip-category",
         "passed": has_section(audit, "### Round-trip test coverage"), "evidence": ""},
        {"id": "drift-categories-present",
         "passed": len(re.findall(r"^### Drift\s*$", audit, re.MULTILINE)) >= 3,
         "evidence": f"{len(re.findall(r'^### Drift\\s*$', audit, re.MULTILINE))} drift sections"},
        {"id": "uses-blocker-severity",
         "passed": "[blocker]" in text, "evidence": ""},
        {"id": "findings-cite-spec-section",
         "passed": bool(re.search(r"SPEC\.md\s*§", audit)), "evidence": ""},
        {"id": "findings-cite-go-file",
         "passed": any(f in audit for f in ["tokenizer.go", "parser.go", "printer.go"]),
         "evidence": ""},
        {"id": "missing-token-string",
         "passed": "tokenstring" in text or ("string" in text and "token" in text), "evidence": ""},
        {"id": "missing-record-ast",
         "passed": "record" in text and ("ast" in text or "struct" in text or "type" in text),
         "evidence": ""},
        {"id": "missing-block-ast",
         "passed": "block" in text and ("ast" in text or "struct" in text or "type" in text),
         "evidence": ""},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover files"},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0, "evidence": str(edits_to_source)},
        {"id": "no-test-edits",
         "passed": len(edits_to_tests) == 0, "evidence": str(edits_to_tests)},
        {"id": "no-spec-edits",
         "passed": len(edits_to_spec) == 0, "evidence": str(edits_to_spec)},
        # New iteration-2 assertions:
        {"id": "round-trip-gap-is-blocker",
         "passed": line_contains_blocker_about(audit, "round-trip") or
                   line_contains_blocker_about(audit, "round trip"),
         "evidence": "[blocker] round-trip line"},
        {"id": "untested-token-class-is-blocker",
         "passed": line_contains_blocker_about(audit, "untested") or
                   line_contains_blocker_about(audit, "no test") or
                   line_contains_blocker_about(audit, "no tokenizer test") or
                   line_contains_blocker_about(audit, "tokenizer_test.go") or
                   line_contains_blocker_about(audit, "unverified"),
         "evidence": "[blocker] untested-token-class line"},
    ]


def grade_eval1(run_dir, audit):
    target_dir = run_dir / "outputs" / "tlv-binary"
    leftover_scratch = list(target_dir.glob("_audit_*.md"))
    edits_to_source = diff_against_fixture(
        run_dir, "tlv-binary",
        ["types.go", "decoder.go", "encoder.go"]
    )
    edits_to_spec = diff_against_fixture(
        run_dir, "tlv-binary",
        ["SPEC.md", "structures/header.md", "structures/record.md",
         "structures/trailer.md", "encoding-tables/record-type.md"]
    )

    if audit is None:
        return [{"id": "audit-md-exists", "passed": False,
                 "evidence": "AUDIT.md does not exist"}]

    text = audit.lower()
    crc_keywords = ["crc32", "crc 32", "checksum", "trailer.crc"]
    return [
        {"id": "audit-md-exists", "passed": True, "evidence": ""},
        {"id": "header-says-binary", "passed": "binary file library" in text, "evidence": ""},
        {"id": "header-has-test-status",
         "passed": "**tests:**" in text or "tests:" in text, "evidence": ""},
        {"id": "header-cites-chunked-tree",
         "passed": "structures/" in text and "encoding-tables/" in text, "evidence": ""},
        {"id": "types-section-present",
         "passed": has_section(audit, "## Types findings"), "evidence": ""},
        {"id": "decoder-section-present",
         "passed": has_section(audit, "## Decoder findings"), "evidence": ""},
        {"id": "encoder-section-present",
         "passed": has_section(audit, "## Encoder findings"), "evidence": ""},
        {"id": "missing-structs-category",
         "passed": has_section(audit, "### Missing struct types"), "evidence": ""},
        {"id": "encoding-tables-category",
         "passed": has_section(audit, "### Encoding-table coverage"), "evidence": ""},
        {"id": "unread-fields-category",
         "passed": has_section(audit, "### Unread fields"), "evidence": ""},
        {"id": "checksum-category",
         "passed": has_section(audit, "### Missing length/offset/checksum checks"), "evidence": ""},
        {"id": "unwritten-fields-category",
         "passed": has_section(audit, "### Unwritten fields"), "evidence": ""},
        {"id": "encoder-round-trip-category",
         "passed": has_section(audit, "### Round-trip test coverage"), "evidence": ""},
        {"id": "findings-cite-chunked-files",
         "passed": any(f in audit for f in
                       ["structures/header.md", "structures/record.md",
                        "structures/trailer.md", "encoding-tables/record-type.md"]),
         "evidence": ""},
        {"id": "crc32-blocker",
         "passed": any(k in text for k in crc_keywords) and "[blocker]" in text,
         "evidence": ""},
        {"id": "record-type-enum-blocker",
         "passed": "recordtype" in text and "[blocker]" in text, "evidence": ""},
        {"id": "header-flags-bit-field",
         "passed": ("flags" in text and "bit" in text) or "compressed" in text or "encrypted" in text,
         "evidence": ""},
        {"id": "missing-header-struct",
         "passed": "header" in text and ("struct" in text or "type" in text), "evidence": ""},
        {"id": "missing-record-struct",
         "passed": re.search(r"\brecord\b.*struct|struct.*\brecord\b", text) is not None,
         "evidence": ""},
        {"id": "missing-trailer-struct",
         "passed": "trailer" in text and ("struct" in text or "type" in text), "evidence": ""},
        {"id": "uses-blocker-severity",
         "passed": "[blocker]" in text, "evidence": ""},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover files"},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0, "evidence": str(edits_to_source)},
        {"id": "no-spec-edits",
         "passed": len(edits_to_spec) == 0, "evidence": str(edits_to_spec)},
        # New iteration-2 assertions:
        {"id": "round-trip-gap-is-blocker",
         "passed": line_contains_blocker_about(audit, "round-trip") or
                   line_contains_blocker_about(audit, "round trip"),
         "evidence": "[blocker] round-trip line"},
        {"id": "error-chain-test-gap-is-blocker",
         "passed": line_contains_blocker_about(audit, "errors.is") or
                   line_contains_blocker_about(audit, "errors.as") or
                   line_contains_blocker_about(audit, "error chain") or
                   line_contains_blocker_about(audit, "fielderror") or
                   line_contains_blocker_about(audit, "failure-path") or
                   line_contains_blocker_about(audit, "failure path") or
                   line_contains_blocker_about(audit, "rejection"),
         "evidence": "[blocker] error-chain test gap line"},
    ]


def grade_eval2(run_dir, agent_response):
    target_dir = run_dir / "outputs" / "kvr-no-spec"
    audit_md = target_dir / "AUDIT.md"
    leftover_scratch = list(target_dir.glob("_audit_*.md"))
    edits_to_source = diff_against_fixture(
        run_dir, "kvr-no-spec",
        ["tokenizer.go", "parser.go", "printer.go",
         "tokenizer_test.go", "parser_test.go", "printer_test.go"]
    )
    response_text = (agent_response or "").lower()
    return [
        {"id": "no-audit-md", "passed": not audit_md.exists(),
         "evidence": "AUDIT.md absent" if not audit_md.exists() else "AUDIT.md present"},
        {"id": "agent-detected-missing-spec",
         "passed": "spec.md" in response_text and ("missing" in response_text or "no " in response_text),
         "evidence": ""},
        {"id": "agent-points-at-extract-skill",
         "passed": "extract-text-spec" in response_text or "extract_text_spec" in response_text,
         "evidence": ""},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0, "evidence": str(edits_to_source)},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover files"},
    ]


def normalize(raw_expectations):
    return [{"text": e["id"], "passed": e["passed"], "evidence": e.get("evidence", "")}
            for e in raw_expectations]


def main():
    plan = [
        ("kvr-text-scaffold", "kvr-text", grade_eval0),
        ("tlv-binary-multi-file-spec", "tlv-binary", grade_eval1),
        ("kvr-no-spec-refusal", "kvr-no-spec", grade_eval2),
    ]
    for eval_name, fixture, grader in plan:
        for variant in ("with_skill", "old_skill"):
            run_dir = WORKSPACE / f"eval-{eval_name}" / variant / "run-1"
            if eval_name == "kvr-no-spec-refusal":
                resp_path = run_dir / "agent_response.txt"
                resp = resp_path.read_text() if resp_path.exists() else ""
                raw = grader(run_dir, resp)
            else:
                audit = read_audit(run_dir, fixture)
                raw = grader(run_dir, audit)
            expectations = normalize(raw)
            passed = sum(1 for e in expectations if e["passed"])
            total = len(expectations)
            grading = {
                "expectations": expectations,
                "summary": {
                    "pass_rate": (passed / total) if total else 0.0,
                    "passed": passed,
                    "failed": total - passed,
                    "total": total,
                },
            }
            (run_dir / "grading.json").write_text(json.dumps(grading, indent=2))
            print(f"{eval_name}/{variant}: {passed}/{total}")


if __name__ == "__main__":
    main()
