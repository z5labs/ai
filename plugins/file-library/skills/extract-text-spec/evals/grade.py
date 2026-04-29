#!/usr/bin/env python3
"""Grade extract-text-spec eval outputs against assertion lists.

Walks an iteration directory, finds every (eval-*-name)/(with_skill|without_skill)/outputs/
tree, and writes grading.json next to it.

Usage:
    python grade.py <iteration-dir>
"""
from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path


def find_spec_md(outputs_dir: Path) -> Path | None:
    candidates = list(outputs_dir.glob("*/SPEC.md"))
    if candidates:
        return candidates[0]
    candidates = list(outputs_dir.rglob("SPEC.md"))
    return candidates[0] if candidates else None


def find_format_dir(outputs_dir: Path) -> Path | None:
    for child in outputs_dir.iterdir():
        if child.is_dir():
            return child
    return None


def read(p: Path) -> str:
    try:
        return p.read_text(encoding="utf-8", errors="replace")
    except Exception:
        return ""


SECTION_HEADINGS = [
    "## Overview",
    "## Lexical Elements (Tokens)",
    "## Structure (Grammar)",
    "## Semantics",
    "## Examples",
    "## Appendix",
]


def section_present(text: str, heading: str) -> bool:
    pattern = rf"^{re.escape(heading)}\s*$"
    return bool(re.search(pattern, text, re.MULTILINE))


def section_order_correct(text: str) -> tuple[bool, str]:
    positions: list[tuple[str, int]] = []
    for h in SECTION_HEADINGS:
        m = re.search(rf"^{re.escape(h)}\s*$", text, re.MULTILINE)
        if m:
            positions.append((h, m.start()))
    if len(positions) < len(SECTION_HEADINGS):
        missing = [h for h in SECTION_HEADINGS if h not in [p[0] for p in positions]]
        return False, f"missing sections: {missing}"
    sorted_positions = sorted(positions, key=lambda x: x[1])
    if [p[0] for p in sorted_positions] != SECTION_HEADINGS:
        return False, f"order observed: {[p[0] for p in sorted_positions]}"
    return True, "all six sections in correct order"


def extract_section(text: str, heading: str) -> str:
    """Return text from `heading` up to the next ## heading."""
    m = re.search(rf"^{re.escape(heading)}\s*$", text, re.MULTILINE)
    if not m:
        return ""
    start = m.end()
    next_m = re.search(r"^##\s+\S", text[start:], re.MULTILINE)
    end = start + next_m.start() if next_m else len(text)
    return text[start:end]


def extract_all_examples_blocks(examples_section: str) -> list[str]:
    """Extract ```...``` fenced blocks from examples section."""
    blocks = re.findall(r"```[a-zA-Z]*\n(.*?)```", examples_section, re.DOTALL)
    return [b.strip() for b in blocks if b.strip()]


def count_h3_under_examples(examples_section: str) -> int:
    return len(re.findall(r"^### ", examples_section, re.MULTILINE))


