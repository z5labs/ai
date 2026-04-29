# Audit Report: word-count Skill

**Target:** `./.claude/skills/word-count/`
**Files audited:** `SKILL.md` (29 lines, sole file)
**Date:** 2026-04-29

## Summary

The `word-count` skill is small, focused, and well-scoped. The metadata, workflow, and intent are coherent. There are several correctness concerns and a few minor documentation issues worth flagging. Nothing is catastrophically broken; the skill would likely succeed on the happy path, but edge cases could produce incorrect output silently.

---

## Strengths

1. **Tight scope.** The skill does one thing (three numbers, one report file). The description explicitly lists what to *skip* (sentiment, readability, frequency), which is a strong trigger-disambiguation pattern.
2. **Clear inputs/outputs contract.** A required positional `path`, validation rules, deterministic output location (`<path>.count.md`), and idempotency are all documented.
3. **Validation guidance is explicit.** Both existence (`[ -f "$path" ]`) and content-type (`file --mime-type`) checks are called out, with a fail-closed posture (do not write on error).
4. **Frontmatter is well-formed.** `name`, `description`, and `argument-hint` are all present and match Claude Code skill conventions.
5. **"When to skip" section.** Good practice — explicitly steers Claude away from misuse, complementing the description.

---

## Issues

### High severity

**H1. `wc -wlc` field-extraction instructions are ambiguous and likely wrong.**
Line 24 says: *"Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields."*
- `wc` always emits counts in a fixed order: **lines, words, bytes** — regardless of the flag order on the command line. So `wc -wlc file` outputs `<lines> <words> <bytes> <file>`, not `<words> <lines> <bytes>`.
- The instruction "capture the three numbers from the first whitespace-delimited fields" without specifying which number goes to which label invites swapping `words` and `lines` in the report. A naive implementation that mirrors the flag order (`-w -l -c` → words, lines, bytes) would write incorrect labels.
- **Fix:** Either use three single-flag invocations (`wc -w`, `wc -l`, `wc -c`) for unambiguous extraction, or document the actual `wc` output ordering (`lines words bytes`) explicitly.

**H2. Byte vs. character ambiguity in flags.**
The skill calls for `wc -c` (bytes), which is consistent with the documented "bytes" output and the "When to skip" note about character vs. byte counts. Good. But the workflow at step 2 says `wc -wlc` and the output spec says `bytes: N` — these are aligned only if `-c` is byte count. On GNU `wc`, `-c` is bytes and `-m` is characters; on some BSD locales `-c` can behave differently with multibyte input. Worth a one-line note pinning the expectation to GNU `wc -c` = bytes.

### Medium severity

**M1. MIME-type validation is too narrow.**
Line 15: *"reject if it doesn't start with `text/`"*. This rejects legitimate text-bearing MIME types like `application/json`, `application/xml`, `application/x-shellscript`, and `application/javascript`, which `file --mime-type` commonly returns. A user who asks "word-count this JSON file" gets a refusal that the description does not warn about.
- **Fix:** Either broaden the allowlist (`text/*` plus a curated set of `application/*` text formats), or document that only `text/*` MIME files are supported.

**M2. No guidance for files without trailing newlines.**
`wc -l` counts newline characters, not lines. A file with content but no trailing `\n` will report one fewer line than a user expects. Worth a sentence acknowledging this, or switch to a definition that matches user intuition.

**M3. Output format is markdown-extension but not markdown-formatted.**
The output file is `<path>.count.md` but its body is three `key: N` lines — that's YAML-ish, not markdown. It will render fine, but the `.md` extension implies structure. Either rename to `.count.txt`, or wrap in a code fence / use a real markdown table.

### Low severity

**L1. `argument-hint` shows one file but description mentions "a specific file or set of files".**
Line 3 says "a specific file **or set of files**", but the schema (line 15) says "a regular file" (singular) and `argument-hint` is `<path-to-file>` (singular). The skill does not support multiple files. Tighten the description to remove "or set of files" to avoid mis-triggering on multi-file requests.

**L2. Em-dash in description.**
Line 3 uses an em-dash ("— those need a different skill"). This is fine, but if the skill registry / matcher tokenizes on punctuation, an ASCII `--` is safer for portability. Cosmetic.

**L3. No example invocation.**
A short example (`/word-count README.md` → writes `README.md.count.md`) would help the model and the user. Optional but high-value for a 29-line skill.

**L4. Idempotency claim is partial.**
Line 11 says re-running with a different input "produces a separate report", which is true, but if the *same* output path is implied by two different inputs (impossible here since output is derived from input), there's no collision. The claim is correct but slightly over-stated; a single sentence would suffice.

**L5. No handling for paths with spaces or special characters in the documented commands.**
The skill quotes `"$path"` in the validation step but the workflow text describing `wc -wlc "$path"` is fine. However, the output path `<path>.count.md` is not shown quoted in any example; an implementer might forget. Minor.

---

## Recommended changes (priority order)

1. **Fix H1**: Rewrite step 2 to make field-to-label mapping unambiguous. Suggested:
   > Run `wc -l`, `wc -w`, and `wc -c` separately on `"$path"`, capturing each first integer.
2. **Fix M1**: Decide on MIME policy and document it. If `text/*`-only is intentional, say so in the description so users with JSON/XML aren't surprised.
3. **Tighten L1**: Remove "or set of files" from the description.
4. **Address M3**: Pick `.txt` or commit to real markdown formatting.
5. **Add an example** (L3).

---

## Verdict

**Ship-blocking?** No.
**Action recommended?** Yes — H1 is a real correctness bug in the documented procedure that will produce mislabeled reports if followed literally. Fix it before users rely on the output.

The skill is 80% there. The structural choices (scope, validation posture, idempotency, when-to-skip) are good. The remaining work is precision in the workflow steps and the MIME contract.
