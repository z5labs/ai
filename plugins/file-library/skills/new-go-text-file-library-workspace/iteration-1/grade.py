#!/usr/bin/env python3
"""Grade scaffolded text-file-library packages against assertion sets."""
import json
import os
import re
import subprocess
import sys
from pathlib import Path

ITERATION = Path(__file__).parent

EVALS = [
    ("eval-0-scaffold-toml", "toml"),
    ("eval-1-scaffold-ini", "ini"),
    ("eval-2-scaffold-graphql", "graphql"),
]
CONFIGS = ["with_skill", "old_skill"]


def read(p: Path) -> str:
    try:
        return p.read_text()
    except Exception:
        return ""


def go_build_ok(verify_log: str) -> tuple[bool, str]:
    """`go build` (or compilation via go test) must succeed."""
    if not verify_log:
        return False, "verify.log missing"
    if "FAIL" in verify_log or "build failed" in verify_log.lower():
        return False, "FAIL/build failed in verify.log"
    if "ok " in verify_log and re.search(r"ok\s+\S+", verify_log):
        return True, "go test reports ok (compilation succeeded)"
    if "PASS" in verify_log and "go test" in verify_log:
        return True, "go test PASSed"
    return False, "no ok/PASS marker in verify.log"


def go_test_ok(verify_log: str) -> tuple[bool, str]:
    if not verify_log:
        return False, "verify.log missing"
    if "FAIL" in verify_log:
        return False, "FAIL in verify.log"
    if re.search(r"ok\s+\S+\s+\S+", verify_log) or "PASS" in verify_log:
        return True, "tests PASS / ok"
    return False, "no PASS marker"


