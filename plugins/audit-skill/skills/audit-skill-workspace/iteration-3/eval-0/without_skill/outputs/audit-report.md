# Audit Report: `bad-skill` (Release Notes Generator)

**Target:** `./skills/bad-skill/SKILL.md`
**Audited:** 2026-04-29
**Verdict:** Needs significant revision before use. The skill mixes a benign "summarize git log" task with destructive side effects (pushing tags, creating PRs, modifying CHANGELOG.md), the description is too thin to trigger reliably, and the embedded template contains contradictory and over-prescriptive guidance that will produce inconsistent output.

---

## Summary of findings

| # | Severity | Area | Finding |
|---|----------|------|---------|
| 1 | High | Description / triggering | Description is a single sentence with no trigger guidance. |
| 2 | High | Scope | Skill performs destructive operations (`git push` of a tag, `gh pr create`) that go far beyond "generate release notes". |
| 3 | High | Workflow | Step 5 ("Handle any errors") is non-actionable. |
| 4 | High | Template | The mandatory template embeds ~15 directives inside a single `{{features}}` placeholder, mixing schema with prose guidance. |
| 5 | High | Internal contradictions | Length guidance contradicts itself ("engaging and reasonable in length" vs. "keep it short" vs. "three to five sentences"). |
| 6 | Medium | Determinism | "Use your judgment if something looks off" is a smell — skills should be deterministic where they can be. |
| 7 | Medium | Hidden side effects | CHANGELOG.md update is buried in Notes, not Workflow. |
| 8 | Medium | Inputs / preconditions | No documented preconditions (clean tree, branch, remote, prior tag). |
| 9 | Medium | Versioning | Step 1 leans on `date` to pick a tag. Tags are conventionally semver, not date-based; no policy is given. |
| 10 | Medium | Output path | `RELEASE_NOTES.md` is a relative path; behavior depends on cwd. |
| 11 | Low | Legacy noise | Note about old `notes.md` location adds confusion without value. |
| 12 | Low | Examples | Examples describe section shape only; they don't show actual filled content. |
| 13 | Low | Naming | The skill name `bad-skill` is presumably a placeholder; the `name:` field in frontmatter must match the directory and should describe the function. |

---

## Detailed findings

### 1. Description is too thin to trigger reliably (High)

```yaml
description: Generate release notes from git history.
```

A description of this size won't tell Claude *when* to invoke the skill versus when to skip. Compare with the example skills in this repo (`word-count`, `simplify`, etc.), which name triggers, file types, and skip conditions. Concretely the description should answer: *what user phrases should fire this?* (e.g., "draft release notes", "prep a release", "summarize what's new since vX") and *when should it skip?* (e.g., "not for changelog-only edits", "not when the user just wants a git log summary").

### 2. Scope creep — destructive side effects (High)

The skill is named "generate release notes" but Workflow steps 4 actually:

- runs `git tag` and pushes it (implied by "push a tag"),
- runs `gh pr create` to open a PR.

These are irreversible from inside the skill's perspective (a pushed tag is not a local-only artifact; a PR notifies reviewers). A skill that *generates* a document should not silently *publish* anything. Recommend:

- Split into two skills, or
- Make publishing an explicit, opt-in step gated on user confirmation, with the document-generation step working purely on the local working tree.

### 3. Step 5 is not a workflow step (High)

> 5. Handle any errors that come up.

This is a non-instruction. It does not specify what failure modes are expected (e.g., no prior tag, dirty working tree, `gh` not authenticated, push rejected) or what to do for each. Either remove it or replace with a concrete error-handling table.

### 4. Template overloads `{{...}}` placeholders with directives (High)

The Features placeholder is a single `{{...}}` block containing roughly 15 distinct directives:

- write a paragraph per feature, explain *what / why / how to invoke*
- use as much detail as relevant
- nest subfeatures
- call out replacements / deprecations
- friendly-but-precise tone
- avoid jargon but don't dumb things down
- aim for 3–5 sentences but use judgment
- cross-reference issues / PRs with `#1234`
- tag external contributors with `@username`
- flag experimental / flag-gated features prominently
- summarize migration paths and link to full guides
- examples can help but keep short
- mention caveats but don't dwell

