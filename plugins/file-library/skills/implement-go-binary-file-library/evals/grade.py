#!/usr/bin/env python3
"""Grade implement-go-binary-file-library eval runs.

Usage: python grade.py <iteration_dir>

Walks each eval-N/<config>/run-1/outputs/tlv/ directory, evaluates assertions,
and writes grading.json into each run-1/ directory in the format expected by
the eval-viewer (each expectation has fields: text, passed, evidence).
"""
import json
import re
import subprocess
import sys
from pathlib import Path


def load_metadata(eval_dir: Path) -> dict:
    return json.loads((eval_dir / "eval_metadata.json").read_text())


def read(path: Path) -> str:
    return path.read_text() if path.exists() else ""


def grep_count(pattern: str, text: str) -> int:
    return len(re.findall(pattern, text, re.MULTILINE))


def file_exists(pkg: Path, name: str) -> bool:
    return (pkg / name).exists()


def go_test_log_passed(run_dir: Path) -> bool:
    log = run_dir / "outputs" / "verify.log"
    if not log.exists():
        return False
    txt = log.read_text()
    if "FAIL" in txt or "build failed" in txt:
        return False
    return "ok" in txt or "PASS" in txt


def has_t_parallel_at_both_levels(text: str) -> bool:
    # crude: at least 2 t.Parallel() calls per test function we care about
    return text.count("t.Parallel()") >= 2


CONTEXT_FILENAMES = ("_context_types.md", "_context_decoder.md")
SOURCE_FILENAMES = ("types.go", "decoder.go", "encoder.go")
PHASE_NAMES = ("types", "decoder", "encoder")
PACKAGE_NAMES = {0: "tlv", 1: "tlv", 2: "tlv", 3: "tlvx"}


def parse_partition_plan(pkg: Path) -> dict:
    """Parse `<pkg>/partition_plan.md` into a structured summary.

    See the matching helper in the implement-go-text-file-library grade.py for
    the full file format. Phases for binary are `types` / `decoder` / `encoder`.
    """
    plan_path = pkg / "partition_plan.md"
    result = {
        "exists": plan_path.exists(),
        "raw": "",
        "phases": {p: {"partitioned": False, "sub_unit_count": 0, "summary_text": ""} for p in PHASE_NAMES},
        "sub_unit_log": [],
    }
    if not result["exists"]:
        return result

    text = plan_path.read_text()
    result["raw"] = text

    for phase in PHASE_NAMES:
        section = extract_section(text, f"{phase} phase")
        partitioned = False
        sub_unit_count = 0
        m = re.search(r"partitioned\s+into\s+(\d+)\s+sub[-\s]?units?", section, re.IGNORECASE)
        if m:
            partitioned = True
            sub_unit_count = int(m.group(1))
        result["phases"][phase] = {
            "partitioned": partitioned,
            "sub_unit_count": sub_unit_count,
            "summary_text": section,
        }

    pattern = re.compile(
        r"^##\s+(\w+)\s+sub[-\s]?unit\s+(\d+)\s*[—–\-]\s*(starting|done)(?:\s*\(([^)]*)\))?",
        re.MULTILINE | re.IGNORECASE,
    )
    for m in pattern.finditer(text):
        phase = m.group(1).lower()
        idx = int(m.group(2))
        kind = m.group(3).lower()
        ts = m.group(4) or ""
        if phase in PHASE_NAMES:
            result["sub_unit_log"].append((phase, idx, kind, ts))

    return result


def assertion_partition_plan_exists(pkg: Path) -> tuple[bool, str]:
    plan = parse_partition_plan(pkg)
    if not plan["exists"]:
        return False, "partition_plan.md not found in package"
    return True, f"partition_plan.md exists ({len(plan['raw'])} bytes)"


def assertion_partition_plan_announces_sub_units(pkg: Path) -> tuple[bool, str]:
    plan = parse_partition_plan(pkg)
    if not plan["exists"]:
        return False, "partition_plan.md not found"
    findings = []
    max_count = 0
    any_partitioned = False
    for phase, info in plan["phases"].items():
        if info["partitioned"]:
            any_partitioned = True
            findings.append(f"{phase}: {info['sub_unit_count']} sub-units")
            max_count = max(max_count, info["sub_unit_count"])
    if not any_partitioned:
        return False, "no phase reported as partitioned in partition_plan.md"
    if max_count < 2:
        return False, f"max sub-unit count is {max_count} (need >= 2): {findings}"
    return True, "; ".join(findings)


