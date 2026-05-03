"""Grade iteration-1 runs against assertions in evals.json.

Writes grading.json into each run's directory.
"""

import json
import re
import subprocess
from pathlib import Path

WORKSPACE = Path(__file__).parent
SKILL = WORKSPACE.parent.parent / "review-file-library"
EVALS = SKILL / "evals" / "evals.json"
FIXTURES = SKILL / "evals" / "fixtures"


def load_evals():
    return json.loads(EVALS.read_text())["evals"]


def read_audit(run_dir, fixture_subdir):
    p = run_dir / "outputs" / fixture_subdir / "AUDIT.md"
    return p.read_text() if p.exists() else None


def list_files(d):
    return sorted(p.name for p in d.iterdir() if p.is_file())


def diff_against_fixture(run_dir, fixture_name, paths):
    """Return list of paths that differ from the canonical fixture."""
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
    """True if text has a markdown section with exactly this heading line."""
    pattern = r"^" + re.escape(heading) + r"\s*$"
    return bool(re.search(pattern, text, re.MULTILINE))


def grade_eval0(run_dir, audit, agent_response=None):
    """Grade eval-0 (kvr-text-scaffold) against its assertions."""
    fixture_files = ["tokenizer.go", "parser.go", "printer.go",
                     "tokenizer_test.go", "parser_test.go", "printer_test.go", "SPEC.md"]
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
         "evidence": "header text" if "text file library" in text else "no 'text file library' phrase"},
        {"id": "header-has-test-status", "passed": "**tests:**" in text or "tests:" in text,
         "evidence": "tests: line in header"},
        {"id": "header-has-spec-line-count",
         "passed": bool(re.search(r"SPEC\.md\s*\(\s*\d+\s*lines", audit, re.IGNORECASE)),
         "evidence": "SPEC.md (N lines) format"},
        {"id": "tokenizer-section-present",
         "passed": has_section(audit, "## Tokenizer findings"),
         "evidence": "## Tokenizer findings header"},
        {"id": "parser-section-present",
         "passed": has_section(audit, "## Parser findings"),
         "evidence": "## Parser findings header"},
        {"id": "printer-section-present",
         "passed": has_section(audit, "## Printer findings"),
         "evidence": "## Printer findings header"},
        {"id": "missing-token-types-category",
         "passed": has_section(audit, "### Missing token types"),
         "evidence": "### Missing token types subsection"},
        {"id": "grammar-gaps-category",
         "passed": has_section(audit, "### Grammar gaps"),
         "evidence": "### Grammar gaps subsection"},
        {"id": "printer-rules-category",
         "passed": has_section(audit, "### Missing/incomplete printer rules"),
         "evidence": "### Missing/incomplete printer rules subsection"},
        {"id": "round-trip-category",
         "passed": has_section(audit, "### Round-trip test coverage"),
         "evidence": "### Round-trip test coverage subsection"},
        {"id": "drift-categories-present",
         "passed": len(re.findall(r"^### Drift\s*$", audit, re.MULTILINE)) >= 3,
         "evidence": f"{len(re.findall(r'^### Drift\\s*$', audit, re.MULTILINE))} ### Drift sections"},
        {"id": "uses-blocker-severity",
         "passed": "[blocker]" in text,
         "evidence": "[blocker] severity prefix used"},
        {"id": "findings-cite-spec-section",
         "passed": bool(re.search(r"SPEC\.md\s*§", audit)),
         "evidence": "SPEC.md § citations present"},
        {"id": "findings-cite-go-file",
         "passed": any(f in audit for f in ["tokenizer.go", "parser.go", "printer.go"]),
         "evidence": "Go file paths cited"},
        {"id": "missing-token-string",
         "passed": "tokenstring" in text or ("string" in text and "token" in text),
         "evidence": "TokenString gap noted"},
        {"id": "missing-record-ast",
         "passed": "record" in text and ("ast" in text or "struct" in text or "type" in text),
         "evidence": "Record AST gap noted"},
        {"id": "missing-block-ast",
         "passed": "block" in text and ("ast" in text or "struct" in text or "type" in text),
         "evidence": "Block AST gap noted"},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover _audit_*.md files: {[p.name for p in leftover_scratch]}"},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0,
         "evidence": f"differing source files: {edits_to_source}"},
        {"id": "no-test-edits",
         "passed": len(edits_to_tests) == 0,
         "evidence": f"differing test files: {edits_to_tests}"},
        {"id": "no-spec-edits",
         "passed": len(edits_to_spec) == 0,
         "evidence": f"differing spec files: {edits_to_spec}"},
    ]


