#!/usr/bin/env python3
"""Grade an extract-binary-spec eval run against its eval_metadata.json assertions."""
import argparse
import json
import re
import sys
from pathlib import Path


def find_root_with_spec(outputs_dir: Path) -> Path | None:
    """Locate the directory containing SPEC.md (skill-conformant) or fall back to outputs_dir."""
    if (outputs_dir / "SPEC.md").is_file():
        return outputs_dir
    for child in outputs_dir.iterdir():
        if child.is_dir() and (child / "SPEC.md").is_file():
            return child
    return None


def find_root(outputs_dir: Path) -> Path:
    """Locate the deepest single-child directory containing the output (skill-conformant or not)."""
    root = find_root_with_spec(outputs_dir)
    if root is not None:
        return root
    children = [c for c in outputs_dir.iterdir() if c.is_dir()]
    if len(children) == 1:
        return children[0]
    return outputs_dir


def read_text(path: Path) -> str:
    try:
        return path.read_text(encoding="utf-8", errors="replace")
    except Exception:
        return ""


def all_md_text(root: Path) -> str:
    return "\n".join(read_text(p) for p in sorted(root.rglob("*.md")))


def list_md(root: Path, subdir: str) -> list[Path]:
    sub = root / subdir
    if not sub.is_dir():
        return []
    return sorted([p for p in sub.iterdir() if p.is_file() and p.suffix == ".md"])


# Per-assertion checkers --------------------------------------------------------

def check_layout_spec_md_exists(root: Path) -> tuple[bool, str]:
    p = root / "SPEC.md"
    return p.is_file(), f"SPEC.md {'present' if p.is_file() else 'missing'} at {p}"


def check_layout_structures_dir(root: Path, min_count: int) -> tuple[bool, str]:
    files = list_md(root, "structures")
    return len(files) >= min_count, f"structures/ has {len(files)} .md files (need ≥{min_count})"


def check_layout_encoding_tables_dir(root: Path, min_count: int) -> tuple[bool, str]:
    files = list_md(root, "encoding-tables")
    return len(files) >= min_count, f"encoding-tables/ has {len(files)} .md files (need ≥{min_count})"


def check_layout_examples_dir(root: Path, min_count: int) -> tuple[bool, str]:
    files = list_md(root, "examples")
    return len(files) >= min_count, f"examples/ has {len(files)} .md files (need ≥{min_count})"


def check_no_empty_files(root: Path, min_bytes: int = 200) -> tuple[bool, str]:
    bad = []
    for p in root.rglob("*.md"):
        if p.stat().st_size < min_bytes:
            bad.append(f"{p.relative_to(root)} ({p.stat().st_size}B)")
    if bad:
        return False, "files under min size: " + ", ".join(bad)
    return True, "all .md files ≥ 200 bytes"


def check_no_scratch_leftovers(outputs_dir: Path) -> tuple[bool, str]:
    leftovers = []
    for pattern in ["_spec.txt", "_spec.html", "_toc.md", "_structures_index.md"]:
        for p in outputs_dir.rglob(pattern):
            leftovers.append(str(p.relative_to(outputs_dir)))
    if leftovers:
        return False, "scratch files left behind: " + ", ".join(leftovers)
    return True, "no scratch files in output tree"


def check_spec_conventions_section(root: Path) -> tuple[bool, str]:
    text = read_text(root / "SPEC.md").lower()
    has = "conventions" in text and re.search(r"byte\s*[- ]?\s*order", text) is not None
    return has, f"SPEC.md {'has' if has else 'missing'} a Conventions section that mentions byte order"


def check_byte_order_little_endian(root: Path) -> tuple[bool, str]:
    text = read_text(root / "SPEC.md").lower()
    pat = re.search(r"little[- ]endian|lsb[- ]first|intel\s*byte\s*order", text)
    return bool(pat), f"SPEC.md byte-order term: {pat.group(0) if pat else 'NOT FOUND'}"


def check_byte_order_network(root: Path) -> tuple[bool, str]:
    text = read_text(root / "SPEC.md").lower()
    pat = re.search(r"network\s*byte\s*order|big[- ]endian|msb[- ]first", text)
    return bool(pat), f"SPEC.md byte-order term: {pat.group(0) if pat else 'NOT FOUND'}"


