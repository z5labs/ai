# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-29
- Findings: 22  (idempotency: 5, reproducibility: 9, context-management: 1, strict-definitions: 7)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses, so the model doesn't have to infer it from the workflow.
- `SKILL.md:14` — output path `RELEASE_NOTES.md` is written without specifying overwrite/append behavior on re-run.
- `SKILL.md:15` — `git tag` (implied by "Push a tag") mutates state without a precondition; tagging the same commit twice fails and there is no "if tag exists" branch.
- `SKILL.md:15` — `gh pr create` runs without a precondition (`gh pr list --head $branch`) or a documented "duplicates acceptable" stance; a second run will create a duplicate PR or error.
- `SKILL.md:81` — `CHANGELOG.md` update is described without specifying append-vs-overwrite behavior or what happens if the file is missing.

### Reproducibility

- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input; same prompt on a different day yields a different release tag.
- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion; reproducibility requires a stated test (semver bump rule, derive-from-commits regex, or explicit user input).
- `SKILL.md:13` — depends on `git log` / repo state without listing it as a declared input; results vary by checkout.
- `SKILL.md:13` — "summarize the changes as needed" gives no objective criterion for what to include or omit.
- `SKILL.md:20` — "fill in every section appropriately" gives no rubric for what "appropriate" content per section means.
- `SKILL.md:26` — "engaging and reasonable in length" gives no length bound or tone rubric; runs will diverge in voice and size.
- `SKILL.md:35` — "use judgment based on the complexity of the change" leaves the per-feature length entirely to the model.
- `SKILL.md:74` — auto-generates Contributors from `git log` without listing git state as a declared input.
- `SKILL.md:79` — "works reasonably well for most cases. Use your judgment if something looks off" gives the model latitude with no rubric for what "off" looks like or what to do about it.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" gives no length or content threshold.
- `SKILL.md:87` — Example 1 output "a Features section with 'No user-facing features in this release.'" contradicts the template at line 28, which describes Features as "bulleted list of new features" with no provision for an empty-list sentinel.

### Context management

- `SKILL.md:22` — the inline release-notes template is 54 lines (lines 22–75) and the placeholder bodies at lines 28–42, 44–55, and 57–71 are themselves ~12-line prose specs of *how to write each section*. Consider moving the template to `references/release-notes-template.md` so the per-section guidance only loads when the skill actually fires.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (concrete user phrasings or contexts) — likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when the user wants a CHANGELOG entry, not release notes") — likely to over-trigger on near-misses.
- `SKILL.md:1` — no Inputs / Arguments section. The skill consumes current date, git log range, and target branch; none are declared with name, source, required-ness, or validation.
- `SKILL.md:14` — output `RELEASE_NOTES.md` has no documented behavior on pre-existing file (overwrite / append / error).
- `SKILL.md:81` — output `CHANGELOG.md` has no documented path location, format, or pre-existing-file behavior.
- `SKILL.md:15` — step 4 bundles two distinct stateful actions ("Push a tag" and "create a release PR") into one bullet and states no precondition (clean tree? CI green? tag computed in step 1 must be passed in?).
- `SKILL.md:16` — "Handle any errors that come up" is a vague verb with no concrete follow-up; specify which errors, how to recover, when to abort.

## Passing checks

- SKILL.md is 93 lines — well under the 500-line target.
- Output path is named (`RELEASE_NOTES.md`) rather than left implicit.
- The template uses `{{placeholder}}` syntax consistently, which is easy for the model to parse.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1/eval-0/with_skill/work/audit-bad-skill-2026-04-29.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
