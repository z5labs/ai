#!/usr/bin/env python3
"""Grade implement-go-text-file-library eval runs.

Usage: python grade.py <iteration_dir>

Walks each eval-N/<config>/run-1/outputs/kvr/ directory, evaluates assertions,
and writes grading.json into each run-1/ directory in the format expected by
the eval-viewer (each expectation has fields: text, passed, evidence).
"""
import json
import re
import sys
from pathlib import Path


def load_metadata(eval_dir: Path) -> dict:
    return json.loads((eval_dir / "eval_metadata.json").read_text())


def read(path: Path) -> str:
    return path.read_text() if path.exists() else ""


def go_test_log_passed(run_dir: Path) -> bool:
    log = run_dir / "outputs" / "verify.log"
    if not log.exists():
        return False
    txt = log.read_text()
    if "FAIL" in txt or "build failed" in txt or "compile" in txt.lower() and "error" in txt.lower():
        return False
    return "ok" in txt or "PASS" in txt


def has_t_parallel_at_both_levels(text: str) -> bool:
    return text.count("t.Parallel()") >= 2


CONTEXT_FILENAMES = ("_context_tokens.md", "_context_ast.md")
SOURCE_FILENAMES = ("tokenizer.go", "parser.go", "printer.go")


def skill_md_text() -> str:
    """Read SKILL.md from the parent directory of this grade.py."""
    skill_md_path = Path(__file__).resolve().parent.parent / "SKILL.md"
    return skill_md_path.read_text() if skill_md_path.exists() else ""


def extract_section(text: str, heading: str) -> str:
    """Return the body of a top-level `## heading` section.

    Walks lines and tracks fenced code blocks so that `## ...` markers inside
    fenced templates (e.g. the literal `## TokenType` line shown in the
    context-summary shape templates) are not mistaken for the next section's
    heading. Returns "" if the heading is not found.
    """
    lines = text.splitlines()
    in_fence = False
    start = -1
    end = len(lines)
    for i, line in enumerate(lines):
        if line.startswith("```"):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if start == -1:
            if line.strip() == f"## {heading}":
                start = i + 1
        elif line.startswith("## "):
            end = i
            break
    if start == -1:
        return ""
    return "\n".join(lines[start:end])


def assertion_context_summary_spec_tightened() -> tuple[bool, str]:
    """Verify SKILL.md specifies the tightened _context_*.md format and cap.

    Acceptance from issue #46: the SKILL.md must specify, for each
    _context_*.md, (a) a strict signature-only format with no rationale or
    examples, (b) a hard 400-line cap, and (c) that exceeding the cap signals
    the work-unit was sized too large and should be chunked.

    The check is anchored to the `## Context summary format` section so that
    stray phrases elsewhere in the SKILL.md (e.g. an unrelated paragraph that
    happens to contain "no examples") cannot cause a false positive. Both
    per-file filenames must also appear inside that section, so the spec
    cannot silently regress to covering only one of them.
    """
    text = skill_md_text()
    section = extract_section(text, "Context summary format")
    findings = []

    section_present = bool(section.strip())
    findings.append(f"section_present={section_present}")
    if not section_present:
        return False, "; ".join(findings) + " (no `## Context summary format` section in SKILL.md)"

    section_lower = section.lower()

    # (a) Strict signature-only format, no rationale, no examples.
    strict_format = (
        "signature only" in section_lower
        and "no rationale" in section_lower
        and "no examples" in section_lower
    )
    findings.append(f"strict_format={strict_format}")

    # (b) Hard 400-line cap.
    line_cap = "400" in section and ("hard cap" in section_lower or "400 lines" in section_lower)
    findings.append(f"400_line_cap={line_cap}")

    # (c) Cap overflow → chunk-and-relaunch protocol.
    chunk_protocol = "sized too large" in section_lower and "chunk" in section_lower
    findings.append(f"chunk_on_overflow={chunk_protocol}")

    # (d) Both per-file shapes are documented inside this section.
    missing = [name for name in CONTEXT_FILENAMES if name not in section]
    files_documented = not missing
    findings.append(
        f"files_documented={files_documented}"
        + (f" (missing: {missing})" if missing else "")
    )

    ok = strict_format and line_cap and chunk_protocol and files_documented
    return ok, "; ".join(findings)


