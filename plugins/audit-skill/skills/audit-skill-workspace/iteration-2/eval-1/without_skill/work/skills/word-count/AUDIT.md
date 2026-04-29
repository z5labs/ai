# Audit Report: word-count skill

**Target:** `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-1/without_skill/work/skills/word-count/SKILL.md`
**Date:** 2026-04-29
**Scope:** Single-file skill (SKILL.md only, 1760 bytes). No supporting scripts, references, or assets.

## Summary

The skill is small, well-scoped, and largely well-written. The frontmatter is correct, the workflow is clear, and "when to skip" is appropriately narrow. There are a handful of correctness/robustness issues worth fixing, plus a few smaller polish items. None are blocking, but the `wc -wlc` parsing instruction is the most impactful fix.

Overall grade: solid skill with minor issues. Recommended fixes are itemized below in priority order.

---

## Findings

### High priority

#### H1. `wc -wlc` field-order mismatch with workflow text
The workflow says: "Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields."

`wc` always emits its counts in a fixed order regardless of flag order: **lines, words, bytes** (then filename). So `wc -wlc file.txt` prints `<lines> <words> <bytes> file.txt`, not `<words> <lines> <bytes>` as the flag order might suggest.

The Output section requires writing them in `words, lines, bytes` order. A naive reader following "first whitespace-delimited fields" in the order the flags are written will get the words and lines values swapped.

**Fix:** Either
- Spell out the field order explicitly: "`wc` prints fields in the fixed order lines, words, bytes — re-order them when writing the report," or
- Compute each count with a separate invocation: `wc -w`, `wc -l`, `wc -c`. More commands but no foot-gun.