This conflates the *schema* of the section (what fields exist) with *style guidance* (how to write them). When Claude renders the template, it cannot tell where the schema ends and the directive ends — the result is template syntax leaking into output, or directives being silently ignored. Recommend separating into:

- a clean schema/example: literal text showing what the rendered Features section looks like;
- a separate `## Style guide` section in the SKILL.md with the directives as a numbered list.

The Fixes and Breaking Changes placeholders have the same problem.

### 5. Internal contradictions on length (High)

Within the same template:

- Highlights: "keep it engaging and reasonable in length"
- Features: "Aim for around three to five sentences per feature, but use judgment"
- Features (later): "keep them short — long code samples belong in the docs"
- Fixes: "one to two sentences describing the bug"
- Fixes (later): "Keep fix descriptions terse — users skim this section"

Some of this is reconcilable but not as written. Pick concrete, non-overlapping length budgets per section and remove the hedges ("reasonable", "use your judgment", "as much detail as relevant").

### 6. "Use your judgment if something looks off" (Medium)

> This skill works reasonably well for most cases. Use your judgment if something looks off.

This is a smell. A skill is supposed to be the codified judgment. If there are known edge cases ("if there's no prior tag", "if the diff includes only docs changes"), they should be enumerated with explicit handling.

### 7. Hidden side effect: CHANGELOG.md (Medium)

> If the repository has a CHANGELOG.md, also update it.

This is a second file edit, not mentioned in the Workflow or the Output section. Either:

- promote it to a Workflow step with concrete instructions on *how* to update it (prepend? insert under "Unreleased"? format?), or
- remove it.

Currently a user reading the Workflow won't know CHANGELOG.md will also be touched.

### 8. No documented preconditions (Medium)

The skill assumes — but does not state — that:

- the working tree is clean,
- a prior tag exists ("git log since the last tag"),
- a remote named `origin` is configured and authenticated,
- `gh` is authenticated,
- the user wants to release the *current* branch.

If any are false the skill will fail mid-run, possibly after writing a partial RELEASE_NOTES.md. Add a Prerequisites section that the skill checks first.

### 9. Versioning policy missing (Medium)

> Look at the current date with `date` and figure out an appropriate release tag.

This conflates two unrelated things. `date` produces a date string; release tags in most repos follow semver (`vX.Y.Z`) derived from the prior tag plus the change set. The skill should:

- read the latest tag (`git describe --tags --abbrev=0`),
- inspect commit messages / labels to choose major/minor/patch,
- only fall back to date-based tagging if the project uses CalVer, and the skill should detect that.

### 10. Relative output path (Medium)

> Write the release notes to `RELEASE_NOTES.md`.

This is relative to the cwd at invocation time. If Claude is in a subdirectory the file lands in the wrong place. Specify "the repository root" and resolve via `git rev-parse --show-toplevel`.

### 11. Legacy migration note adds noise (Low)

> Older versions of this skill wrote to `notes.md` instead of `RELEASE_NOTES.md`. The new path is correct.

Skill files are not a changelog. Drop this line — it gives Claude two paths to consider and no useful disambiguator.

### 12. Examples are placeholders (Low)

Both examples describe what the *sections* look like, not what an actual rendered release-notes file looks like. They give Claude no concrete pattern to imitate. Replace with one or two short, real-looking rendered examples (~30–60 lines each).

### 13. Skill name (Low)

`name: bad-skill` is presumably a placeholder, but worth flagging:

- the `name:` field must match the directory name,
- it should describe the function (e.g., `release-notes`).

---

## Recommended next steps

In priority order:

1. Split publishing (tag/push/PR) out of this skill or gate it behind explicit user confirmation.
2. Rewrite the description to include trigger phrases and skip conditions.
3. Restructure the template: clean schema + literal example, with style guidance moved to a separate `## Style guide` section.
4. Remove length contradictions; pick one budget per section.
5. Add a Prerequisites section with explicit precondition checks.
6. Replace step 5 ("handle errors") with a concrete error-handling table or remove it.
7. Either fully document the CHANGELOG.md update or drop it.
8. Resolve the output path against `git rev-parse --show-toplevel`.
9. Define the versioning policy (semver vs. CalVer; how to pick the bump).
10. Rename the skill (`name:` and directory) to its real name.
11. Remove the legacy `notes.md` note.
12. Replace the abstract examples with one or two concrete rendered ones.