def grade_assertion(aid: str, text: str, outputs_dir: Path, format_dir: Path | None,
                    spec_path: Path | None) -> tuple[bool, str]:
    """Return (passed, evidence)."""
    spec_text = read(spec_path) if spec_path else ""

    # ---- layout assertions ----
    if aid == "spec-md-exists":
        if spec_path and spec_path.exists():
            return True, f"found at {spec_path.relative_to(outputs_dir)}"
        return False, "no SPEC.md found anywhere under outputs/"

    if aid == "spec-no-empty":
        # Threshold differs per eval; check from assertion text.
        m = re.search(r"at least (\d+) bytes", text)
        threshold = int(m.group(1)) if m else 1000
        size = spec_path.stat().st_size if spec_path else 0
        return size >= threshold, f"SPEC.md is {size} bytes (threshold {threshold})"

    if aid == "no-scratch-leftovers":
        if not format_dir:
            return False, "no format dir present"
        leftovers = []
        for pattern in ["_spec.txt", "_spec.html", "_toc.md", "_sections_index.md"]:
            leftovers += [str(p.relative_to(outputs_dir)) for p in format_dir.rglob(pattern)]
        leftovers += [str(p.relative_to(outputs_dir)) for p in format_dir.rglob("_spec_part_*.md")]
        return (len(leftovers) == 0), ("none" if not leftovers else f"found: {leftovers}")

    # ---- section presence ----
    section_map = {
        "section-overview": "## Overview",
        "section-lexical": "## Lexical Elements (Tokens)",
        "section-grammar": "## Structure (Grammar)",
        "section-semantics": "## Semantics",
        "section-examples": "## Examples",
        "section-appendix": "## Appendix",
    }
    if aid in section_map:
        h = section_map[aid]
        ok = section_present(spec_text, h)
        return ok, f"`{h}` heading {'found' if ok else 'NOT found'}"

    if aid == "section-order":
        return section_order_correct(spec_text)

    # ---- JSON-specific ----
    if aid == "json-string-escapes":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        # Look for each escape sequence (allowing backtick wrapping or plain)
        escapes = [r'\\"', r"\\\\", r"\\/", r"\\b", r"\\f", r"\\n", r"\\r", r"\\t", r"\\uXXXX"]
        readable = ['\\"', '\\\\', '\\/', '\\b', '\\f', '\\n', '\\r', '\\t', '\\uXXXX']
        # Match each in raw form (not regex-escaped, look for the literal characters)
        missing = []
        # Build literal patterns
        for esc, label in zip(['\\"', '\\\\', '\\/', '\\b', '\\f', '\\n', '\\r', '\\t'], readable):
            if esc not in lex and esc.replace("\\", "\\\\") not in lex:
                missing.append(label)
        # uXXXX
        if not re.search(r'\\u[Xx]{4}|\\u[0-9a-fA-F]{4}', lex):
            missing.append("\\uXXXX")
        return (len(missing) == 0), (f"all escapes present" if not missing else f"missing: {missing}")

    if aid == "json-number-format":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        checks = {
            "minus": bool(re.search(r"\bminus\b|\b-\b|leading\s+-", lex, re.IGNORECASE)),
            "integer/int": bool(re.search(r"\bint\b|integer", lex, re.IGNORECASE)),
            "fraction/decimal": bool(re.search(r"fractio|decimal|\bfrac\b", lex, re.IGNORECASE)),
            "exponent": bool(re.search(r"exponent|[eE]\s*[+-]?", lex, re.IGNORECASE)),
        }
        missing = [k for k, v in checks.items() if not v]
        return (len(missing) == 0), (f"all components present" if not missing else f"missing: {missing}")

    if aid == "json-keywords":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        ok = all(k in lex for k in ["true", "false", "null"])
        return ok, "all three documented" if ok else "one of true/false/null missing"

    if aid == "json-structural-tokens":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        symbols = ["{", "}", "[", "]", ":", ","]
        missing = [s for s in symbols if s not in lex]
        return (len(missing) == 0), "all six structural tokens present" if not missing else f"missing: {missing}"

    if aid == "json-grammar-productions":
        gram = extract_section(spec_text, "## Structure (Grammar)")
        # Look for object/array/member or value/element with = or ::= or /
        has_object = bool(re.search(r"\b(object|JSON-object|begin-object)\b", gram, re.IGNORECASE))
        has_array = bool(re.search(r"\b(array|JSON-array|begin-array)\b", gram, re.IGNORECASE))
        has_member = bool(re.search(r"\b(member|members|key-value|name-separator)\b", gram, re.IGNORECASE))
        has_grammar_op = bool(re.search(r"=|::=|→|->|/", gram))
        ok = has_object and has_array and has_member and has_grammar_op
        return ok, f"object={has_object} array={has_array} member={has_member} ebnf-op={has_grammar_op}"

    # ---- examples ----
    if aid == "examples-three-or-more":
        ex = extract_section(spec_text, "## Examples")
        h3 = count_h3_under_examples(ex)
        blocks = extract_all_examples_blocks(ex)
        # Use whichever metric is larger; assert at least 3
        n = max(h3, len(blocks))
        return n >= 3, f"### subsections={h3}, fenced blocks={len(blocks)}"

    if aid == "examples-parseable":
        ex = extract_section(spec_text, "## Examples")
        blocks = extract_all_examples_blocks(ex)
        if not blocks:
            return False, "no fenced examples found"
        bad = []
        for i, b in enumerate(blocks):
            try:
                json.loads(b)
            except Exception as e:
                bad.append(f"block {i+1}: {type(e).__name__}: {str(e)[:60]}")
        return (len(bad) == 0), ("all parse" if not bad else f"failures: {bad}")

    if aid == "json-out-of-scope-ijson-skipped":
        # Heuristic: I-JSON / security considerations should not appear as their own ## sections
        bad_terms = [
            ("I-JSON", "I-JSON"),
            ("Security Considerations as own section",
             "## Security Considerations" if re.search(r"^## Security Considerations", spec_text, re.MULTILINE) else None),
        ]
        observed = []
        if "I-JSON" in spec_text:
            observed.append("I-JSON mentioned")
        if re.search(r"^## Security Considerations", spec_text, re.MULTILINE):
            observed.append("Security Considerations as ##")
        return (len(observed) == 0), "neither present" if not observed else f"found: {observed}"

    # ---- TOML-specific ----
    if aid == "toml-comments-documented":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        ok = "#" in lex and re.search(r"comment", lex, re.IGNORECASE)
        return bool(ok), "# comments documented" if ok else "missing"

    if aid == "toml-string-variants":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        variants = {
            "basic (\"...\")": '"' in lex and re.search(r"basic", lex, re.IGNORECASE),
            "multi-line basic (\"\"\"...\"\"\")": '"""' in lex,
            "literal ('...')": "'" in lex and re.search(r"literal", lex, re.IGNORECASE),
            "multi-line literal ('''...''')": "'''" in lex,
        }
        missing = [k for k, v in variants.items() if not v]
        return (len(missing) == 0), "all four variants present" if not missing else f"missing: {missing}"

    if aid == "toml-integer-bases":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        bases = {"0x": "0x" in lex, "0o": "0o" in lex, "0b": "0b" in lex,
                 "underscore separators": "_" in lex and re.search(r"under", lex, re.IGNORECASE)}
        missing = [k for k, v in bases.items() if not v]
        return (len(missing) == 0), "all bases" if not missing else f"missing: {missing}"

    if aid == "toml-float-special":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        ok = bool(re.search(r"\binf\b", lex, re.IGNORECASE)) and bool(re.search(r"\bnan\b", lex, re.IGNORECASE))
        return ok, "inf/nan present" if ok else "missing inf or nan"

    if aid == "toml-datetime-four-variants":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        variants = {
            "Offset Date-Time": bool(re.search(r"offset[\s-]date[\s-]time", lex, re.IGNORECASE)),
            "Local Date-Time": bool(re.search(r"local[\s-]date[\s-]time", lex, re.IGNORECASE)),
            "Local Date": bool(re.search(r"local[\s-]date(?![\s-]time)", lex, re.IGNORECASE)),
            "Local Time": bool(re.search(r"local[\s-]time", lex, re.IGNORECASE)),
        }
        missing = [k for k, v in variants.items() if not v]
        return (len(missing) == 0), "all four datetime variants" if not missing else f"missing: {missing}"

    if aid == "toml-table-vs-inline":
        gram = extract_section(spec_text, "## Structure (Grammar)") + extract_section(spec_text, "## Lexical Elements (Tokens)")
        ok = re.search(r"\[[a-zA-Z]", gram) and re.search(r"inline[\s-]table", gram, re.IGNORECASE)
        return bool(ok), "both forms documented" if ok else "missing one form"

    if aid == "toml-array-of-tables":
        gram = extract_section(spec_text, "## Structure (Grammar)") + extract_section(spec_text, "## Lexical Elements (Tokens)")
        ok = "[[" in gram or re.search(r"array[\s-]of[\s-]tables?", gram, re.IGNORECASE)
        return bool(ok), "documented" if ok else "missing"

    if aid == "toml-dotted-keys":
        body = spec_text
        ok = re.search(r"dotted[\s-]key", body, re.IGNORECASE)
        return bool(ok), "documented" if ok else "missing"

    if aid == "examples-exercise-multiple-types":
        ex = extract_section(spec_text, "## Examples")
        blocks = extract_all_examples_blocks(ex)
        if not blocks:
            return False, "no fenced examples"
        complex_block = max(blocks, key=len)
        features = {
            "string": '"' in complex_block or "'" in complex_block,
            "integer": bool(re.search(r"=\s*-?\d+(?!\.\d)", complex_block)),
            "float": bool(re.search(r"=\s*-?\d+\.\d", complex_block)),
            "boolean": bool(re.search(r"=\s*(true|false)\b", complex_block)),
            "datetime": bool(re.search(r"\d{4}-\d{2}-\d{2}", complex_block)),
            "array": "[" in complex_block,
            "table": "[" in complex_block and "]" in complex_block,
        }
        missing = [k for k, v in features.items() if not v]
        return (len(missing) == 0), "all features in complex example" if not missing else f"missing in complex example: {missing}"

    if aid == "ambiguity-callouts-or-explicit-resolutions":
        # TOML spec is well-defined; either zero callouts or each has a resolution recommendation.
        callouts = re.findall(r">\s*\*\*Ambiguity:\*\*([^\n]*(?:\n>[^\n]*)*)", spec_text)
        if len(callouts) == 0:
            return True, "zero ambiguity callouts (spec is well-defined)"
        # Each callout should mention a recommendation/resolution/decision keyword.
        unresolved = []
        for c in callouts:
            if not re.search(r"recommend|resolve|decide|treat as|implement as|use|choose", c, re.IGNORECASE):
                unresolved.append(c[:80])
        return (len(unresolved) == 0), f"{len(callouts)} callouts, unresolved: {len(unresolved)}"

    # ---- gitconfig-specific ----
    if aid == "gitconfig-section-header-quoted-subsection":
        ok = bool(re.search(r'\[\w+\s+"[^"]+"\]', spec_text))
        return ok, "quoted subsection example present" if ok else "missing"

    if aid == "gitconfig-comments-both":
        lex = extract_section(spec_text, "## Lexical Elements (Tokens)")
        # both # and ; mentioned in proximity to "comment"
        ok = "#" in lex and ";" in lex and re.search(r"comment", lex, re.IGNORECASE)
        return bool(ok), "both comment introducers present" if ok else "missing one"

    if aid == "gitconfig-line-continuation":
        ok = bool(re.search(r"line[\s-]continuation|trailing\s+backslash|\\\s*\n", spec_text, re.IGNORECASE))
        return ok, "documented" if ok else "missing"

    if aid == "gitconfig-value-types":
        body = spec_text
        types = {
            "boolean": bool(re.search(r"\bboolean\b", body, re.IGNORECASE)),
            "integer with k/M/G": bool(re.search(r"\bk\b|\bm\b|\bg\b.*suffix|kilo|mega|giga", body, re.IGNORECASE)),
            "color or path": bool(re.search(r"\bcolor\b|\bpath\b", body, re.IGNORECASE)),
        }
        missing = [k for k, v in types.items() if not v]
        return (len(missing) == 0), "all listed" if not missing else f"missing: {missing}"

    if aid == "gitconfig-case-rules":
        body = spec_text
        ok = bool(re.search(r"case[\s-]insensitive|case[\s-]sensitiv", body, re.IGNORECASE))
        return ok, "case rules documented" if ok else "missing"

    if aid == "gitconfig-include-mechanism":
        ok = bool(re.search(r"\binclude\b", spec_text, re.IGNORECASE)) and bool(re.search(r"includeIf", spec_text))
        return ok, "include + includeIf documented" if ok else "missing"

    if aid == "ambiguity-callout-present":
        callouts = re.findall(r">\s*\*\*Ambiguity:\*\*", spec_text)
        return len(callouts) >= 1, f"{len(callouts)} ambiguity callouts"

    return False, f"no grader implemented for {aid}"