#### H2. "byte count" vs `wc -c` on multibyte text — name vs reality
The skill correctly distinguishes "byte count" from "character count" in the *When to skip* section, which is good. But the report label `bytes: N` paired with `wc -c` is accurate only for ASCII; for UTF-8 with multibyte characters, `wc -c` returns bytes (correct per the skill's contract) while a user who skims `bytes: N` may still expect "characters."

**Fix:** No code change needed if the contract is "bytes." Consider one explicit sentence in the Output section: "`bytes` is byte length as reported by `wc -c`; for character counts use a Unicode-aware tool." This already lives in *When to skip* but is worth surfacing next to the output spec so downstream readers don't misinterpret the report.

#### H3. Validation: `file --mime-type` is portable but the `text/` prefix check rejects useful cases
`file --mime-type` returns things like:
- `text/plain`, `text/x-shellscript`, `text/markdown` — accepted (good)
- `application/json`, `application/xml`, `application/x-yaml` — **rejected**, even though these are text the user probably wants to count
- `inode/x-empty` for empty files — rejected, but `wc` handles empty files fine and the user may want to confirm a file is empty

**Fix:** Either broaden the accept list (`text/*` plus a small allowlist of `application/json`, `application/xml`, `application/yaml`, `application/x-yaml`, and explicitly accept `inode/x-empty`), or change the validation to "must not be detected as a known binary type" (blocklist style). At minimum, document the JSON/XML/YAML rejection so users aren't surprised.

### Medium priority

#### M1. Output path collision / re-run semantics on a non-skill file
The output is `<path>.count.md`. If the user runs the skill on `notes.count.md` (perhaps a previous report), the new output is `notes.count.md.count.md` — surprising but not broken. If the user runs the skill on a file that already happens to be named `something.count.md`, there's no detection that it's likely a generated artifact.

**Fix:** Optional. Either ignore (acceptable — it's a corner case), or add a sentence: "If `<path>` already ends in `.count.md`, warn the user that they may be counting a previously generated report."

#### M2. No handling of paths with embedded newlines / unusual filenames
The skill assumes shell-style `[ -f "$path" ]` and `wc -wlc "$path"`. Quoted properly in the doc, but if Claude expands `$path` without quoting on its own, paths with spaces or special chars will break. Worth being explicit.

**Fix:** Add a one-liner: "Always double-quote `$path` in shell commands."

#### M3. Symlink behavior is unspecified
`[ -f "$path" ]` follows symlinks (returns true for a symlink to a regular file). `wc` also follows symlinks. Probably the right default, but the skill doesn't say so. A user may expect a `[ -f "$path" ] && [ ! -L "$path" ]` style stricter check.

**Fix:** Add a sentence: "Symlinks to regular text files are accepted; the report is written next to the symlink, not the target."

#### M4. "regular file" claim vs `[ -f ]`
The Inputs section says "must exist and be a text file" with the test `[ -f "$path" ]`. `[ -f ]` is "regular file (after symlink resolution)," which is fine but doesn't distinguish text from binary — that's what the `file --mime-type` check is for. Wording is slightly redundant; not a bug.

**Fix:** Cosmetic. "Validation: file must exist (`[ -f "$path" ]`) **and** be a text file (test with `file --mime-type`)" — the existing wording already does this; the redundancy I flagged is mild.

### Low priority / polish

#### L1. `argument-hint` says `<path-to-file>` (singular), description says "a specific file or set of files"
The description says "a specific file or set of files" but the skill only handles one file per invocation. The argument hint correctly shows a single positional. This is a minor mismatch that could either be:
- Tighten description to "a single specific file," or
- Loosen the skill to iterate over multiple positional args.

**Fix:** Tighten the description; current behavior is single-file.

#### L2. Frontmatter completeness
Frontmatter has `name`, `description`, `argument-hint`. This is fine for a skill of this size. No `model`, `tools`, or `allowed-tools` keys — appropriate, since the skill only needs `Bash` and `Write`. If the skill were to be tightened against accidental tool sprawl, an `allowed-tools: [Bash, Write]` would be defensible.

**Fix:** Optional. Add `allowed-tools` if you want a tighter sandbox.

#### L3. No example invocation
The skill is small enough that an example is arguably overkill, but a one-line example like:
```
$ /word-count ./README.md
# writes ./README.md.count.md
```
would help a reader who skims.

**Fix:** Optional polish.

#### L4. Error-message contract is vague
"Exit with a one-line error to the user and do not write anything." Doesn't specify whether to use stderr, prefix with `error:`, exit code, etc. For a Claude skill (where output is conversational, not a CLI), "one-line error" is fine — but if anyone ever wraps this in a script, they'll wish there were an exit code contract.

**Fix:** Optional. Spell out the user-facing error format.

#### L5. Idempotence claim is correct but worth stress-testing
The skill claims "re-running on the same input file overwrites the report." That's true given the workflow as written. No issue; flagging only because idempotence is a load-bearing claim and should be preserved through any future changes.

#### L6. No mention of what to do if `wc` itself fails
`wc` rarely fails, but on a permission-denied or vanished-mid-read file it can. The skill validates existence up front but doesn't say to check `wc`'s exit code.

**Fix:** Optional. Add: "If `wc` fails, surface its stderr and don't write the report."

---

## What looks good

- Frontmatter is well-formed: `name`, `description`, `argument-hint` are all present and appropriate.
- The description's "Use when / Skip when" structure is exemplary — this is exactly the shape that helps Claude trigger the skill correctly.
- Scope is tight. The skill resists feature creep (frequency analysis, sentiment, etc.) and explicitly defers.
- Workflow is numbered, short, and testable.
- Output location (`<path>.count.md`) is predictable and idempotent.
- No surprising tool requirements; relies on standard `wc` and `file`.

## Recommended action order

1. Fix H1 (`wc` field-order foot-gun) — this is a real correctness bug as written.
2. Decide on H3 (validation list for non-`text/*` mime types). If the answer is "stay strict," document it so users aren't surprised.
3. Add a note for H2 (bytes-vs-characters at the output site).
4. Sweep the medium items as time allows.
5. Polish items (L*) are optional.

## Files reviewed

- `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-1/without_skill/work/skills/word-count/SKILL.md` (1760 bytes, 30 lines)

No other files were present in the skill directory.
