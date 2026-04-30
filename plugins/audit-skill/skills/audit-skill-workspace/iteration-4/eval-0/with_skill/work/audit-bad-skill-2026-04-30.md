# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-30
- Findings: 22  (idempotency: 4, reproducibility: 8, context-management: 1, strict-definitions: 8, security: 1)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites the prior `RELEASE_NOTES.md`, appends to it, or refuses.
- `SKILL.md:14` — output path `RELEASE_NOTES.md` is written without specifying overwrite/append behavior on re-run.
- `SKILL.md:15` — `gh pr create` is a stateful external call with no precondition (e.g. `gh pr list --head $branch`) and no doc that duplicates are expected. A second run will either fail or open a duplicate PR.
- `SKILL.md:15` — "Push a tag" mutates remote state with no check that the tag does not already exist; a re-run on the same version will fail or clobber.

### Reproducibility

- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion; reproducibility requires a stated test (semver bump rule, a regex, "next patch unless breaking changes").
- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input.
- `SKILL.md:13` — depends on `git log` (repository state) without listing it under Inputs / Preconditions.
- `SKILL.md:13` — "summarize the changes as needed" gives no objective criterion for what counts as "needed".
- `SKILL.md:20` — "fill in every section appropriately" gives no objective criterion for "appropriately".
- `SKILL.md:26` — "keep it engaging and reasonable in length" gives no objective criterion (no length bound, no tone rubric).
- `SKILL.md:35` — "use judgment based on the complexity of the change" gives no objective criterion; two runs on the same diff will produce different lengths.
- `SKILL.md:79` — "Use your judgment if something looks off" gives no objective criterion for what "off" means or what action to take.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" gives no objective criterion (no word/line/section bound).

### Context management

- `SKILL.md:22` — inline release-notes template runs lines 22–75 (54 lines) and embeds long instructional prose inside `{{…}}` placeholders rather than rules in the body; consider moving the template to `references/release-notes-template.md` and the per-section authoring rules to `references/section-rules.md`.

### Strict definitions

- `SKILL.md:3` — description "Generate release notes from git history." has no "when to use" examples (no concrete user phrasings or contexts) — likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case — likely to over-trigger on adjacent tasks (changelog edits, commit-message rewriting, PR descriptions).
- `SKILL.md:3` — description claims only "Generate release notes from git history" but the workflow at `SKILL.md:15` also pushes a tag and opens a release PR; either narrow the description or extend it to surface those side-effects so triggering decisions account for them.
- `SKILL.md:8` — no Inputs / Arguments / Preconditions section. The skill consumes the current date, the git log, and an implicit "last tag" reference, but none are declared with source / required-ness / validation.
- `SKILL.md:14` — output `RELEASE_NOTES.md` has a path but no documented format reference (the inline template), no statement of what happens when the file pre-exists, and no relative-path anchor (cwd? repo root?).
- `SKILL.md:13` — step 2 ("git log since the last tag") depends on a previous tag existing; precondition is implicit, not stated. Says nothing about a first release with no prior tag.
- `SKILL.md:15` — step 4 ("push a tag and create a release PR") doesn't state its preconditions (tag chosen at step 1? `RELEASE_NOTES.md` written and committed?) and bundles two distinct mutating actions into one numbered step.
- `SKILL.md:16` — "Handle any errors that come up." gives the model latitude with no concrete follow-up; specify the failure modes to look for and what to do (stop, retry, report which step failed).

### Security

- `SKILL.md:15` — `gh pr create` requires `GH_TOKEN` / `gh auth` state, but the skill doesn't document the credential precondition or what to do if auth is missing. Add a refuse-and-instruct: state that `gh auth status` must succeed before invocation, and exit with a clear error if it doesn't (so the model never falls back to prompting).

## Passing checks

- No model-prompted secrets, no credentials in arguments, no URL-form connection strings, no hardcoded secret values written to disk — the skill stays clear of the four leak shapes under security checks #1–#4.
- `SKILL.md` is 92 lines (well under the 500-line target).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-0/with_skill/work/audit-bad-skill-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
