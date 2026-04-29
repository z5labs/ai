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


def assertion_eval0_header(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
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

    return False, f"unknown assertion id: {asid}"


def assertion_eval1_record(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
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

    return False, f"unknown assertion id: {asid}"


def assertion_eval2_trailer(asid: str, pkg: Path, run_dir: Path) -> tuple[bool, str]:
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

    return False, f"unknown assertion id: {asid}"


GRADERS = {
    0: assertion_eval0_header,
    1: assertion_eval1_record,
    2: assertion_eval2_trailer,
}


def grade_run(eval_dir: Path, config: str) -> dict:
    meta = load_metadata(eval_dir)
    eval_id = meta["eval_id"]
    grader = GRADERS[eval_id]
    run_dir = eval_dir / config / "run-1"
    pkg = run_dir / "outputs" / "tlv"
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