def check(pkg_dir: Path, verify_log: str, pkg_name: str, aid: str, atext: str) -> tuple[bool, str]:
    tok = read(pkg_dir / "tokenizer.go")
    tok_t = read(pkg_dir / "tokenizer_test.go")
    par = read(pkg_dir / "parser.go")
    par_t = read(pkg_dir / "parser_test.go")
    pri = read(pkg_dir / "printer.go")
    pri_t = read(pkg_dir / "printer_test.go")
    cmd = read(pkg_dir / "CLAUDE.md")
    doc = read(pkg_dir / "doc.go")

    # File existence
    if aid == "files-doc-go":
        return doc != "", f"{pkg_name}/doc.go {'present' if doc else 'missing'}"
    if aid == "files-tokenizer-go":
        return tok != "", "tokenizer.go present" if tok else "missing"
    if aid == "files-tokenizer-test":
        return tok_t != "", "tokenizer_test.go present" if tok_t else "missing"
    if aid == "files-parser-go":
        return par != "", "parser.go present" if par else "missing"
    if aid == "files-parser-test":
        return par_t != "", "parser_test.go present" if par_t else "missing"
    if aid == "files-printer-go":
        return pri != "", "printer.go present" if pri else "missing"
    if aid == "files-printer-test":
        return pri_t != "", "printer_test.go present" if pri_t else "missing"
    if aid == "files-claude-md":
        return cmd != "", "CLAUDE.md present" if cmd else "missing"

    # Build/test
    if aid == "go-build-passes":
        return go_build_ok(verify_log)
    if aid == "go-test-passes":
        return go_test_ok(verify_log)

    # Function signatures (regex on whitespace-collapsed text)
    def has_pattern(text: str, pat: str) -> bool:
        collapsed = re.sub(r"\s+", " ", text)
        return re.search(pat, collapsed) is not None

    if aid == "tokenize-signature":
        ok = has_pattern(tok, r"func\s+Tokenize\s*\(\s*r\s+io\.Reader\s*\)\s+iter\.Seq2\[\s*Token\s*,\s*error\s*\]")
        return ok, "Tokenize signature found" if ok else "Tokenize signature not matched"
    if aid == "parse-signature":
        ok = has_pattern(par, r"func\s+Parse\s*\(\s*r\s+io\.Reader\s*\)\s+\(\s*\*File\s*,\s*error\s*\)")
        return ok, "Parse signature found" if ok else "Parse signature not matched"
    if aid == "print-signature":
        ok = has_pattern(pri, r"func\s+Print\s*\(\s*w\s+io\.Writer\s*,\s*f\s+\*File\s*\)\s+error")
        return ok, "Print signature found" if ok else "Print signature not matched"

    if aid == "tokenizer-action-type":
        ok = has_pattern(tok, r"type\s+tokenizerAction\s+func\s*\(\s*t\s+\*tokenizer\s*,\s*yield\s+func\s*\(\s*Token\s*,\s*error\s*\)\s+bool\s*\)\s+tokenizerAction")
        return ok, "tokenizerAction type matches" if ok else "tokenizerAction signature not matched"
    if aid == "parser-action-type":
        ok = has_pattern(par, r"type\s+parserAction\[\s*T\s+any\s*\]\s+func\s*\(\s*p\s+\*parser\s*,\s*t\s+T\s*\)\s+\(\s*parserAction\[T\]\s*,\s*error\s*\)")
        return ok, "parserAction[T any] type matches" if ok else "parserAction signature not matched"
    if aid == "printer-action-type":
        ok = has_pattern(pri, r"type\s+printerAction\s+func\s*\(\s*pr\s+\*printer\s*,\s*f\s+\*File\s*\)\s+printerAction")
        return ok, "printerAction type matches" if ok else "printerAction signature not matched"

    # Tests
    if aid == "tests-parallel":
        # The assertion requires t.Parallel() at BOTH function and subtest level
        # in every test file, so every file must contain at least 2 calls.
        files = list(zip(["tokenizer_test", "parser_test", "printer_test"], [tok_t, par_t, pri_t]))
        counts = [(label, t.count("t.Parallel()")) for label, t in files]
        deficient = [f"{label}.go has {c}" for label, c in counts if c < 2]
        if deficient:
            return False, f"t.Parallel() must appear ≥2× per test file (function + subtest); {', '.join(deficient)}"
        total = sum(c for _, c in counts)
        return True, f"t.Parallel() appears ≥2× in every test file (total {total})"

    if aid == "tests-testify-require":
        for label, t in zip(["tokenizer_test", "parser_test", "printer_test"], [tok_t, par_t, pri_t]):
            if "github.com/stretchr/testify/require" not in t:
                return False, f"{label}.go does not import testify/require"
        return True, "all three test files import testify/require"

    if aid == "parser-test-uses-parse":
        # The parser test must call Parse(...) — the public function — and not construct AST nodes by hand.
        calls_parse = bool(re.search(r"\bParse\s*\(", par_t))
        # Check for AST hand-construction smell: literal &File{ with non-empty body or building Type values directly
        # Lenient: if Parse() is called, accept; we only fail if Parse is missing.
        if calls_parse:
            return True, "parser_test.go calls Parse()"
        return False, "parser_test.go does not call Parse()"

    if aid == "printer-test-roundtrip":
        # Look for a TestPrinterRoundTrip function or a sequence of Parse(...) -> Print(...) -> Parse(...)
        if re.search(r"func\s+TestPrinterRoundTrip\b", pri_t):
            return True, "TestPrinterRoundTrip present"
        # fallback: presence of both Parse and Print calls in test file
        if re.search(r"\bParse\s*\(", pri_t) and re.search(r"\bPrint\s*\(", pri_t):
            return True, "Parse+Print round-trip pattern found"
        return False, "no round-trip test found"

    if aid == "claude-md-content":
        # CLAUDE.md must mention: action loop, inner action loop (or 'inner action' / nested), testing conventions
        c_lower = cmd.lower()
        action_loop = "action" in c_lower and ("loop" in c_lower or "state machine" in c_lower)
        inner = "inner action" in c_lower or "inner-action" in c_lower or ("complex" in c_lower and "action" in c_lower) or "no inline for" in c_lower or "nested" in c_lower and "action" in c_lower
        testing = ("t.parallel" in c_lower or "parallel" in c_lower) and ("require" in c_lower or "testify" in c_lower)
        missing = []
        if not action_loop:
            missing.append("action-loop pattern")
        if not inner:
            missing.append("inner-action-loop rule")
        if not testing:
            missing.append("testing conventions")
        if missing:
            return False, f"CLAUDE.md missing: {', '.join(missing)}"
        return True, "CLAUDE.md covers action-loop, inner-action rule, testing conventions"

    return False, f"unknown assertion id: {aid}"


def grade_run(eval_dir: Path, config: str, pkg_name: str) -> dict:
    metadata = json.loads((eval_dir / "eval_metadata.json").read_text())
    run_dir = eval_dir / config / "run-1"
    pkg_dir = run_dir / "outputs" / pkg_name
    verify_log = read(run_dir / "outputs" / "verify.log")

    expectations = []
    passed = 0
    failed = 0
    for assertion in metadata["assertions"]:
        aid = assertion["id"]
        atext = assertion["text"]
        ok, evidence = check(pkg_dir, verify_log, pkg_name, aid, atext)
        expectations.append({
            "id": aid,
            "text": atext,
            "passed": bool(ok),
            "evidence": evidence,
        })
        if ok:
            passed += 1
        else:
            failed += 1

    grading = {
        "summary": {
            "passed": passed,
            "failed": failed,
            "total": passed + failed,
            "pass_rate": round(passed / (passed + failed), 4) if (passed + failed) else 0.0,
        },
        "expectations": expectations,
    }
    (run_dir / "grading.json").write_text(json.dumps(grading, indent=2))
    return grading


def main():
    for eval_name, pkg_name in EVALS:
        eval_dir = ITERATION / eval_name
        for config in CONFIGS:
            grading = grade_run(eval_dir, config, pkg_name)
            print(f"{eval_name:25s} {config:10s} {grading['summary']['passed']}/{grading['summary']['total']} ({grading['summary']['pass_rate']*100:.0f}%)")


if __name__ == "__main__":
    main()