def grade_run(run_dir: Path, eval_metadata: dict) -> dict:
    outputs = run_dir / "outputs"
    if not outputs.exists():
        return {
            "run_id": str(run_dir.name),
            "expectations": [
                {"text": a["text"], "passed": False, "evidence": "outputs/ missing"}
                for a in eval_metadata.get("assertions", [])
            ],
        }
    spec_path = find_spec_md(outputs)
    format_dir = find_format_dir(outputs)

    expectations = []
    for a in eval_metadata.get("assertions", []):
        try:
            passed, evidence = grade_assertion(a["id"], a["text"], outputs, format_dir, spec_path)
        except Exception as e:
            passed, evidence = False, f"grader error: {type(e).__name__}: {e}"
        expectations.append({
            "text": a["text"],
            "passed": passed,
            "evidence": evidence,
        })
    passed_count = sum(1 for e in expectations if e["passed"])
    total = len(expectations)
    return {
        "expectations": expectations,
        "summary": {
            "pass_rate": (passed_count / total) if total else 0.0,
            "passed": passed_count,
            "failed": total - passed_count,
            "total": total,
        },
    }


def main():
    iteration_dir = Path(sys.argv[1]).resolve()
    for eval_dir in sorted(iteration_dir.glob("eval-*")):
        if not eval_dir.is_dir():
            continue
        meta_path = eval_dir / "eval_metadata.json"
        if not meta_path.exists():
            continue
        meta = json.loads(meta_path.read_text())
        for variant in ("with_skill", "without_skill", "old_skill"):
            run_dir = eval_dir / variant
            if not run_dir.exists():
                continue
            result = grade_run(run_dir, meta)
            (run_dir / "grading.json").write_text(json.dumps(result, indent=2))
            passed = sum(1 for e in result["expectations"] if e["passed"])
            total = len(result["expectations"])
            print(f"{eval_dir.name}/{variant}: {passed}/{total}")


if __name__ == "__main__":
    main()