def assertion_partition_plan_serial_order(pkg: Path) -> tuple[bool, str]:
    plan = parse_partition_plan(pkg)
    if not plan["exists"]:
        return False, "partition_plan.md not found"

    by_phase: dict = {p: [] for p in PHASE_NAMES}
    for phase, idx, kind, ts in plan["sub_unit_log"]:
        by_phase[phase].append((idx, kind, ts))

    findings = []
    for phase, info in plan["phases"].items():
        if not info["partitioned"]:
            continue
        log = by_phase[phase]
        expected = info["sub_unit_count"]
        if len(log) < expected * 2:
            return False, f"{phase}: log has {len(log)} entries but {expected*2} expected (start+done per sub-unit)"
        for i in range(expected):
            s_idx, s_kind, _ = log[2 * i]
            d_idx, d_kind, _ = log[2 * i + 1]
            if (s_idx, s_kind) != (i + 1, "starting"):
                return False, f"{phase}: entry {2*i} = (sub-unit {s_idx}, {s_kind}); want (sub-unit {i+1}, starting)"
            if (d_idx, d_kind) != (i + 1, "done"):
                return False, f"{phase}: entry {2*i+1} = (sub-unit {d_idx}, {d_kind}); want (sub-unit {i+1}, done)"
        findings.append(f"{phase}: {expected} sub-units logged in serial order")
    if not findings:
        return False, "no partitioned phase had its serial-order log validated"
    return True, "; ".join(findings)


def assertion_small_fixture_no_partition(pkg: Path) -> tuple[bool, str]:
    plan = parse_partition_plan(pkg)
    if not plan["exists"]:
        return False, "partition_plan.md not found"
    findings = []
    for phase, info in plan["phases"].items():
        if info["partitioned"]:
            return False, f"{phase}: unexpectedly partitioned (sub-unit count: {info['sub_unit_count']})"
        if "no partitioning" not in info["summary_text"].lower():
            return False, f"{phase}: summary missing 'no partitioning' phrase"
        findings.append(f"{phase}: no partitioning")
    return True, "; ".join(findings)


def assertion_partition_common(asid: str, pkg: Path):
    """Dispatch the four partition-related assertions shared across all evals."""
    if asid == "small-fixture-no-partition":
        return assertion_small_fixture_no_partition(pkg)
    if asid == "partition-plan-exists":
        return assertion_partition_plan_exists(pkg)
    if asid == "partition-plan-announces-sub-units":
        return assertion_partition_plan_announces_sub_units(pkg)
    if asid == "partition-plan-serial-order":
        return assertion_partition_plan_serial_order(pkg)
    return None


def skill_md_text() -> str:
    """Read SKILL.md from the parent directory of this grade.py."""
    skill_md_path = Path(__file__).resolve().parent.parent / "SKILL.md"
    return skill_md_path.read_text() if skill_md_path.exists() else ""


