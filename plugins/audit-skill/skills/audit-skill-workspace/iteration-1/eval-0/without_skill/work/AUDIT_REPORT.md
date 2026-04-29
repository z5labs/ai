# Audit Report: `bad-skill` (release-notes generator)

**Target:** `./skills/bad-skill/SKILL.md`
**Date:** 2026-04-29
**Verdict:** Not ready for use. Multiple high-severity correctness, safety, and design issues. Recommend rewriting before merging.

---

## Summary

The skill ostensibly generates release notes from git history. It has the right *outline* (frontmatter, workflow, output format, examples), but the substance has serious problems:

1. It mixes "draft a markdown file" with destructive, network-effecting actions (tag push, PR creation) without authorization, dry-run, or rollback guardrails.
2. The frontmatter `description` is too thin to drive reliable model invocation.
3. The output template embeds prose *instructions* inside `{{...}}` placeholders that the model is supposed to follow while filling in — a confused pattern that conflates schema with prompt.
4. The workflow has gaps (no "last tag" lookup, no monorepo handling, no empty-history handling) and contains dead/contradictory guidance.
5. Examples are too vague to be useful for grounding.
6. Stale-version commentary leaks legacy detail into the live skill.

Each issue is detailed below with severity and a concrete fix.

---

## Issue 1 — Skill performs destructive actions silently. **(Severity: critical)**

> Workflow step 4: "Push a tag and create a release PR with `gh pr create`."

The skill's *stated* purpose is generating a release-notes document. Pushing a tag (`git push --tags` or `git tag ... && git push`) and opening a PR are side-effects on the remote. A user who invokes "generate release notes" almost certainly does not expect:

- a new annotated/lightweight tag created and pushed to `origin`,
- a PR opened on GitHub,
- both actions taken without confirmation.

This is the single biggest problem. Tags are effectively immutable in shared repos (deleting a pushed tag breaks downstream caches, mirrors, package registries). PRs notify reviewers and trigger CI.

**Fix:**
- Drop tag-push and PR-creation from the default workflow entirely, or
- Gate them behind explicit user confirmation per invocation, with a dry-run preview of the tag name and PR title/body, and
- Default to producing only the local `RELEASE_NOTES.md` file.

## Issue 2 — Tag selection is hand-wavy. **(Severity: high)**

> Step 1: "Look at the current date with `date` and figure out an appropriate release tag."

"Figure out an appropriate release tag" delegates a *versioning policy decision* to the model with no rules. The repo's actual versioning scheme (semver? CalVer? per-package monorepo tags?) is unknowable from `date` alone. Calling `date` is also irrelevant unless CalVer is in use.

**Fix:** Replace with a concrete recipe — e.g. "Read the latest tag with `git describe --tags --abbrev=0`, infer scheme from existing tags, and propose the next version. Ask the user to confirm before proceeding."

## Issue 3 — "git log since the last tag" is underspecified. **(Severity: high)**

> Step 2: "Read the git log since the last tag and summarize the changes as needed."

There is no command given, no handling for:
- repos with zero existing tags (first release),
- monorepos with multiple tag prefixes (`pkg-a/v1.2.0` vs `pkg-b/v0.4.1`),
- merges vs squashed commits,
- conventional-commit parsing if used,
- excluding commits already shipped on a release branch.

"Summarize as needed" is a non-instruction.

**Fix:** Provide the actual command (`git log <last-tag>..HEAD --no-merges --pretty=...`) and rules for grouping commits into Features / Fixes / Breaking.

## Issue 4 — Template placeholders contain prompt text, not schema. **(Severity: high)**

The `{{...}}` blocks in the output template are paragraphs of writing guidance ("Aim for around three to five sentences per feature... Cross-reference related issues... If a feature has a known caveat, mention it but don't dwell..."). This is a mash-up of:

- a Mustache-like fillable template, and
- inline editorial direction.

Two problems:

1. The model may literally copy the placeholder prose into the output if it reads the template at face value.
2. Guidance buried inside placeholders is much harder to revise or test than guidance in a dedicated section.

**Fix:** Move all editorial guidance to a top-level "Style guide" section. Keep placeholders short and structural — `{{highlights}}`, `{{features[]}}`, `{{breaking_changes[]}}`.

## Issue 5 — Frontmatter `description` is too generic. **(Severity: high)**

```
description: Generate release notes from git history.
```

Skills are selected by description. This one omits the trigger surface (when should the model invoke it?), the deliverable (a `RELEASE_NOTES.md`?), and the pre-conditions (e.g. "use when the user asks for release notes, a changelog entry, or a draft for a tag").