def check_structures_index(root: Path) -> tuple[bool, str]:
    text = read_text(root / "SPEC.md")
    has_section = re.search(r"#+\s*Structures?\s+(?:index|Index)", text, re.IGNORECASE) is not None
    has_links = re.search(r"\(structures/[\w\-]+\.md\)", text) is not None
    ok = has_section and has_links
    return ok, f"index section: {has_section}, structures/*.md links present: {has_links}"


_VARIANT_MARKERS = re.compile(
    r"recursive\s*/\s*algorithmic\s*encoding|"
    r"recursive\s+or\s+algorithmic\s+encoding|"
    r"bit[- ]only\s+structure|"
    r"##\s*Encoding\b",
    re.IGNORECASE,
)


def check_structures_field_table(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    if not files:
        return False, "no structures/*.md files"
    cols_pat = re.compile(r"\|\s*offset[^|]*\|\s*size[^|]*\|\s*type[^|]*\|\s*name[^|]*\|\s*description", re.IGNORECASE)
    bad = []
    for f in files:
        text = read_text(f)
        if cols_pat.search(text):
            continue
        # Accept the recursive / algorithmic / bit-only variant per output-format.md
        if _VARIANT_MARKERS.search(text):
            continue
        bad.append(f.name)
    if bad:
        return False, f"missing field table cols in: {', '.join(bad)}"
    return True, f"all {len(files)} structures have a canonical field table or declared variant"


# Allowed Go type forms in the Type column of a field table:
#   - primitives: uint8/16/32/64, int8/16/32/64
#   - fixed-size byte arrays: [N]byte
#   - variable-length byte arrays: []byte
#   - reference to another structure: PascalCase identifier (>= 2 chars)
_TYPE_CELL = re.compile(
    r"\|\s*("
    r"uint8|uint16|uint32|uint64|int8|int16|int32|int64|"
    r"\[\d+\]byte|\[\]byte|"
    r"[A-Z][A-Za-z0-9]{2,}"
    r")\s*\|"
)
# Header-row words to exclude when matching the third column position
_HEADER_WORDS = {"offset", "size", "type", "name", "description", "value", "bit", "bits", "field", "notes"}


def check_structures_go_types(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    if not files:
        return False, "no structures/*.md files"
    bad = []
    for f in files:
        text = read_text(f)
        # Find any allowed type token in a "| ... |" cell that isn't a header word.
        ok = False
        for m in _TYPE_CELL.finditer(text):
            tok = m.group(1)
            if tok.lower() in _HEADER_WORDS:
                continue
            ok = True
            break
        if not ok:
            bad.append(f.name)
    if bad:
        return False, f"no Go-friendly types found in: {', '.join(bad)}"
    return True, f"all {len(files)} structures use Go-friendly types"


# gzip-specific checks ----------------------------------------------------------

def check_gzip_flg_bits(root: Path) -> tuple[bool, str]:
    text = all_md_text(root)
    needed = ["FTEXT", "FHCRC", "FEXTRA", "FNAME", "FCOMMENT"]
    missing = [n for n in needed if n not in text]
    return not missing, f"FLG bit names missing: {missing or 'none'}"


def check_gzip_magic_id(root: Path) -> tuple[bool, str]:
    text = all_md_text(root).lower()
    # accept "0x1f" + "0x8b" anywhere, or "1f 8b"
    has_1f = re.search(r"\b0x1f\b|\b1f\s+8b\b", text) is not None
    has_8b = re.search(r"\b0x8b\b|\b1f\s+8b\b", text) is not None
    ok = has_1f and has_8b
    return ok, f"ID1=0x1f present: {has_1f}, ID2=0x8b present: {has_8b}"


def check_gzip_trailer_crc32_isize(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    for f in files:
        text = read_text(f).upper()
        if "CRC32" in text and "ISIZE" in text:
            return True, f"trailer fields found in {f.name}"
    return False, "no structure file mentions both CRC32 and ISIZE"


def check_gzip_os_table(root: Path) -> tuple[bool, str]:
    # accept: an encoding-tables file, OR an inline OS table with at least 5 entries
    et_files = list_md(root, "encoding-tables")
    for f in et_files:
        if "operating" in f.name.lower() or "os" in f.name.lower():
            text = read_text(f)
            if text.count("|") >= 15:  # at least 5 table rows × 3 columns of pipes
                return True, f"OS encoding table at {f.name}"
    text = all_md_text(root)
    # search for OS values like 0=FAT, 3=Unix, etc.
    matches = re.findall(r"\b(?:0|1|2|3|4|5|6|7|8|9|10|11|12|13|255)\b\s*[|=]\s*[A-Z][A-Za-z\-/ ]+", text)
    if len(matches) >= 5:
        return True, f"inline OS values: {len(matches)} entries"
    return False, "no OS encoding table or sufficient inline values"


def check_gzip_optional_conditional(root: Path) -> tuple[bool, str]:
    text = all_md_text(root).lower()
    needed_opt = ["fextra", "fname", "fcomment", "fhcrc"]
    needed_cond = ["flg"]
    has_opt = all(n in text for n in needed_opt)
    has_cond = "flg" in text and ("conditional" in text or "optional" in text or "if " in text)
    return has_opt and has_cond, f"optional fields documented: {has_opt}, conditional on FLG: {has_cond}"


def check_gzip_deflate_skipped(root: Path) -> tuple[bool, str]:
    structs = list_md(root, "structures")
    bad = [f.name for f in structs if "deflate" in f.name.lower() or "compressed" in f.name.lower()]
    if bad:
        return False, f"deflate/compressed-data structure files present (should be skipped): {bad}"
    return True, "no deflate structure files in structures/ — scope respected"


# DNS-specific checks -----------------------------------------------------------

def check_dns_header_12_bytes(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    target = next((f for f in files if "header" in f.name.lower()), None)
    if not target:
        return False, "no structure file with 'header' in name"
    text = read_text(target).upper()
    needed = ["ID", "QDCOUNT", "ANCOUNT", "NSCOUNT", "ARCOUNT"]
    missing = [n for n in needed if n not in text]
    return not missing, f"header.md missing fields: {missing or 'none'}"


def check_dns_header_bit_fields(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    target = next((f for f in files if "header" in f.name.lower()), None)
    if not target:
        return False, "no structure file with 'header' in name"
    text = read_text(target)
    needed = ["QR", "Opcode", "AA", "TC", "RD", "RA", "Z", "RCODE"]
    # accept case-insensitive
    text_u = text.upper()
    missing = [n for n in needed if n.upper() not in text_u]
    return not missing, f"header.md missing flag bits: {missing or 'none'}"


def check_dns_compression(root: Path) -> tuple[bool, str]:
    text = all_md_text(root)
    has_label = re.search(r"length[- ]prefix(?:ed)?\s+label|label\s+(?:is\s+)?length[- ]prefix", text, re.IGNORECASE) is not None or \
                re.search(r"length\s+(?:byte|octet)", text, re.IGNORECASE) is not None
    has_pointer = (
        re.search(r"\b11\b.*two\s+(?:high|leading)\s+bits|two\s+(?:high|leading)\s+bits.*\b11\b|0xc0|0b11", text, re.IGNORECASE) is not None
        or re.search(r"14[- ]bit\s+offset", text, re.IGNORECASE) is not None
    )
    ok = has_label and has_pointer
    return ok, f"label format: {has_label}, compression pointer: {has_pointer}"


def check_dns_encoding_tables(root: Path) -> tuple[bool, str]:
    files = list_md(root, "encoding-tables")
    names = [f.stem.lower() for f in files]
    text_blob = " ".join(names)
    needs = {
        "opcode": any("opcode" in n for n in names),
        "rcode": any("rcode" in n or "response-code" in n for n in names),
        "type": any(n in ("types", "type", "rr-types", "qtypes") or "type" in n for n in names),
        "class": any("class" in n for n in names),
    }
    missing = [k for k, v in needs.items() if not v]
    return not missing, f"encoding tables: {names}; missing categories: {missing or 'none'}"


def check_dns_rr_format(root: Path) -> tuple[bool, str]:
    files = list_md(root, "structures")
    target = next((f for f in files if "resource" in f.name.lower() or f.name.lower() in ("rr.md", "record.md")), None)
    if not target:
        return False, "no resource-record structure file found"
    text = read_text(target).upper()
    needed = ["NAME", "TYPE", "CLASS", "TTL", "RDLENGTH", "RDATA"]
    missing = [n for n in needed if n not in text]
    return not missing, f"resource-record fields missing: {missing or 'none'}"


def check_dns_zone_files_skipped(root: Path) -> tuple[bool, str]:
    structs = list_md(root, "structures")
    bad = [f.name for f in structs if "zone" in f.name.lower() or "master" in f.name.lower() or "rfc1035-section-5" in f.name.lower()]
    if bad:
        return False, f"zone-file structures present (should be skipped): {bad}"
    return True, "no zone-file structures — scope respected"


# Dispatcher --------------------------------------------------------------------

CHECKS = {
    "layout-spec-md-exists": lambda root, outputs: check_layout_spec_md_exists(root),
    "layout-structures-dir": lambda root, outputs: check_layout_structures_dir(root, 2),
    "layout-encoding-tables-dir": lambda root, outputs: check_layout_encoding_tables_dir(root, 4),
    "layout-examples-dir": lambda root, outputs: check_layout_examples_dir(root, 3),
    "layout-no-empty-files": lambda root, outputs: check_no_empty_files(root),
    "layout-no-scratch-leftovers": lambda root, outputs: check_no_scratch_leftovers(outputs),
    "spec-conventions-section": lambda root, outputs: check_spec_conventions_section(root),
    "spec-byte-order-little-endian": lambda root, outputs: check_byte_order_little_endian(root),
    "spec-byte-order-network": lambda root, outputs: check_byte_order_network(root),
    "spec-structures-index": lambda root, outputs: check_structures_index(root),
    "structures-have-field-table": lambda root, outputs: check_structures_field_table(root),
    "structures-go-types": lambda root, outputs: check_structures_go_types(root),
    "gzip-flg-bit-field": lambda root, outputs: check_gzip_flg_bits(root),
    "gzip-magic-id-bytes": lambda root, outputs: check_gzip_magic_id(root),
    "gzip-trailer-crc32-isize": lambda root, outputs: check_gzip_trailer_crc32_isize(root),
    "gzip-os-encoding-table": lambda root, outputs: check_gzip_os_table(root),
    "gzip-optional-fields-conditional": lambda root, outputs: check_gzip_optional_conditional(root),
    "gzip-out-of-scope-deflate-skipped": lambda root, outputs: check_gzip_deflate_skipped(root),
    "dns-header-12-bytes": lambda root, outputs: check_dns_header_12_bytes(root),
    "dns-header-bit-fields": lambda root, outputs: check_dns_header_bit_fields(root),
    "dns-domain-name-compression": lambda root, outputs: check_dns_compression(root),
    "dns-encoding-tables-present": lambda root, outputs: check_dns_encoding_tables(root),
    "dns-rr-format": lambda root, outputs: check_dns_rr_format(root),
    "dns-out-of-scope-zone-files-skipped": lambda root, outputs: check_dns_zone_files_skipped(root),
}


def grade(run_dir: Path, eval_metadata_path: Path) -> dict:
    metadata = json.loads(eval_metadata_path.read_text())
    outputs_dir = run_dir / "outputs"
    root = find_root(outputs_dir)
    results = {
        "eval_id": metadata["eval_id"],
        "eval_name": metadata["eval_name"],
        "run_dir": str(run_dir),
        "root_used_for_grading": str(root),
        "expectations": [],
    }
    for assertion in metadata["assertions"]:
        aid = assertion["id"]
        text = assertion["text"]
        check = CHECKS.get(aid)
        if check is None:
            results["expectations"].append({
                "text": text,
                "passed": False,
                "evidence": f"no checker registered for id={aid}",
            })
            continue
        try:
            passed, evidence = check(root, outputs_dir)
        except Exception as e:
            passed, evidence = False, f"checker error: {type(e).__name__}: {e}"
        results["expectations"].append({
            "text": text,
            "passed": passed,
            "evidence": evidence,
        })
    return results


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("run_dir", help="path to a run directory containing outputs/ and the eval_metadata.json one level up")
    parser.add_argument("--eval-metadata", help="explicit path to eval_metadata.json")
    args = parser.parse_args()
    run_dir = Path(args.run_dir)
    if args.eval_metadata:
        meta = Path(args.eval_metadata)
    else:
        meta = run_dir.parent / "eval_metadata.json"
    if not meta.exists():
        print(f"missing eval_metadata.json at {meta}", file=sys.stderr)
        sys.exit(1)
    results = grade(run_dir, meta)
    out = run_dir / "grading.json"
    out.write_text(json.dumps(results, indent=2))
    passed = sum(1 for e in results["expectations"] if e["passed"])
    total = len(results["expectations"])
    print(f"{run_dir.name}: {passed}/{total} passed -> {out}")


if __name__ == "__main__":
    main()
