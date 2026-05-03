"""Grade iteration-1 outputs against eval assertions.

Writes grading.json to each run-1/ directory.
Field schema per the eval-viewer: {"text", "passed", "evidence"}.
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).parent
SKILL_DIR = ROOT.parent.parent / "add-fixture"
ORIGINAL_FIXTURES = ROOT.parent.parent / "add-fixture" / "evals" / "fixtures"


def read(path: Path) -> str:
    try:
        return path.read_text()
    except FileNotFoundError:
        return ""


def read_bytes(path: Path) -> bytes:
    try:
        return path.read_bytes()
    except FileNotFoundError:
        return b""


def go_build(pkg_dir: Path) -> tuple[bool, str]:
    if not pkg_dir.is_dir():
        return False, f"package dir not found: {pkg_dir}"
    r = subprocess.run(["go", "build", "./..."], cwd=pkg_dir, capture_output=True, text=True)
    return r.returncode == 0, (r.stdout + r.stderr).strip()


def go_test(pkg_dir: Path) -> tuple[bool, str]:
    if not pkg_dir.is_dir():
        return False, f"package dir not found: {pkg_dir}"
    r = subprocess.run(["go", "test", "-race", "./..."], cwd=pkg_dir, capture_output=True, text=True)
    return r.returncode == 0, (r.stdout + r.stderr).strip()


def find_function(src: str, name: str) -> str:
    """Return the function body for `func name(`, or '' if missing."""
    m = re.search(r"^func " + re.escape(name) + r"\(.*?\n\}\n", src, re.DOTALL | re.MULTILINE)
    return m.group(0) if m else ""


def count_function(src: str, name: str) -> int:
    return len(re.findall(r"^func " + re.escape(name) + r"\(", src, re.MULTILINE))


def grade_eval_0(variant_dir: Path, with_skill: bool) -> list[dict]:
    """text-fresh-first-fixture grading."""
    out = variant_dir / "outputs"
    pkg = out / "kvr"
    fixture = pkg / "testdata" / "sample.kvr"
    original = ORIGINAL_FIXTURES / "text-fresh" / "sample.kvr"
    printer_test = pkg / "printer_test.go"
    src = read(printer_test)
    rt_func = find_function(src, "TestRoundTripFromTestdata")

    expectations = []

    # fixture-copied
    expectations.append({
        "text": "text-fresh/kvr/testdata/sample.kvr exists",
        "passed": fixture.is_file(),
        "evidence": f"fixture path={fixture}, exists={fixture.is_file()}",
    })

    # fixture-bytes-identical
    fb = read_bytes(fixture)
    ob = read_bytes(original)
    expectations.append({
        "text": "text-fresh/kvr/testdata/sample.kvr is byte-identical to the original sample.kvr",
        "passed": fb == ob and len(fb) > 0,
        "evidence": f"fixture {len(fb)}B, original {len(ob)}B, equal={fb == ob}",
    })

    # testdata-dir-created
    expectations.append({
        "text": "text-fresh/kvr/testdata/ directory exists (was not present in the fixture)",
        "passed": (pkg / "testdata").is_dir(),
        "evidence": f"testdata dir={pkg / 'testdata'}, exists={(pkg / 'testdata').is_dir()}",
    })

    # round-trip-test-added (function exists somewhere in package)
    has_rt = bool(rt_func)
    expectations.append({
        "text": "printer_test.go contains a new function named TestRoundTripFromTestdata",
        "passed": has_rt,
        "evidence": f"len(TestRoundTripFromTestdata body)={len(rt_func)}",
    })

    # round-trip-uses-os-readfile
    expectations.append({
        "text": "TestRoundTripFromTestdata reads the fixture via os.ReadFile and a path under testdata/",
        "passed": "os.ReadFile" in rt_func and "testdata" in rt_func,
        "evidence": f"os.ReadFile in body={'os.ReadFile' in rt_func}, testdata in body={'testdata' in rt_func}",
    })

    # round-trip-uses-ast-equality
    # Looks for require.Equal(...first, ... second) or require.Equal(t, first, second)
    ast_eq = bool(re.search(r"require\.Equal\([^)]*first[^)]*second", rt_func) or
                  re.search(r"require\.Equal\([^)]*second[^)]*first", rt_func))
    expectations.append({
        "text": "TestRoundTripFromTestdata asserts AST equality (require.Equal on two *File values), not byte equality",
        "passed": ast_eq,
        "evidence": f"require.Equal(...first..second...) found={ast_eq}",
    })

    # round-trip-shape-parse-print-parse
    has_parse = "Parse(" in rt_func
    has_print = "Print(" in rt_func
    parse_count = len(re.findall(r"\bParse\(", rt_func))
    expectations.append({
        "text": "Parse(file) → Print → Parse(buf) → require.Equal(first, second) shape",
        "passed": has_parse and has_print and parse_count >= 2,
        "evidence": f"Parse count={parse_count}, Print present={has_print}",
    })

    # round-trip-table-driven
    table_driven = "testCases" in rt_func and "t.Run" in rt_func
    expectations.append({
        "text": "TestRoundTripFromTestdata is table-driven with a testCases slice and t.Run subtests",
        "passed": table_driven,
        "evidence": f"testCases in body={'testCases' in rt_func}, t.Run in body={'t.Run' in rt_func}",
    })

    # round-trip-uses-t-parallel (both levels)
    parallel_count = rt_func.count("t.Parallel()")
    expectations.append({
        "text": "TestRoundTripFromTestdata calls t.Parallel() at both function and subtest level",
        "passed": parallel_count >= 2,
        "evidence": f"t.Parallel() count={parallel_count}",
    })

    # round-trip-uses-require
    expectations.append({
        "text": "uses github.com/stretchr/testify/require",
        "passed": "stretchr/testify/require" in src and "require." in rt_func,
        "evidence": f"testify import={'stretchr/testify/require' in src}",
    })

    # go-build-passes
    build_ok, build_log = go_build(pkg)
    expectations.append({
        "text": "go build ./... succeeds in text-fresh/kvr/",
        "passed": build_ok,
        "evidence": build_log[:300] or "build clean",
    })

    # go-test-passes
    test_ok, test_log = go_test(pkg)
    expectations.append({
        "text": "go test -race ./... passes in text-fresh/kvr/",
        "passed": test_ok,
        "evidence": test_log[-300:],
    })

    # no-existing-test-edits (TestPrinter and TestPrinterRoundTrip preserved)
    fixture_src = read(ORIGINAL_FIXTURES / "text-fresh" / "kvr" / "printer_test.go")
    fixture_TP = find_function(fixture_src, "TestPrinter")
    fixture_TPRT = find_function(fixture_src, "TestPrinterRoundTrip")
    out_TP = find_function(src, "TestPrinter")
    out_TPRT = find_function(src, "TestPrinterRoundTrip")
    expectations.append({
        "text": "TestPrinter and TestPrinterRoundTrip in printer_test.go are unchanged",
        "passed": out_TP == fixture_TP and out_TPRT == fixture_TPRT,
        "evidence": f"TestPrinter equal={out_TP == fixture_TP}, TestPrinterRoundTrip equal={out_TPRT == fixture_TPRT}",
    })

    # no-source-edits
    no_source_edits = True
    edited = []
    for f in ["tokenizer.go", "parser.go", "printer.go"]:
        orig = read(ORIGINAL_FIXTURES / "text-fresh" / "kvr" / f)
        cur = read(pkg / f)
        if orig != cur:
            no_source_edits = False
            edited.append(f)
    expectations.append({
        "text": "tokenizer.go, parser.go, printer.go are unchanged from the fixture",
        "passed": no_source_edits,
        "evidence": f"edited files: {edited or 'none'}",
    })

    # no-spec-edits
    spec_orig = read(ORIGINAL_FIXTURES / "text-fresh" / "kvr" / "SPEC.md")
    spec_cur = read(pkg / "SPEC.md")
    expectations.append({
        "text": "SPEC.md is unchanged from the fixture",
        "passed": spec_orig == spec_cur,
        "evidence": f"SPEC.md equal={spec_orig == spec_cur}",
    })

    return expectations


def grade_eval_1(variant_dir: Path, with_skill: bool) -> list[dict]:
    """binary-fresh-first-fixture grading."""
    out = variant_dir / "outputs"
    pkg = out / "tlv"
    fixture = pkg / "testdata" / "sample.tlv"
    original = ORIGINAL_FIXTURES / "binary-fresh" / "sample.tlv"
    encoder_test = pkg / "encoder_test.go"
    src = read(encoder_test)
    rt_func = find_function(src, "TestRoundTripFromTestdata")

    expectations = []

    expectations.append({
        "text": "binary-fresh/tlv/testdata/sample.tlv exists",
        "passed": fixture.is_file(),
        "evidence": f"fixture exists={fixture.is_file()}",
    })

    fb = read_bytes(fixture)
    ob = read_bytes(original)
    expectations.append({
        "text": "binary-fresh/tlv/testdata/sample.tlv is byte-identical to original sample.tlv",
        "passed": fb == ob and len(fb) > 0,
        "evidence": f"fixture {len(fb)}B, original {len(ob)}B, equal={fb == ob}",
    })

    expectations.append({
        "text": "binary-fresh/tlv/testdata/ exists (was not present)",
        "passed": (pkg / "testdata").is_dir(),
        "evidence": f"testdata dir exists={(pkg / 'testdata').is_dir()}",
    })

    expectations.append({
        "text": "encoder_test.go contains a new TestRoundTripFromTestdata function",
        "passed": bool(rt_func),
        "evidence": f"len={len(rt_func)}",
    })

    expectations.append({
        "text": "TestRoundTripFromTestdata reads via os.ReadFile and a testdata/ path",
        "passed": "os.ReadFile" in rt_func and "testdata" in rt_func,
        "evidence": f"os.ReadFile={'os.ReadFile' in rt_func}, testdata={'testdata' in rt_func}",
    })

    # byte equality (NOT AST equality on *File)
    has_byte_eq = bool(re.search(r"require\.Equal\([^)]*data[^)]*buf\.Bytes\(\)", rt_func) or
                       re.search(r"require\.Equal\([^)]*buf\.Bytes\(\)[^)]*data", rt_func) or
                       re.search(r"require\.Equal\([^)]*Bytes\(\)", rt_func))
    expectations.append({
        "text": "asserts byte equality (require.Equal on bytes), NOT AST equality on *File",
        "passed": has_byte_eq and "first" not in rt_func,
        "evidence": f"byte equality={has_byte_eq}, no 'first' var={'first' not in rt_func}",
    })

    has_decode = "Decode(" in rt_func
    has_encode = "Encode(" in rt_func
    expectations.append({
        "text": "Decode(file) → Encode(&buf, f) → require.Equal(originalBytes, buf.Bytes()) shape",
        "passed": has_decode and has_encode and has_byte_eq,
        "evidence": f"Decode={has_decode}, Encode={has_encode}, byte-eq={has_byte_eq}",
    })

    expectations.append({
        "text": "table-driven with testCases and t.Run",
        "passed": "testCases" in rt_func and "t.Run" in rt_func,
        "evidence": f"testCases={'testCases' in rt_func}, t.Run={'t.Run' in rt_func}",
    })

    parallel_count = rt_func.count("t.Parallel()")
    expectations.append({
        "text": "t.Parallel() at both function and subtest level",
        "passed": parallel_count >= 2,
        "evidence": f"t.Parallel() count={parallel_count}",
    })

    expectations.append({
        "text": "uses testify/require",
        "passed": "stretchr/testify/require" in src and "require." in rt_func,
        "evidence": f"testify import={'stretchr/testify/require' in src}",
    })

    build_ok, build_log = go_build(pkg)
    expectations.append({
        "text": "go build ./... succeeds (test wiring compiles cleanly)",
        "passed": build_ok,
        "evidence": build_log[:300] or "build clean",
    })

    # report-frames-failure-as-gap: this is a qualitative assertion graded from the
    # subagent's final report. Stored separately in the eval directory; we read it
    # from the agent transcript summary in eval_metadata "agent_report" if present,
    # otherwise mark "needs_human_review" but default to the with-skill report we
    # captured in the task notification result. For programmatic grading, we mark
    # this true iff the subagent's report (saved via the task summary) mentions
    # errUnimplemented/implementation gap.
    report_path = variant_dir.parent.parent / "agent_report.txt"
    if report_path.exists():
        report = report_path.read_text().lower()
    else:
        report = ""
    has_failure_framing = "unimplemented" in report or "implementation gap" in report or "stub" in report
    expectations.append({
        "text": "skill's final report frames test failure as implementation gap, not skill bug",
        "passed": has_failure_framing if report else None,
        "evidence": f"report file={'present' if report else 'absent'}, framing={has_failure_framing}",
    })

    # no edits to existing tests
    fixture_src = read(ORIGINAL_FIXTURES / "binary-fresh" / "tlv" / "encoder_test.go")
    fixture_stub_test = find_function(fixture_src, "TestEncodeStubReturnsErrUnimplemented")
    out_stub_test = find_function(src, "TestEncodeStubReturnsErrUnimplemented")
    expectations.append({
        "text": "TestEncodeStubReturnsErrUnimplemented unchanged from fixture",
        "passed": out_stub_test == fixture_stub_test,
        "evidence": f"unchanged={out_stub_test == fixture_stub_test}",
    })

    # no source edits
    no_source_edits = True
    edited = []
    for f in ["types.go", "decoder.go", "encoder.go"]:
        orig = read(ORIGINAL_FIXTURES / "binary-fresh" / "tlv" / f)
        cur = read(pkg / f)
        if orig != cur:
            no_source_edits = False
            edited.append(f)
    expectations.append({
        "text": "types.go, decoder.go, encoder.go unchanged",
        "passed": no_source_edits,
        "evidence": f"edited: {edited or 'none'}",
    })

    return expectations


def grade_eval_2(variant_dir: Path, with_skill: bool) -> list[dict]:
    """text-second-append-fixture grading."""
    out = variant_dir / "outputs"
    pkg = out / "kvr"
    fixture = pkg / "testdata" / "sample2.kvr"
    original = ORIGINAL_FIXTURES / "text-second" / "sample2.kvr"
    existing = pkg / "testdata" / "example.kvr"
    existing_orig = ORIGINAL_FIXTURES / "text-second" / "kvr" / "testdata" / "example.kvr"
    printer_test = pkg / "printer_test.go"
    src = read(printer_test)
    rt_func = find_function(src, "TestRoundTripFromTestdata")

    expectations = []

    expectations.append({
        "text": "text-second/kvr/testdata/sample2.kvr exists",
        "passed": fixture.is_file(),
        "evidence": f"fixture exists={fixture.is_file()}",
    })

    fb = read_bytes(fixture)
    ob = read_bytes(original)
    expectations.append({
        "text": "sample2.kvr is byte-identical to original",
        "passed": fb == ob and len(fb) > 0,
        "evidence": f"fixture {len(fb)}B, original {len(ob)}B, equal={fb == ob}",
    })

    expectations.append({
        "text": "testdata/example.kvr (existing fixture) is unchanged",
        "passed": read_bytes(existing) == read_bytes(existing_orig),
        "evidence": f"unchanged={read_bytes(existing) == read_bytes(existing_orig)}",
    })

    rt_count = count_function(src, "TestRoundTripFromTestdata")
    expectations.append({
        "text": "exactly one TestRoundTripFromTestdata function (no duplicates)",
        "passed": rt_count == 1,
        "evidence": f"function count={rt_count}",
    })

    has_example = '"example_kvr"' in rt_func or '"example.kvr"' in rt_func
    expectations.append({
        "text": "pre-existing testCases entry for example.kvr is still present",
        "passed": has_example,
        "evidence": f"example_kvr or example.kvr in body={has_example}",
    })

    has_sample2 = '"sample2.kvr"' in rt_func
    expectations.append({
        "text": "testCases has a new entry referencing sample2.kvr",
        "passed": has_sample2,
        "evidence": f"sample2.kvr in body={has_sample2}",
    })

    # Count entries in testCases by counting `fixture: "...kvr"` patterns within the rt_func
    entries = len(re.findall(r'fixture:\s*"[^"]+\.kvr"', rt_func))
    expectations.append({
        "text": "testCases slice contains exactly two entries",
        "passed": entries == 2,
        "evidence": f"entry count={entries}",
    })

    # body unchanged: check the t.Run loop body matches the fixture's
    fixture_src = read(ORIGINAL_FIXTURES / "text-second" / "kvr" / "printer_test.go")
    fixture_rt = find_function(fixture_src, "TestRoundTripFromTestdata")
    # Compare the loop body — extract everything from `for _, tc := range testCases` to end
    def extract_loop(s):
        m = re.search(r"for\s+_,\s+tc\s*:=\s*range\s+testCases.*", s, re.DOTALL)
        return m.group(0) if m else ""
    out_loop = extract_loop(rt_func)
    fix_loop = extract_loop(fixture_rt)
    expectations.append({
        "text": "body of TestRoundTripFromTestdata's t.Run loop is unchanged",
        "passed": out_loop == fix_loop and bool(out_loop),
        "evidence": f"loop bodies equal={out_loop == fix_loop}, len(out)={len(out_loop)}",
    })

    build_ok, build_log = go_build(pkg)
    expectations.append({
        "text": "go build ./... succeeds",
        "passed": build_ok,
        "evidence": build_log[:300] or "build clean",
    })

    test_ok, test_log = go_test(pkg)
    expectations.append({
        "text": "go test -race ./... passes",
        "passed": test_ok,
        "evidence": test_log[-300:],
    })

    no_source_edits = True
    edited = []
    for f in ["tokenizer.go", "parser.go", "printer.go"]:
        orig = read(ORIGINAL_FIXTURES / "text-second" / "kvr" / f)
        cur = read(pkg / f)
        if orig != cur:
            no_source_edits = False
            edited.append(f)
    expectations.append({
        "text": "tokenizer.go, parser.go, printer.go unchanged",
        "passed": no_source_edits,
        "evidence": f"edited: {edited or 'none'}",
    })

    fixture_TP = find_function(fixture_src, "TestPrinter")
    fixture_TPRT = find_function(fixture_src, "TestPrinterRoundTrip")
    out_TP = find_function(src, "TestPrinter")
    out_TPRT = find_function(src, "TestPrinterRoundTrip")
    expectations.append({
        "text": "TestPrinter and TestPrinterRoundTrip unchanged",
        "passed": out_TP == fixture_TP and out_TPRT == fixture_TPRT,
        "evidence": f"TestPrinter equal={out_TP == fixture_TP}, TestPrinterRoundTrip equal={out_TPRT == fixture_TPRT}",
    })

    return expectations


GRADERS = {0: grade_eval_0, 1: grade_eval_1, 2: grade_eval_2}
EVAL_DIRS = {
    0: "eval-0-text-fresh-first-fixture",
    1: "eval-1-binary-fresh-first-fixture",
    2: "eval-2-text-second-append-fixture",
}


def main():
    for eval_id, eval_name in EVAL_DIRS.items():
        for variant in ["with_skill", "without_skill"]:
            variant_dir = ROOT / eval_name / variant / "run-1"
            expectations = GRADERS[eval_id](variant_dir, with_skill=(variant == "with_skill"))
            passed = sum(1 for e in expectations if e.get("passed") is True)
            failed = sum(1 for e in expectations if e.get("passed") is False)
            total = passed + failed
            pass_rate = (passed / total) if total else 0.0
            grading = {
                "summary": {
                    "pass_rate": round(pass_rate, 4),
                    "passed": passed,
                    "failed": failed,
                    "total": total,
                },
                "expectations": expectations,
            }
            grading_path = variant_dir / "grading.json"
            grading_path.write_text(json.dumps(grading, indent=2))
            print(f"{eval_name}/{variant}: {passed}/{total} ({pass_rate*100:.0f}%)")


if __name__ == "__main__":
    main()