def assertion_phase_chunking_spec() -> tuple[bool, str]:
    """Verify SKILL.md specifies the issue #47 phase-chunking protocol.

    Acceptance from issue #47: SKILL.md must specify (a) an up-front scope
    gate in `## Before you start` — with a numeric threshold, partitioning
    into sub-units, and an instruction to announce the plan to the user
    before any subagent launches — and (b) a `## Phase chunking` section
    saying sub-calls run serially, append to the running `_context_*.md`,
    and do not full-read the growing source file.

    Anchored to those two sections so wording elsewhere in the SKILL.md
    cannot cause a false positive (the words "partition" and "serial"
    already appear in the unrelated tokenizer/parser discussion).
    """
    text = skill_md_text()
    findings = []

    # Part 1: scope gate in `## Before you start`.
    before_start = extract_section(text, "Before you start")
    before_lower = before_start.lower()

    bs_present = bool(before_start.strip())
    findings.append(f"before_start_present={bs_present}")
    if not bs_present:
        return False, "; ".join(findings) + " (no `## Before you start` section)"

    has_scope_gate = "scope gate" in before_lower
    findings.append(f"scope_gate={has_scope_gate}")

    # Numeric threshold: at least one >=3-digit number — the threshold
    # itself (e.g. "600", "300") is what we want, not the existing
    # "400-line cap" which only appears in `## Context summary format`.
    has_threshold = bool(re.search(r"\b\d{3,}\b", before_start))
    findings.append(f"numeric_threshold={has_threshold}")

    has_partition_subunits = (
        "partition" in before_lower
        and ("sub-unit" in before_lower or "sub-units" in before_lower)
    )
    findings.append(f"partition_subunits={has_partition_subunits}")

    has_announce = (
        "tell the user" in before_lower
        or "announce" in before_lower
        or "up front" in before_lower
    )
    findings.append(f"announce={has_announce}")

    gate_ok = (
        has_scope_gate
        and has_threshold
        and has_partition_subunits
        and has_announce
    )

    # Part 2: `## Phase chunking` section.
    chunking = extract_section(text, "Phase chunking")
    chunking_lower = chunking.lower()

    chunking_present = bool(chunking.strip())
    findings.append(f"chunking_section_present={chunking_present}")
    if not chunking_present:
        return False, "; ".join(findings) + " (no `## Phase chunking` section)"

    has_serial = "serial" in chunking_lower  # matches "serial" and "serially"
    findings.append(f"serial={has_serial}")

    has_append = "append" in chunking_lower
    findings.append(f"append={has_append}")

    # No full-Read of the growing source file: a forbidding phrase plus a
    # source filename plus the word "read" must all appear inside the
    # `## Phase chunking` section. The forbidding-phrase set covers the
    # current "no full-`Read`" / "without a fresh whole-file read" /
    # "never the whole file" wordings without pinning the exact phrase.
    forbid_phrase = any(
        p in chunking_lower for p in ("no full", "whole file", "whole-file")
    )
    has_source_filename = any(name in chunking for name in SOURCE_FILENAMES)
    forbids_full_read = (
        forbid_phrase and has_source_filename and "read" in chunking_lower
    )
    findings.append(
        f"forbids_full_read={forbids_full_read}"
        f" (forbid_phrase={forbid_phrase}, has_source={has_source_filename})"
    )

    # Both running-summary filenames present in the chunking section, so
    # the spec cannot silently regress to covering only one of them.
    missing = [name for name in CONTEXT_FILENAMES if name not in chunking]
    files_documented = not missing
    findings.append(
        f"summary_filenames={files_documented}"
        + (f" (missing: {missing})" if missing else "")
    )

    chunk_ok = (
        chunking_present
        and has_serial
        and has_append
        and forbids_full_read
        and files_documented
    )

    ok = gate_ok and chunk_ok
    return ok, "; ".join(findings)


