# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-3/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-29
- Findings: 19  (idempotency: 4, reproducibility: 7, context-management: 1, strict-definitions: 7)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses, so the model doesn't have to infer it from the workflow.
- `SKILL.md:14` — `RELEASE_NOTES.md` is written without specifying overwrite/append behavior on a pre-existing file. The note about an older `notes.md` path does not address re-runs at the new path.
- `SKILL.md:15` — `git tag` (push a tag) is a stateful operation with no precondition (no check that the tag does not already exist) and no documented intent for re-runs.
- `SKILL.md:15` — `gh pr create` is a stateful external call with no precondition (`gh pr list --head $branch`) and no doc that duplicates are acceptable; re-running creates a duplicate PR.

### Reproducibility

- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion (semver bump rule, format, source of truth); two runs may pick different tags.
- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input.
- `SKILL.md:13` — depends on `git log` / git state without listing it as a declared input under Inputs / Preconditions.
- `SKILL.md:13` — "summarize the changes as needed" gives no objective criterion for what to include or omit.
- `SKILL.md:16` — "Handle any errors that come up" has no rule for which errors are recoverable, which abort, or how to report them.
- `SKILL.md:79` — "Use your judgment if something looks off" gives no test for "looks off"; behavior is unpredictable across runs.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" gives no threshold (line count, section count, word count) for what counts as either extreme.

### Context management

- `SKILL.md:22` — the inline release-notes template is ~54 lines (lines 22–75) and the `{{Features}}` and `{{Breaking changes}}` placeholders contain ~15 lines of instructional prose each that read as reference material rather than a literal template the model copies verbatim. Consider moving the long placeholder guidance to `references/release-notes-format.md` and keeping only the section skeleton inline.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (concrete user phrasings or contexts that should trigger); likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip if the user just wants a changelog summary, not a release"); likely to over-trigger on near-misses.
- `SKILL.md:3` — description claims only "Generate release notes from git history" but the workflow at lines 14–15 also writes `RELEASE_NOTES.md`, pushes a git tag, and opens a PR; either narrow the workflow or surface those side-effects in the description so triggering decisions account for them.
- `SKILL.md:12` — input "release tag" is referenced ("figure out an appropriate release tag") but its source (CLI arg? prior tag + bump rule?), required-ness, and validation (semver regex?) are not stated. No Inputs section exists.
- `SKILL.md:16` — step 5 "Handle any errors that come up" is a vague imperative with no concrete follow-up; matches the `(handle|process|deal with|...)` pattern.
- `SKILL.md:81` — second output `CHANGELOG.md` ("also update") has no documented path location, format, insertion point, or behavior on missing/existing file.
- `SKILL.md:14` — workflow steps are presented as a flat list but step 4 ("Push a tag and create a release PR") has implicit preconditions on step 3 (RELEASE_NOTES.md must exist; presumably committed) that are not stated; tag-then-PR ordering relative to the commit is also unspecified.

## Passing checks

- SKILL.md is well under the 500-line context budget (92 lines).
- Output path for the primary artifact (`RELEASE_NOTES.md`) is named explicitly at line 20.
- The note at line 82 about the renamed output path (`notes.md` → `RELEASE_NOTES.md`) is the kind of historical pin worth keeping in a revision.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