def grade_eval1(run_dir, audit):
    """Grade eval-1 (tlv-binary)."""
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
        {"id": "audit-md-exists", "passed": True, "evidence": "AUDIT.md exists"},
        {"id": "header-says-binary", "passed": "binary file library" in text,
         "evidence": "header identifies binary"},
        {"id": "header-has-test-status",
         "passed": "**tests:**" in text or "tests:" in text,
         "evidence": "tests: line in header"},
        {"id": "header-cites-chunked-tree",
         "passed": "structures/" in text and "encoding-tables/" in text,
         "evidence": "chunked tree mentioned in header"},
        {"id": "types-section-present",
         "passed": has_section(audit, "## Types findings"),
         "evidence": "## Types findings header"},
        {"id": "decoder-section-present",
         "passed": has_section(audit, "## Decoder findings"),
         "evidence": "## Decoder findings header"},
        {"id": "encoder-section-present",
         "passed": has_section(audit, "## Encoder findings"),
         "evidence": "## Encoder findings header"},
        {"id": "missing-structs-category",
         "passed": has_section(audit, "### Missing struct types"),
         "evidence": "### Missing struct types subsection"},
        {"id": "encoding-tables-category",
         "passed": has_section(audit, "### Encoding-table coverage"),
         "evidence": "### Encoding-table coverage subsection"},
        {"id": "unread-fields-category",
         "passed": has_section(audit, "### Unread fields"),
         "evidence": "### Unread fields subsection"},
        {"id": "checksum-category",
         "passed": has_section(audit, "### Missing length/offset/checksum checks"),
         "evidence": "### Missing length/offset/checksum checks subsection"},
        {"id": "unwritten-fields-category",
         "passed": has_section(audit, "### Unwritten fields"),
         "evidence": "### Unwritten fields subsection"},
        {"id": "encoder-round-trip-category",
         "passed": has_section(audit, "### Round-trip test coverage"),
         "evidence": "### Round-trip test coverage subsection"},
        {"id": "findings-cite-chunked-files",
         "passed": any(f in audit for f in
                       ["structures/header.md", "structures/record.md",
                        "structures/trailer.md", "encoding-tables/record-type.md"]),
         "evidence": "chunked spec file paths cited in findings"},
        {"id": "crc32-blocker",
         "passed": any(k in text for k in crc_keywords) and "[blocker]" in text,
         "evidence": "CRC/checksum [blocker] finding"},
        {"id": "record-type-enum-blocker",
         "passed": "recordtype" in text and "[blocker]" in text,
         "evidence": "RecordType [blocker] finding"},
        {"id": "header-flags-bit-field",
         "passed": ("flags" in text and "bit" in text) or "compressed" in text or "encrypted" in text,
         "evidence": "Header.Flags bit field gap noted"},
        {"id": "missing-header-struct",
         "passed": "header" in text and ("struct" in text or "type" in text),
         "evidence": "Header struct gap noted"},
        {"id": "missing-record-struct",
         "passed": re.search(r"\brecord\b.*struct|struct.*\brecord\b", text) is not None,
         "evidence": "Record struct gap noted"},
        {"id": "missing-trailer-struct",
         "passed": "trailer" in text and ("struct" in text or "type" in text),
         "evidence": "Trailer struct gap noted"},
        {"id": "uses-blocker-severity",
         "passed": "[blocker]" in text,
         "evidence": "[blocker] severity prefix used"},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover scratch files"},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0,
         "evidence": f"differing source files: {edits_to_source}"},
        {"id": "no-spec-edits",
         "passed": len(edits_to_spec) == 0,
         "evidence": f"differing spec files: {edits_to_spec}"},
    ]


def grade_eval2(run_dir, agent_response):
    """Grade eval-2 (kvr-no-spec-refusal)."""
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
         "passed": "spec.md" in response_text and ("missing" in response_text or "no " in response_text or "no spec" in response_text),
         "evidence": "response identifies missing SPEC.md"},
        {"id": "agent-points-at-extract-skill",
         "passed": "extract-text-spec" in response_text or "extract_text_spec" in response_text,
         "evidence": "response names extract-text-spec"},
        {"id": "no-source-edits",
         "passed": len(edits_to_source) == 0,
         "evidence": f"differing source/test files: {edits_to_source}"},
        {"id": "no-leftover-scratch",
         "passed": len(leftover_scratch) == 0,
         "evidence": f"{len(leftover_scratch)} leftover scratch files"},
    ]


def normalize_for_aggregator(raw_expectations):
    """Add `text` field (copy of id) so viewer/aggregator can render it."""
    out = []
    for e in raw_expectations:
        out.append({
            "text": e["id"],
            "passed": e["passed"],
            "evidence": e.get("evidence", ""),
        })
    return out


def main():
    iter1 = WORKSPACE
    plan = [
        ("kvr-text-scaffold", "kvr-text", grade_eval0),
        ("tlv-binary-multi-file-spec", "tlv-binary", grade_eval1),
        ("kvr-no-spec-refusal", "kvr-no-spec", grade_eval2),
    ]
    for eval_name, fixture, grader in plan:
        for variant in ("with_skill", "without_skill"):
            run_dir = iter1 / f"eval-{eval_name}" / variant / "run-1"
            if eval_name == "kvr-no-spec-refusal":
                resp = (run_dir / "agent_response.txt").read_text()
                raw = grader(run_dir, resp)
            else:
                audit = read_audit(run_dir, fixture)
                raw = grader(run_dir, audit)
            expectations = normalize_for_aggregator(raw)
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