def extract_section(text: str, heading: str) -> str:
    """Return the body of a top-level `## heading` section.

    Walks lines and tracks fenced code blocks so that `## ...` markers inside
    fenced templates (e.g. the literal `## Decode` line shown in the
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
    already appear in the unrelated decoder/encoder discussion).
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

    # Sliced-line gate trigger: a comparator phrase ("more than", ">",
    # "exceeds", "over") immediately preceding a >=3-digit number,
    # followed by "lines" within the same clause. Anchoring to the
    # comparator keeps the example total ("950 lines"), the per-subunit
    # cap ("<= 300 sliced lines"), and the unrelated "400-line cap" in
    # `## Context summary format` from satisfying the check on their own
    # — only the gate trigger ("more than 600 lines") qualifies.
    has_line_threshold = bool(re.search(
        r"(?:more\s+than|exceeds?|over|>)\s+\d{3,}[^,;.\n]*?\s+lines?",
        before_start,
        re.IGNORECASE,
    ))
    findings.append(f"line_threshold={has_line_threshold}")

    # Chunked-file gate trigger: same comparator-anchored shape, with
    # "chunked file(s)" in the tail. This isolates the gate trigger
    # ("more than 8 chunked files") from the per-subunit cap ("<= 4
    # chunked files each"), so removing the trigger fails the assertion
    # even while the cap remains.
    has_chunked_threshold = bool(re.search(
        r"(?:more\s+than|exceeds?|over|>)\s+\d+[^,;.\n]*?chunked\s+files?",
        before_start,
        re.IGNORECASE,
    ))
    findings.append(f"chunked_threshold={has_chunked_threshold}")

    # Partition along spec-section boundaries (issue #47's exact
    # acceptance phrasing). "section bound" matches both "section
    # boundary" and "section boundaries"; the "spec-" prefix is allowed
    # but not required.
    has_partition_subunits = (
        "partition" in before_lower
        and ("sub-unit" in before_lower or "sub-units" in before_lower)
    )
    findings.append(f"partition_subunits={has_partition_subunits}")

    has_section_boundary = "section bound" in before_lower
    findings.append(f"section_boundary={has_section_boundary}")

    has_announce = (
        "tell the user" in before_lower
        or "announce" in before_lower
        or "up front" in before_lower
    )
    findings.append(f"announce={has_announce}")

    gate_ok = (
        has_scope_gate
        and has_line_threshold
        and has_chunked_threshold
        and has_partition_subunits
        and has_section_boundary
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

    # No full-Read of the growing source file: a forbidding phrase plus
    # all three source filenames plus the word "read" must appear inside
    # the `## Phase chunking` section. Requiring *all* of SOURCE_FILENAMES
    # (mirroring the CONTEXT_FILENAMES check below) prevents the spec
    # from silently regressing to covering only one or two of the per-
    # phase source files. The forbidding-phrase set covers the current
    # "no full-`Read`" / "without a fresh whole-file read" /
    # "never the whole file" wordings without pinning the exact phrase.
    forbid_phrase = any(
        p in chunking_lower for p in ("no full", "whole file", "whole-file")
    )
    missing_sources = [name for name in SOURCE_FILENAMES if name not in chunking]
    all_sources_documented = not missing_sources
    forbids_full_read = (
        forbid_phrase and all_sources_documented and "read" in chunking_lower
    )
    findings.append(
        f"forbids_full_read={forbids_full_read}"
        f" (forbid_phrase={forbid_phrase}, "
        + (f"missing_sources={missing_sources}" if missing_sources else "all_sources_present=True")
        + ")"
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


def assertion_eval0_header(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    common = assertion_partition_common(asid, pkg)
    if common is not None:
        return common
    types = read(pkg / "types.go")
    decoder = read(pkg / "decoder.go")
    encoder = read(pkg / "encoder.go")
    decoder_test = read(pkg / "decoder_test.go")
    encoder_test = read(pkg / "encoder_test.go")
    types_test = read(pkg / "types_test.go")
    all_test = decoder_test + encoder_test + types_test

    if asid == "header-struct-defined":
        if not re.search(r"type\s+Header\s+struct", types):
            return False, "no Header struct in types.go"
        # check fields
        missing = [f for f in ("Magic", "Version", "Flags", "Reserved")
                   if not re.search(rf"\b{f}\b", types)]
        if missing:
            return False, f"Header missing fields: {missing}"
        return True, "Header struct present with Magic/Version/Flags/Reserved"

    if asid == "flags-bit-field-constants":
        names = ["COMPRESSED", "ENCRYPTED", "SIGNED", "Compressed", "Encrypted", "Signed"]
        found = [n for n in names if re.search(rf"\b\w*{n}\w*\s*[=:]", types) or re.search(rf"\bFlag\w*{n}\b", types)]
        # need all three concepts
        compressed = any("ompressed" in n for n in found)
        encrypted = any("ncrypted" in n for n in found)
        signed = any("igned" in n for n in found)
        if compressed and encrypted and signed:
            return True, f"found bit-field constants: {found}"
        return False, f"missing some of COMPRESSED/ENCRYPTED/SIGNED in types.go (found {found})"

    if asid == "file-includes-header":
        return ("Header" in types and re.search(r"type\s+File\s+struct\s*\{[^}]*Header", types) is not None,
                "File struct references Header")

    if asid == "read-header-method":
        ok = re.search(r"func\s*\(\s*\w+\s*\*decoder\s*\)\s*readHeader\b", decoder) is not None
        return ok, "readHeader on decoder" if ok else "no readHeader method on *decoder"

    if asid == "write-header-method":
        ok = re.search(r"func\s*\(\s*\w+\s*\*encoder\s*\)\s*writeHeader\b", encoder) is not None
        return ok, "writeHeader on encoder" if ok else "no writeHeader method on *encoder"

    if asid == "errors-funnel-through-wrapErr":
        # heuristic: in decoder.go and encoder.go, count occurrences of
        # FieldError{ or OffsetError{ outside the wrapErr function definitions.
        def bad_sites(src: str) -> list[str]:
            # strip the wrapErr definition body (one big regex)
            stripped = re.sub(r"func\s*\(\s*\w+\s*\*\w+\s*\)\s*wrapErr[\s\S]*?\n\}", "", src, count=1)
            sites = re.findall(r"&(?:FieldError|OffsetError)\{", stripped)
            return sites
        bad = bad_sites(decoder) + bad_sites(encoder)
        return len(bad) == 0, "all errors via wrapErr" if not bad else f"direct construction outside wrapErr: {len(bad)} sites"

    if asid == "decode-test-hex-literal":
        # Accept hex byte literals (0x54), hex strings ("54 4C 56 31"), or rune literals ('T','L','V','1')
        t = decoder_test
        has_magic = (
            "0x54" in t
            or '"TLV1"' in t
            or re.search(r"54\s*[, ]\s*4[Cc]\s*[, ]\s*56\s*[, ]\s*31", t) is not None
            or re.search(r"'T'\s*,\s*'L'\s*,\s*'V'\s*,\s*'1'", t) is not None
        )
        return has_magic, "decoder_test has TLV1 magic in byte/hex literal" if has_magic else "no TLV1 magic literal in decoder_test"

    if asid == "decode-failure-test-chain":
        has_is = "ErrorIs" in decoder_test or "errors.Is" in decoder_test
        has_as = "ErrorAs" in decoder_test or "errors.As" in decoder_test
        has_field = "FieldError" in decoder_test
        return (has_is and has_as and has_field,
                f"errorIs={has_is} errorAs={has_as} FieldError={has_field}")

    if asid == "encode-test-hex-literal":
        t = encoder_test
        ok = (
            "0x54" in t
            or '"TLV1"' in t
            or re.search(r"54\s*[, ]\s*4[Cc]\s*[, ]\s*56\s*[, ]\s*31", t) is not None
            or re.search(r"'T'\s*,\s*'L'\s*,\s*'V'\s*,\s*'1'", t) is not None
        )
        return ok, "encoder_test asserts hex-literal output" if ok else "no hex-literal encode assertion"

    if asid == "round-trip-test":
        ok = ("RoundTrip" in encoder_test or "roundtrip" in encoder_test.lower() or "round_trip" in encoder_test) \
             and "Encode" in encoder_test and "Decode" in encoder_test
        return ok, "round-trip test present" if ok else "no round-trip test in encoder_test"

    if asid == "tests-parallel":
        for name, txt in [("decoder_test.go", decoder_test), ("encoder_test.go", encoder_test)]:
            if "t.Parallel()" not in txt:
                return False, f"{name} missing t.Parallel()"
            # at least 2 occurrences (function + subtest)
            if txt.count("t.Parallel()") < 2:
                return False, f"{name} only one t.Parallel() (no subtest parallelism)"
        return True, "t.Parallel at both levels in decoder/encoder tests"

    if asid == "tests-testify-require":
        ok = "github.com/stretchr/testify/require" in (decoder_test + encoder_test + types_test)
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


def assertion_eval1_record(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    common = assertion_partition_common(asid, pkg)
    if common is not None:
        return common
    types = read(pkg / "types.go")
    decoder = read(pkg / "decoder.go")
    encoder = read(pkg / "encoder.go")
    decoder_test = read(pkg / "decoder_test.go")
    encoder_test = read(pkg / "encoder_test.go")

    if asid == "record-struct-defined":
        ok = re.search(r"type\s+Record\s+struct", types) is not None
        if ok:
            missing = [f for f in ("Type", "Length", "Value") if not re.search(rf"\b{f}\b", types)]
            if missing:
                return False, f"Record missing fields: {missing}"
            return True, "Record struct with Type/Length/Value"
        return False, "no Record struct"

    if asid == "record-type-enum":
        ok = "type RecordType" in types
        for name in ("STRING", "String", "INT", "Int", "BLOB", "Blob", "NESTED", "Nested"):
            pass
        # check at least one constant per concept
        groups = [
            ["STRING", "String"],
            ["INT", "Int"],
            ["BLOB", "Blob"],
            ["NESTED", "Nested"],
        ]
        present = []
        for g in groups:
            present.append(any(re.search(rf"\bRecordType{n}\b", types) for n in g))
        return (ok and all(present),
                f"RecordType enum + all 4 constants" if ok and all(present) else f"missing RecordType or constants: groups present={present}")

    if asid == "record-type-string-method":
        ok = re.search(r"func\s*\(\s*\w+\s+RecordType\s*\)\s*String\s*\(\s*\)\s*string", types) is not None
        return ok, "String() on RecordType" if ok else "no String() method on RecordType"

    if asid == "file-includes-records":
        ok = re.search(r"type\s+File\s+struct\s*\{[^}]*Records\s+\[\]Record", types) is not None
        return ok, "File.Records []Record" if ok else "no Records slice on File"

    if asid == "read-record-method":
        ok = re.search(r"func\s*\(\s*\w+\s*\*decoder\s*\)\s*readRecord\w*", decoder) is not None
        return ok, "readRecord(...) on decoder" if ok else "no readRecord* method"

    if asid == "write-record-method":
        ok = re.search(r"func\s*\(\s*\w+\s*\*encoder\s*\)\s*writeRecord\w*", encoder) is not None
        return ok, "writeRecord(...) on encoder" if ok else "no writeRecord* method"

    if asid == "uses-readfull-for-value":
        ok = "io.ReadFull" in decoder
        return ok, "decoder uses io.ReadFull" if ok else "decoder does not use io.ReadFull"

    if asid == "decode-test-string-record":
        ok = ("STRING" in decoder_test or "String" in decoder_test) and "Length" not in decoder_test or True
        # weaker check: any non-empty value test for STRING type
        ok = ("RecordTypeString" in decoder_test or "RecordTypeSTRING" in decoder_test or "0x01" in decoder_test) \
             and ("hello" in decoder_test or "0x68" in decoder_test or "Value" in decoder_test)
        return ok, "STRING record decode test present" if ok else "no clear STRING decode test"

    if asid == "decode-test-empty-record":
        ok = re.search(r"Length\s*[:=]\s*0\b", decoder_test) is not None or "empty" in decoder_test.lower()
        return ok, "empty-record decode test present" if ok else "no empty (Length=0) test"

    if asid == "round-trip-test":
        ok = ("RoundTrip" in encoder_test or "round" in encoder_test.lower()) \
             and "Encode" in encoder_test and "Decode" in encoder_test
        return ok, "round-trip test present" if ok else "no round-trip test"

    if asid == "errors-funnel-through-wrapErr":
        def bad_sites(src: str) -> list[str]:
            stripped = re.sub(r"func\s*\(\s*\w+\s*\*\w+\s*\)\s*wrapErr[\s\S]*?\n\}", "", src, count=1)
            return re.findall(r"&(?:FieldError|OffsetError)\{", stripped)
        bad = bad_sites(decoder) + bad_sites(encoder)
        return len(bad) == 0, "all errors via wrapErr" if not bad else f"direct construction outside wrapErr: {len(bad)} sites"

    if asid == "tests-parallel":
        for name, txt in [("decoder_test.go", decoder_test), ("encoder_test.go", encoder_test)]:
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


def assertion_eval2_trailer(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    common = assertion_partition_common(asid, pkg)
    if common is not None:
        return common
    types = read(pkg / "types.go")
    decoder = read(pkg / "decoder.go")
    encoder = read(pkg / "encoder.go")
    decoder_test = read(pkg / "decoder_test.go")
    encoder_test = read(pkg / "encoder_test.go")

    if asid == "trailer-or-crc-on-file":
        ok = re.search(r"type\s+Trailer\s+struct", types) is not None or \
             re.search(r"type\s+File\s+struct\s*\{[^}]*CRC32", types) is not None
        return ok, "Trailer struct or File.CRC32 present" if ok else "no Trailer or CRC32 field"

    if asid == "checksum-error-sentinel":
        ok = "ErrChecksumMismatch" in types or "ErrChecksum" in types or "ErrCRC" in types
        return ok, "checksum-mismatch sentinel defined" if ok else "no ErrChecksumMismatch sentinel"

    if asid == "decoder-uses-crc32-ieee":
        ok = "hash/crc32" in decoder and ("IEEETable" in decoder or "NewIEEE" in decoder or "ChecksumIEEE" in decoder)
        return ok, "decoder uses hash/crc32 IEEE" if ok else "decoder missing IEEE crc32 usage"

    if asid == "encoder-uses-crc32-ieee":
        ok = "hash/crc32" in encoder and ("IEEETable" in encoder or "NewIEEE" in encoder or "ChecksumIEEE" in encoder)
        return ok, "encoder uses hash/crc32 IEEE" if ok else "encoder missing IEEE crc32 usage"

    if asid == "decoder-verifies-crc":
        ok = ("ErrChecksum" in decoder or "ErrCRC" in decoder) and ("crc" in decoder.lower())
        return ok, "decoder compares CRC and returns error" if ok else "decoder does not verify CRC properly"

    if asid == "encoder-writes-crc":
        # encoder must write 4 bytes from a CRC32 sum
        ok = ("Sum32" in encoder or "ChecksumIEEE" in encoder) and \
             ("PutUint32" in encoder or "binary.Write" in encoder or "binary.BigEndian" in encoder)
        return ok, "encoder writes CRC32" if ok else "encoder does not write CRC32 properly"

    if asid == "happy-path-test":
        ok = "TestDecode" in decoder_test or "happy" in decoder_test.lower() or "valid" in decoder_test.lower()
        return ok, "happy-path decode test present" if ok else "no happy-path decode test"

    if asid == "mismatch-test":
        has_mismatch = "Mismatch" in decoder_test or "Corrupt" in decoder_test or "corrupt" in decoder_test
        has_assert = "ErrorIs" in decoder_test and "ErrChecksum" in decoder_test
        return has_assert, "asserts errors.Is(err, ErrChecksumMismatch)" if has_assert else f"mismatch_phrase={has_mismatch} ErrorIs+ErrChecksum={has_assert}"

    if asid == "encoder-roundtrip-test":
        ok = ("RoundTrip" in encoder_test or "round" in encoder_test.lower()) \
             and "Encode" in encoder_test and "Decode" in encoder_test
        return ok, "round-trip test present" if ok else "no round-trip in encoder_test"

    if asid == "errors-funnel-through-wrapErr":
        # crc mismatch site funnels through wrapErr
        ok = re.search(r"wrapErr\s*\([^)]*Trailer", decoder) is not None or \
             re.search(r"wrapErr\s*\([^)]*CRC", decoder) is not None
        return ok, "CRC error funnels through wrapErr" if ok else "CRC error site does not use wrapErr"

    if asid == "tests-parallel":
        for name, txt in [("decoder_test.go", decoder_test), ("encoder_test.go", encoder_test)]:
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


def assertion_eval3_tlvx_header_extended(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
    """Grader for the gate-tripping tlvx eval (issue #55).

    The eval exercises the `partition_plan.md` mechanism: tlvx's decoder-phase
    slices clear the 600-line gate, so the orchestrator must partition. The
    partition-plan assertions are dispatched via `assertion_partition_common`.
    The remaining assertions check the actual implementation: a Header struct
    with the eight TLVX fields (Magic / Version / Flags / ChecksumAlg /
    Reserved1 / IndexCount / ExtCount / TrailerOffset), the seven defined Flags
    bit-field constants, and the ChecksumAlg enum with its String() method.
    """
    common = assertion_partition_common(asid, pkg)
    if common is not None:
        return common

    types = read(pkg / "types.go")
    decoder = read(pkg / "decoder.go")
    encoder = read(pkg / "encoder.go")
    decoder_test = read(pkg / "decoder_test.go")
    encoder_test = read(pkg / "encoder_test.go")
    types_test = read(pkg / "types_test.go")
    all_test = decoder_test + encoder_test + types_test

    if asid == "header-struct-defined":
        if not re.search(r"type\s+Header\s+struct", types):
            return False, "no Header struct in types.go"
        missing = [f for f in ("Magic", "Version", "Flags", "ChecksumAlg",
                               "Reserved1", "IndexCount", "ExtCount", "TrailerOffset")
                   if not re.search(rf"\b{f}\b", types)]
        if missing:
            return False, f"Header missing fields: {missing}"
        return True, "Header struct present with all eight TLVX fields"

    if asid == "flags-bit-field-constants":
        # The seven defined Flags: COMPRESSED, ENCRYPTED, SIGNED, INDEXED,
        # EXTENDED, STRICT, SEALED. (TLVX bit 7 is reserved.) Accept "Flag"
        # prefix or bare names.
        required = ["COMPRESSED", "ENCRYPTED", "SIGNED", "INDEXED", "EXTENDED", "STRICT", "SEALED"]
        missing = [f for f in required if not re.search(rf"\b\w*{f}\w*\b", types)]
        if missing:
            return False, f"Flags constants missing: {missing}"
        return True, "all seven defined TLVX flag constants present"

    if asid == "checksum-alg-enum":
        # Look for a named type ChecksumAlg with constants. Accept any of the
        # five algorithm names being present (we ask for at least three in the
        # assertion text).
        if not re.search(r"type\s+ChecksumAlg\b", types):
            return False, "no ChecksumAlg type defined in types.go"
        algs = [a for a in ("CRC32_IEEE", "CRC32IEEE", "CRC64_ECMA", "CRC64ECMA",
                            "SHA256_T32", "SHA256T32", "XXH64", "BLAKE3_T32", "BLAKE3T32")
                if re.search(rf"\b{a}\b", types)]
        # Normalise: collapse equivalent names down to a unique-algorithm count
        normalized = set()
        for a in algs:
            if "CRC32" in a: normalized.add("CRC32")
            elif "CRC64" in a: normalized.add("CRC64")
            elif "SHA256" in a: normalized.add("SHA256")
            elif "XXH64" in a: normalized.add("XXH64")
            elif "BLAKE3" in a: normalized.add("BLAKE3")
        if len(normalized) < 3:
            return False, f"only {len(normalized)} ChecksumAlg constants found ({sorted(normalized)}); need >= 3"
        return True, f"ChecksumAlg enum with {len(normalized)} algorithms: {sorted(normalized)}"

    if asid == "checksum-alg-string-method":
        ok = re.search(r"func\s*\(\s*\w+\s+ChecksumAlg\s*\)\s+String\s*\(\s*\)\s+string", types) is not None
        return ok, "ChecksumAlg.String() defined" if ok else "no String() method on ChecksumAlg"

    if asid == "file-includes-header":
        ok = re.search(r"type\s+File\s+struct\s*\{[^}]*Header", types) is not None
        return ok, "File struct references Header" if ok else "File struct does not include Header"

    if asid == "read-header-method":
        ok = re.search(r"func\s*\(\s*\w+\s+\*?decoder\s*\)\s+readHeader\b", decoder) is not None
        return ok, "readHeader method on decoder" if ok else "no readHeader method"

    if asid == "write-header-method":
        ok = re.search(r"func\s*\(\s*\w+\s+\*?encoder\s*\)\s+writeHeader\b", encoder) is not None
        return ok, "writeHeader method on encoder" if ok else "no writeHeader method"

    if asid == "errors-funnel-through-wrapErr":
        # Any direct construction of FieldError or OffsetError outside the
        # wrapErr helpers fails this check. Strip the helper bodies first.
        d_stripped = re.sub(r"func\s*\(\s*\w\s+\*?decoder\s*\)\s+wrapErr[\s\S]*?\n\}", "", decoder, count=1)
        e_stripped = re.sub(r"func\s*\(\s*\w\s+\*?encoder\s*\)\s+wrapErr[\s\S]*?\n\}", "", encoder, count=1)
        bad = []
        for label, txt in (("decoder.go", d_stripped), ("encoder.go", e_stripped)):
            for direct in re.findall(r"&FieldError\s*\{|&OffsetError\s*\{", txt):
                bad.append(f"{label}: {direct}")
        ok = len(bad) == 0
        return ok, ("no direct FieldError/OffsetError outside wrapErr" if ok else f"direct error construction: {bad[:3]}")

    if asid == "decode-test-hex-literal":
        # Test that decodes a hex byte literal containing the TLVX magic
        # (54 4C 56 58 or "TLVX") and asserts Header field values.
        has_magic = ("0x54" in decoder_test and "0x4C" in decoder_test) or '"TLVX"' in decoder_test
        has_decode = re.search(r"\bDecode\s*\(", decoder_test) is not None
        ok = has_magic and has_decode
        return ok, f"decode test with TLVX magic (has_magic={has_magic}, has_decode={has_decode})"

    if asid == "decode-failure-test-chain":
        # Failure-path test asserting errors.Is(leaf) AND errors.As(*FieldError)
        has_is = re.search(r"errors\.Is\b|require\.ErrorIs\b", decoder_test) is not None
        has_as = re.search(r"errors\.As\b|require\.ErrorAs\b", decoder_test) is not None
        has_field_error = "FieldError" in decoder_test
        ok = has_is and has_as and has_field_error
        return ok, f"failure-path chain test (Is={has_is}, As={has_as}, FieldError={has_field_error})"

    if asid == "encode-test-hex-literal":
        has_encode = re.search(r"\bEncode\s*\(", encoder_test) is not None
        # Hex bytes in the assertion (TLVX magic) or a string-literal "TLVX"
        has_hex = "0x54" in encoder_test or "0x4C" in encoder_test or '"TLVX"' in encoder_test
        ok = has_encode and has_hex
        return ok, f"encode happy-path test (has_encode={has_encode}, has_hex={has_hex})"

    if asid == "round-trip-test":
        ok = ("Encode" in encoder_test and "Decode" in encoder_test) and \
             ("RoundTrip" in encoder_test or "round" in encoder_test.lower())
        return ok, "round-trip test present" if ok else "no round-trip test"

    if asid == "tests-parallel":
        for name, txt in [("decoder_test.go", decoder_test), ("encoder_test.go", encoder_test), ("types_test.go", types_test)]:
            if txt and txt.count("t.Parallel()") < 2:
                return False, f"{name} insufficient t.Parallel"
        return True, "t.Parallel at both levels in all test files"

    if asid == "tests-testify-require":
        ok = "github.com/stretchr/testify/require" in all_test
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


GRADERS = {
    0: assertion_eval0_header,
    1: assertion_eval1_record,
    2: assertion_eval2_trailer,
    3: assertion_eval3_tlvx_header_extended,
}


def grade_run(eval_dir: Path, config: str) -> dict:
    meta = load_metadata(eval_dir)
    eval_id = meta["eval_id"]
    grader = GRADERS[eval_id]
    run_dir = eval_dir / config / "run-1"
    pkg = run_dir / "outputs" / PACKAGE_NAMES[eval_id]
    expectations = []
    for a in meta["assertions"]:
        passed, evidence = grader(a["id"], pkg, run_dir)
        expectations.append({
            "text": a["text"],
            "passed": bool(passed),
            "evidence": evidence,
        })
    n_passed = sum(1 for e in expectations if e["passed"])
    pass_rate = n_passed / len(expectations)
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
            print(f"{eval_dir.name}/{cfg}: {n_pass}/{len(result['expectations'])} ({result['summary']['pass_rate']*100:.0f}%)")


if __name__ == "__main__":
    main()
