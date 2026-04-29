# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-29
- Findings: 22  (idempotency: 4, reproducibility: 9, context-management: 1, strict-definitions: 8)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses (a release-notes generator that pushes tags has very different re-run semantics than one that only writes a file).
- `SKILL.md:14` — `git tag` / `gh pr create` are stateful external calls with no precondition (no `git tag -l` check, no `gh pr list --head`); a second run will fail or create a duplicate PR.
- `SKILL.md:14` — output path `RELEASE_NOTES.md` is written without specifying overwrite/append behavior.
- `SKILL.md:81` — "If the repository has a CHANGELOG.md, also update it" mutates a file without specifying append vs. overwrite vs. insert-at-top.

### Reproducibility

- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input; the same git log run on a different day produces a different "release tag".
- `SKILL.md:13` — depends on `git log` / current branch state without declaring it as an input or precondition (e.g. "must be on the release branch with the prior tag reachable").
- `SKILL.md:12` — "summarize the changes as needed" gives no objective criterion; reproducibility requires a stated test (length cap, sections required, etc.).
- `SKILL.md:12` — "figure out an appropriate release tag" leaves the versioning rule (semver? calver? bump rule?) to the model.
- `SKILL.md:20` — "fill in every section appropriately" — no rubric for what "appropriately" means per section.
- `SKILL.md:26` — "keep it engaging and reasonable in length" — no length bound; two runs will diverge.
- `SKILL.md:32` — "Use as much detail as relevant" / "Aim for around three to five sentences per feature, but use judgment based on the complexity of the change" — soft bound plus judgment escape hatch; pick one.
- `SKILL.md:79` — "Use your judgment if something looks off" — directive with no criterion for what "off" means or what to do about it.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" — no threshold; this is the canonical reproducibility leak.

### Context management

- `SKILL.md:22` — the inline release-notes template is ~54 lines and the placeholders (lines 32-42, 45-55, 58-71) carry multi-paragraph instructional prose, not literal copy-this content; consider moving to `assets/release-notes-template.md` and having SKILL.md reference it.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (no concrete user phrasings that should trigger); likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip if the user wants a CHANGELOG entry only, or if they want to draft a release without tagging"); likely to over-trigger on adjacent tasks.
- `SKILL.md:3` — description claims only "Generate release notes from git history" but workflow at line 14 also pushes a tag, creates a PR, and at line 81 updates CHANGELOG.md; either narrow the workflow or surface the side-effect class in the description so triggering decisions account for it.
- `SKILL.md:23` — input `{{version}}` is referenced in the template but the skill has no Inputs section declaring its source (CLI arg? prompt? derived from step 1?), required-ness, or validation (semver regex, etc.).
- `SKILL.md:14` — output `RELEASE_NOTES.md` has no documented behavior for a pre-existing file; same gap noted under idempotency, restated here because it is part of the output contract.
- `SKILL.md:14` — step 4 ("Push a tag and create a release PR") doesn't state its preconditions: that step 3 wrote the file, that no tag with this name already exists, that the working tree is clean.
- `SKILL.md:12` — step 1 ("figure out an appropriate release tag") doesn't state its precondition (a prior tag must be reachable from HEAD) or its output contract (the chosen tag must be passed to which later step?).
- `SKILL.md:16` — "Handle any errors that come up" — vague verb with no follow-up; say what "handle" means (stop and report? retry? roll back the tag?).

## Passing checks

- SKILL.md is well under the 500-line target (92 lines).
- Output path is named (`RELEASE_NOTES.md`) — the path itself is concrete even if its overwrite semantics aren't.
- Description names a verb and an artifact (the "what" element of trigger quality).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-0/with_skill/work/audit-bad-skill-2026-04-29.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