def assertion_eval0_string_record(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    tokenizer = read(pkg / "tokenizer.go")
    parser = read(pkg / "parser.go")
    printer = read(pkg / "printer.go")
    tokenizer_test = read(pkg / "tokenizer_test.go")
    parser_test = read(pkg / "parser_test.go")
    printer_test = read(pkg / "printer_test.go")
    all_tests = tokenizer_test + parser_test + printer_test

    if asid == "token-string-defined":
        ok = re.search(r"\bTokenString\b", tokenizer) is not None
        return ok, "TokenString constant defined" if ok else "no TokenString in tokenizer.go"

    if asid == "tokenizer-handles-quoted-string":
        # Look for an action or logic that reads `"` and yields a TokenString
        ok = ('"' in tokenizer or "'\"'" in tokenizer) and "TokenString" in tokenizer
        return ok, "tokenizer handles quoted strings" if ok else "no string-tokenizing logic"

    if asid == "record-ast-node":
        if not re.search(r"type\s+Record\s+struct", parser):
            return False, "no Record struct in parser.go"
        # check at least Type, Key, Value fields (or close equivalents)
        fields = []
        for f in ("Type", "Key", "Value", "Name"):
            if re.search(rf"\b{f}\b", parser):
                fields.append(f)
        has_required = ("Value" in fields) and (("Key" in fields) or ("Name" in fields))
        return has_required, f"Record struct present with fields {fields}" if has_required else f"Record missing Key/Value (found {fields})"

    if asid == "file-holds-records":
        ok = re.search(r"type\s+File\s+struct\s*\{[^}]*Records\b", parser) is not None
        return ok, "File.Records present" if ok else "no Records slice on File"

    if asid == "parse-record-action":
        # Look for any parser action function involving Record
        ok = re.search(r"\bparseRecord\w*\b", parser) is not None
        return ok, "parseRecord-style action present" if ok else "no parseRecord function"

    if asid == "expect-used":
        # The new code should call p.expect for token-type checks. Look for
        # at least one p.expect call and no inline token-type comparisons in
        # the parser file (allowing the existing expect helper definition).
        has_expect_call = re.search(r"p\.expect\s*\(", parser) is not None
        # Inline token comparisons are tok.Type == TokenX patterns OUTSIDE the
        # expect helper. Strip the expect helper body first.
        stripped = re.sub(r"func\s*\(\s*p\s*\*parser\s*\)\s*expect[\s\S]*?\n\}", "", parser, count=1)
        inline_compares = len(re.findall(r"\.Type\s*==\s*Token\w+", stripped))
        ok = has_expect_call and inline_compares == 0
        return ok, f"p.expect used; inline compares={inline_compares}"

    if asid == "print-record-rule":
        ok = re.search(r"\bprintRecord\w*\b", printer) is not None or \
             re.search(r"Record\b", printer) is not None
        # tighten: must reference Record by name in a function context
        ok = ok and ("Record" in printer)
        return ok, "printer handles Record" if ok else "no Record-printing logic"

    if asid == "tokenizer-test-exact-pos":
        # Look for a string-token test asserting Pos{Line: N, Column: M} or similar
        has_pos = re.search(r"Pos\s*\{\s*Line\s*:\s*\d+\s*,\s*Column\s*:\s*\d+", tokenizer_test) is not None
        has_string = "TokenString" in tokenizer_test or '"hello"' in tokenizer_test or "STRING" in tokenizer_test
        ok = has_pos and has_string
        return ok, f"exact Pos in tokenizer test (has_pos={has_pos}, has_string_test={has_string})"

    if asid == "parser-test-via-parse":
        has_parse = "Parse(" in parser_test
        # No hand-constructed Record literal as a test expectation (allowing
        # parsed-result equality which gets a *File, not a Record literal)
        # Disallow `Record{...}` literal in expectation-context
        bad_record_literal = re.search(r"Record\s*\{[^}]*Type\s*:", parser_test) is not None
        ok = has_parse and not bad_record_literal
        return ok, f"parser tests use Parse() (has_parse={has_parse}, bad_literal={bad_record_literal})"

    if asid == "printer-direct-test":
        # Test that calls Print on a *File and asserts string equality
        has_print = re.search(r"\bPrint\s*\(", printer_test) is not None
        has_assert_string = "buf.String()" in printer_test or "buf.Bytes()" in printer_test
        ok = has_print and has_assert_string
        return ok, f"direct print test (has_print={has_print}, has_assert={has_assert_string})"

    if asid == "printer-round-trip-test":
        ok = ("RoundTrip" in printer_test or "round_trip" in printer_test or "round" in printer_test.lower()) \
             and "Parse" in printer_test and "Print" in printer_test
        return ok, "round-trip test present" if ok else "no round-trip test"

    if asid == "tests-parallel":
        for name, txt in [("tokenizer_test.go", tokenizer_test), ("parser_test.go", parser_test), ("printer_test.go", printer_test)]:
            if txt.count("t.Parallel()") < 2:
                return False, f"{name} insufficient t.Parallel"
        return True, "t.Parallel at both levels in all test files"

    if asid == "tests-testify-require":
        ok = "github.com/stretchr/testify/require" in all_tests
        return ok, "tests use testify/require" if ok else "no testify/require import"

    if asid == "go-test-passes":
        ok = go_test_log_passed(run_dir)
        return ok, "go test -race ./... passes" if ok else "go test failed (see verify.log)"

    if asid == "no-scratch-files-left":
        leftovers = [p.name for p in pkg.glob("_*.md")]
        return len(leftovers) == 0, "no _*.md scratch files" if not leftovers else f"leftover scratch: {leftovers}"

    if asid == "no-full-spec-copy":
        leftovers = [p.name for p in pkg.glob("_spec*.md")]
        return len(leftovers) == 0, "no _spec*.md scratch copies" if not leftovers else f"spec copy leftover: {leftovers}"

    if asid == "context-summary-spec-tightened":
        return assertion_context_summary_spec_tightened()

    if asid == "phase-chunking-spec-tightened":
        return assertion_phase_chunking_spec()

    return False, f"unknown assertion id: {asid}"


def assertion_eval1_block(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    tokenizer = read(pkg / "tokenizer.go")
    parser = read(pkg / "parser.go")
    printer = read(pkg / "printer.go")
    tokenizer_test = read(pkg / "tokenizer_test.go")
    parser_test = read(pkg / "parser_test.go")
    printer_test = read(pkg / "printer_test.go")

    if asid == "block-ast-node":
        ok = re.search(r"type\s+Block\s+struct", parser) is not None
        if ok:
            has_records = re.search(r"Records\s+\[\]Record", parser) is not None or \
                          re.search(r"Records\s+\[\]\*?Record", parser) is not None
            return has_records, "Block struct with Records slice" if has_records else "Block struct missing Records slice"
        return False, "no Block struct in parser.go"

    if asid == "file-holds-blocks":
        ok = re.search(r"type\s+File\s+struct\s*\{[^}]*Blocks\b", parser) is not None
        return ok, "File.Blocks slice present" if ok else "no Blocks slice on File"

    if asid == "brace-tokens":
        # Either distinct token types OR symbol tokens with brace values
        has_distinct_types = "TokenLBrace" in tokenizer or "TokenOpenBrace" in tokenizer
        has_distinct_types = has_distinct_types and ("TokenRBrace" in tokenizer or "TokenCloseBrace" in tokenizer)
        # Or yields TokenSymbol with values { and }
        has_symbol_braces = "TokenSymbol" in tokenizer and "{" in tokenizer and "}" in tokenizer
        ok = has_distinct_types or has_symbol_braces
        return ok, "brace tokens distinguishable" if ok else "no brace handling visible"

    if asid == "inner-action-loop-used":
        # Look for a parserAction[*Block] type variable being driven by a for-loop.
        # Heuristic 1: multiple parseBlockX functions (parseBlockOpen, parseBlockMember, etc.)
        block_actions = re.findall(r"\bparseBlock\w+\b", parser)
        unique_block_actions = set(block_actions)
        h1 = len(unique_block_actions) >= 2
        # Heuristic 2: explicit parserAction[*Block] type usage in a for-loop driver
        h2 = re.search(r"parserAction\[\*Block\]", parser) is not None
        # Heuristic 3: a for-loop in parseBlock that drives an action variable
        h3 = re.search(r"for\s+action\s*[:=]+[^;]*action\s*!=\s*nil", parser) is not None
        ok = h1 or (h2 and h3)
        return ok, f"inner action loop (named_actions={sorted(unique_block_actions)}, parserAction[*Block]={h2}, for-action-loop={h3})"

    if asid == "expect-used":
        has_expect_call = re.search(r"p\.expect\s*\(", parser) is not None
        stripped = re.sub(r"func\s*\(\s*p\s*\*parser\s*\)\s*expect[\s\S]*?\n\}", "", parser, count=1)
        inline_compares = len(re.findall(r"\.Type\s*==\s*Token\w+", stripped))
        ok = has_expect_call and inline_compares == 0
        return ok, f"p.expect used; inline compares={inline_compares}"

    if asid == "parser-test-empty-block":
        ok = ("empty_block" in parser_test or "empty block" in parser_test or
              re.search(r"block\s+\w+\s*\{\s*\}", parser_test) is not None)
        return ok, "empty-block test present" if ok else "no empty-block test"

    if asid == "parser-test-multi-record-block":
        # Look for a test source with two record statements inside a block
        ok = re.search(r"block\s+\w+\s*\{[^}]*record[^}]*record", parser_test, re.DOTALL) is not None
        return ok, "multi-record block test present" if ok else "no multi-record block test"

    if asid == "parser-test-via-parse":
        has_parse = "Parse(" in parser_test
        # Disallow direct construction of Block literal in expectations
        bad_block_literal = re.search(r"Block\s*\{[^}]*Records\s*:", parser_test) is not None
        ok = has_parse and not bad_block_literal
        return ok, f"parser tests use Parse() (has_parse={has_parse}, bad_literal={bad_block_literal})"

    if asid == "printer-round-trip-test":
        ok = ("RoundTrip" in printer_test or "round" in printer_test.lower()) \
             and "Parse" in printer_test and "Print" in printer_test \
             and "block" in printer_test.lower()
        return ok, "round-trip test covering block present" if ok else "no block round-trip"

    if asid == "tests-parallel":
        for name, txt in [("tokenizer_test.go", tokenizer_test), ("parser_test.go", parser_test), ("printer_test.go", printer_test)]:
            if txt.count("t.Parallel()") < 2:
                return False, f"{name} insufficient t.Parallel"
        return True, "t.Parallel at both levels"

    if asid == "go-test-passes":
        ok = go_test_log_passed(run_dir)
        return ok, "go test -race ./... passes" if ok else "go test failed"

    if asid == "no-full-spec-copy":
        leftovers = [p.name for p in pkg.glob("_spec*.md")]
        return len(leftovers) == 0, "no _spec*.md leftovers" if not leftovers else f"leftover: {leftovers}"

    if asid == "context-summary-spec-tightened":
        return assertion_context_summary_spec_tightened()

    if asid == "phase-chunking-spec-tightened":
        return assertion_phase_chunking_spec()

    return False, f"unknown assertion id: {asid}"


def assertion_eval2_comments(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    tokenizer = read(pkg / "tokenizer.go")
    parser = read(pkg / "parser.go")
    printer = read(pkg / "printer.go")
    tokenizer_test = read(pkg / "tokenizer_test.go")
    parser_test = read(pkg / "parser_test.go")
    printer_test = read(pkg / "printer_test.go")

    if asid == "token-comment-defined":
        ok = re.search(r"\bTokenComment\b", tokenizer) is not None
        return ok, "TokenComment constant defined" if ok else "no TokenComment in tokenizer.go"

    if asid == "tokenizer-handles-hash-comment":
        ok = ("'#'" in tokenizer or '"#"' in tokenizer) and "TokenComment" in tokenizer
        return ok, "tokenizer handles # comments" if ok else "no #-comment logic"

    if asid == "leading-comments-on-record":
        # Look for a LeadingComments []string field on Record (or equivalent
        # trivia field — name allowed to vary)
        candidates = ["LeadingComments", "Comments", "Trivia", "Leading", "DocComments"]
        # require Record struct to mention at least one
        record_block = re.search(r"type\s+Record\s+struct\s*\{[^}]*\}", parser, re.DOTALL)
        if not record_block:
            return False, "no Record struct"
        block = record_block.group(0)
        present = [c for c in candidates if c in block]
        ok = bool(present)
        return ok, f"Record carries {present}" if ok else "no trivia field on Record"

    if asid == "parser-attaches-comments":
        # Look for code that reads TokenComment and accumulates
        ok = "TokenComment" in parser and (
            "append(" in parser and ("comment" in parser.lower() or "leading" in parser.lower() or "trivia" in parser.lower())
        )
        return ok, "parser captures and attaches comments" if ok else "no comment-attachment logic"

    if asid == "printer-emits-leading-comments":
        ok = ("LeadingComments" in printer or "Comments" in printer or "Trivia" in printer) and (
            "#" in printer or "Comment" in printer
        )
        # Ensure there's an iteration over the comments
        has_iteration = re.search(r"for\s+[^{]*range\s+\w*\.?(Leading)?Comments?", printer) is not None
        ok = ok and (has_iteration or "Comment" in printer)
        return ok, "printer emits leading comments" if ok else "no comment-emission logic"

    if asid == "round-trip-single-comment":
        ok = "Parse" in printer_test and "Print" in printer_test and "#" in printer_test
        return ok, "round-trip with comment present" if ok else "no round-trip-with-comment test"

    if asid == "round-trip-multiple-comments":
        # Look for a test source that has two # lines in a row
        ok = re.search(r'(?:[\\\\]n|\\n)?\s*#[^"\n]*[\\\\]n\s*#', printer_test) is not None or \
             re.search(r'#[^"\n]*\n\s*#', printer_test) is not None or \
             ("two_comments" in printer_test or "multiple_comments" in printer_test or "two comments" in printer_test.lower())
        return ok, "multi-comment round-trip test present" if ok else "no multi-comment round-trip test"

    if asid == "tokenizer-test-comment-pos":
        has_pos = re.search(r"Pos\s*\{\s*Line\s*:\s*\d+\s*,\s*Column\s*:\s*\d+", tokenizer_test) is not None
        has_comment = "TokenComment" in tokenizer_test or "comment" in tokenizer_test.lower()
        ok = has_pos and has_comment
        return ok, f"comment Pos test (has_pos={has_pos}, has_comment={has_comment})"

    if asid == "tests-parallel":
        for name, txt in [("tokenizer_test.go", tokenizer_test), ("parser_test.go", parser_test), ("printer_test.go", printer_test)]:
            if txt.count("t.Parallel()") < 2:
                return False, f"{name} insufficient t.Parallel"
        return True, "t.Parallel at both levels"

    if asid == "go-test-passes":
        ok = go_test_log_passed(run_dir)
        return ok, "go test -race ./... passes" if ok else "go test failed"

    if asid == "no-full-spec-copy":
        leftovers = [p.name for p in pkg.glob("_spec*.md")]
        return len(leftovers) == 0, "no _spec*.md leftovers" if not leftovers else f"leftover: {leftovers}"

    if asid == "context-summary-spec-tightened":
        return assertion_context_summary_spec_tightened()

    if asid == "phase-chunking-spec-tightened":
        return assertion_phase_chunking_spec()

    return False, f"unknown assertion id: {asid}"


GRADERS = {
    0: assertion_eval0_string_record,
    1: assertion_eval1_block,
    2: assertion_eval2_comments,
}


def grade_run(eval_dir: Path, config: str) -> dict:
    meta = load_metadata(eval_dir)
    eval_id = meta["eval_id"]
    grader = GRADERS[eval_id]
    run_dir = eval_dir / config / "run-1"
    pkg = run_dir / "outputs" / "kvr"
    expectations = []
    for a in meta["assertions"]:
        passed, evidence = grader(a["id"], pkg, run_dir)
        expectations.append({
            "text": a["text"],
            "passed": bool(passed),
            "evidence": evidence,
        })
    n_passed = sum(1 for e in expectations if e["passed"])
    pass_rate = n_passed / len(expectations) if expectations else 0.0
    return {
        "eval_id": eval_id,
        "eval_name": meta["eval_name"],
        "config": config,
        "summary": {
            "passed": n_passed,
            "total": len(expectations),
            "pass_rate": pass_rate,
        },
        "expectations": expectations,
    }


def main():
    iter_dir = Path(sys.argv[1])
    for eval_dir in sorted(iter_dir.glob("eval-*")):
        for cfg in ("with_skill", "without_skill"):
            run_dir = eval_dir / cfg / "run-1"
            if not run_dir.exists():
                continue
            result = grade_run(eval_dir, cfg)
            (run_dir / "grading.json").write_text(json.dumps(result, indent=2))
            n_pass = sum(1 for e in result["expectations"] if e["passed"])
            total = len(result["expectations"])
            print(f"{eval_dir.name}/{cfg}: {n_pass}/{total} ({result['summary']['pass_rate']*100:.0f}%)")


if __name__ == "__main__":
    main()