**Fix:** Expand to include trigger phrases and the concrete output. Example: "Use when the user asks to draft release notes, a changelog entry, or a summary of changes since the last tag. Produces a `RELEASE_NOTES.md` file in the repo root."

## Issue 6 — Conflicting / contradictory instructions. **(Severity: medium)**

- The skill is named "bad-skill" but its `name:` field is `bad-skill` and the heading is `# bad-skill`. If this is meant for production it needs a real name; if it's a placeholder, that should be flagged at the top.
- Step 5: "Handle any errors that come up." — no-op instruction. Either enumerate the failure modes (no tags, dirty working tree, no `gh` auth, network failure on push) and what to do, or remove the line.
- Notes section: "If the repository has a CHANGELOG.md, also update it." This contradicts the stated Output ("Write the release notes to `RELEASE_NOTES.md`") — there's no template for CHANGELOG.md, no rule for where to insert the new entry, and no mention of CHANGELOG anywhere else in the skill.
- Notes section: "This skill works reasonably well for most cases. Use your judgment if something looks off." This is editorial reassurance directed at the reader, not actionable instruction for the model.

**Fix:** Remove the reassurance line, drop or fully spec the CHANGELOG branch, and replace the error-handling line with concrete failure-mode handling.

## Issue 7 — Stale legacy reference leaked into live doc. **(Severity: medium)**

> "Older versions of this skill wrote to `notes.md` instead of `RELEASE_NOTES.md`. The new path is correct."

This belongs in a CHANGELOG for the skill itself, not in the skill body. It clutters the prompt and gives the model a chance to second-guess the output path.

**Fix:** Delete the line. Keep skill-version history in commit messages or a separate file.

## Issue 8 — Examples are too thin to ground behavior. **(Severity: medium)**

Example 1 says output is "a release notes file with all sections filled" with a Features section reading "No user-facing features in this release." Example 2 says "a release notes file with detailed Highlights, Features, and Breaking Changes sections."

These descriptions don't show the expected *shape* of the output — they paraphrase it. Useful examples would be a short, complete sample release-notes document (one for the bug-fix-only case, one for a major release) so the model has a concrete target.

**Fix:** Replace with two short, fully-rendered example outputs.

## Issue 9 — "Mandatory" template undermines its own utility. **(Severity: low)**

> "The format below is mandatory — fill in every section appropriately"

For a bug-fix-only release, "fill in every section" forces stub content into Features, Breaking changes, and Contributors. The skill correctly hints at the workaround in Example 1 ("No user-facing features in this release."), but a "mandatory" template combined with "fill in every section" plus per-section instructions to omit empty content creates conflicting pressure.

**Fix:** Make sections optional with explicit rules: "Omit Breaking changes entirely if there are none. Use 'No user-facing features in this release.' for an empty Features section."

## Issue 10 — No guidance on where the file goes or what to do if one exists. **(Severity: low)**

The skill says "Write the release notes to `RELEASE_NOTES.md`." It does not say:
- repo root vs a `releases/` directory vs a per-version filename (`RELEASE_NOTES_v1.2.3.md`),
- what to do if the file already exists (overwrite? append? abort?),
- whether to commit the file.

**Fix:** Specify the path, the collision behavior, and whether the skill commits.

## Issue 11 — No guidance on author/contributor extraction. **(Severity: low)**

The Contributors section says "auto-generate from git log; thank external contributors first, then maintainers." There's no definition of "external" vs "maintainer" — the skill doesn't know who's on the team. `git shortlog -sn` gives names; the external/maintainer split needs either a CODEOWNERS lookup, a maintainers list, or a rule to drop the distinction.

**Fix:** Either drop the external/maintainer split or specify the source of truth (e.g. "Treat anyone listed in `MAINTAINERS.md` as a maintainer; everyone else is external").

---

## Recommended priority order for fixes

1. **Issue 1 (critical):** Remove or gate destructive actions. Until this is fixed the skill should not be used.
2. **Issues 2, 3, 4, 5 (high):** Make tag selection concrete, give a real `git log` recipe, separate template from style guide, strengthen the description.
3. **Issues 6, 7, 8 (medium):** Remove contradictions, drop legacy reference, write real example outputs.
4. **Issues 9, 10, 11 (low):** Tighten collision/path/contributor rules.

## What's good

- The skill follows the standard Claude Code skill layout (frontmatter + workflow + output spec + examples).
- The release-notes outline (Highlights / Features / Fixes / Breaking changes / Contributors) is conventional and reasonable.
- Including examples at all is the right instinct, even if the current ones are too thin.

---

## File audited

`/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1/eval-0/without_skill/work/skills/bad-skill/SKILL.md`
