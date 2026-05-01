# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-0/old_skill/work/skills/bad-skill`
- Date: 2026-04-30
- Findings: 22  (idempotency: 4, reproducibility: 9, context-management: 1, strict-definitions: 8, security: 0)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses.
- `SKILL.md:14` — output path `RELEASE_NOTES.md` is written without specifying overwrite/append behavior when the file already exists.
- `SKILL.md:15` — `git tag` and `gh pr create` are stateful external calls with no precondition check (e.g. `gh pr list --head $branch`) and no documented duplicate-acceptance.
- `SKILL.md:81` — "also update CHANGELOG.md" is a state mutation with no precondition (existence test), no overwrite/append intent, and no schema for the update.

### Reproducibility

- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input; the skill's chosen tag will vary by invocation date.
- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion (semver bump rule? date-based? read from a config?); reproducibility requires a stated test.
- `SKILL.md:13` — depends on `git log` / current branch state without listing it as a declared input or naming the "last tag" resolution rule.
- `SKILL.md:13` — "summarize the changes as needed" — "as needed" gives no rubric; runs will diverge in length and emphasis.
- `SKILL.md:20` — "fill in every section appropriately" gives no objective criterion for what counts as filled.
- `SKILL.md:26` — "engaging and reasonable in length" — no length threshold or tone rubric stated.
- `SKILL.md:36` — "use judgment based on the complexity of the change" delegates the rubric to the model with no anchor.
- `SKILL.md:79` — "works reasonably well for most cases. Use your judgment if something looks off" — pure hedge with no criterion.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" gives no length / density threshold; two runs will land at different lengths.

### Context management

- `SKILL.md:22` — the inline release-notes template is ~53 lines of prose-laden placeholder instructions (lines 22–75); move to `references/release-notes-template.md` (or `assets/`) and reference it from the workflow. The placeholders are paragraphs of guidance, not literal "copy this exactly" content, so progressive disclosure applies.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (no concrete user phrasings or contexts that should trigger); likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. when CHANGELOG-only updates suffice, when a maintainer is using a different release tool); likely to over-trigger on adjacent tasks.
- `SKILL.md:3` — description claims only "Generate release notes from git history" but the workflow at lines 15 and 81 also pushes a git tag, opens a release PR, and updates `CHANGELOG.md`; either narrow the workflow or surface the side-effect class in the description so triggering decisions account for it.
- `SKILL.md:1` — no Inputs / Arguments / Preconditions section; the skill consumes the current branch, the last tag, the working tree's git log, and optionally a `CHANGELOG.md`, none of which are declared with source / required-ness / validation.
- `SKILL.md:14` — output `RELEASE_NOTES.md` has no documented path semantics (project root? cwd?), no behavior on pre-existing file, and no validation that every templated section was filled.
- `SKILL.md:81` — output "update CHANGELOG.md" has no documented format, insertion point, or pre-existence behavior.
- `SKILL.md:16` — step 5 "Handle any errors that come up" is a vague verb with no concrete follow-up (which errors? stop? roll back the tag? close the PR?).
- `SKILL.md:11` — workflow steps 1–5 are presented as a flat numbered list but step 4 (`git tag` push and `gh pr create`) has implicit preconditions on step 3 (the file must exist) and on a clean working tree; state preconditions per step.

### Security

No findings.

## Passing checks

- Security — no model-prompted secrets, no credentialed arguments, no discard-after-read patterns, and no concrete credentials baked into generated files. The skill calls `gh pr create`, which manages its own credentials independently of the model.
- The skill names a concrete output filename (`RELEASE_NOTES.md`) — not a perfect output declaration (see strict-definitions findings) but the path is at least named, which is easy to lose in a rewrite that moves to "an appropriately named file".

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-0/old_skill/work/audit-bad-skill-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
