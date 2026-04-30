# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-5/eval-0/with_skill/work/skills/bad-skill/SKILL.md`
- Date: 2026-04-30
- Findings: 18  (idempotency: 4, reproducibility: 6, context-management: 1, strict-definitions: 7, security: 0)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses, so the model doesn't have to infer it from the workflow.
- `SKILL.md:14` — output `RELEASE_NOTES.md` is written without specifying overwrite/append behavior on re-runs.
- `SKILL.md:15` — `gh pr create` has no precondition check (e.g. `gh pr list --head $branch`) and no documented duplicate-acceptance; a second run will fail or open a duplicate PR.
- `SKILL.md:15` — "Push a tag" mutates remote state with no precondition (does the tag already exist?) and no documented overwrite intent.

### Reproducibility

- `SKILL.md:12` — "an appropriate release tag" gives no objective criterion; reproducibility requires a stated rule (semver bump rule, tag scheme, "next patch from latest tag").
- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input.
- `SKILL.md:13` — "summarize the changes as needed" gives no objective criterion; what determines what makes the cut versus what is dropped?
- `SKILL.md:13` — depends on git log / current git state without listing it as a declared input.
- `SKILL.md:26` — "reasonable in length" in the mandatory template gives the model no length target; runs will diverge.
- `SKILL.md:36,79,80` — "use judgment", "appropriately detailed", "Use your judgment if something looks off" give no test for when the directive applies; reproducibility requires a named threshold or rubric.

### Context management

- `SKILL.md:22` — the mandatory `RELEASE_NOTES.md` template is ~54 lines inline (lines 22–75); consider moving to `references/release-notes-template.md` and referencing it. The body of the skill does not need to load the template into context on every trigger.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (concrete user phrasings or contexts) — likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when the user just wants a CHANGELOG entry, or when no commits since last tag") — likely to over-trigger on near-misses.
- `SKILL.md:1` — no Inputs / Preconditions section. The skill consumes the current branch, the latest git tag, the git log range, and `gh` auth, but none are declared with source / required-ness / validation.
- `SKILL.md:15` — `gh pr create` requires the user to be authenticated to GitHub via `gh auth login`; this precondition is undocumented. Declare it in an Inputs / Preconditions section.
- `SKILL.md:14` — output `RELEASE_NOTES.md` has no documented path-conflict behavior (overwrite / append / refuse).
- `SKILL.md:16` — "Handle any errors that come up" is a vague verb with no concrete follow-up; specify what the model should do on each failure mode (e.g. tag already exists, PR already open, dirty working tree).
- `SKILL.md:81` — "If the repository has a CHANGELOG.md, also update it" introduces an unmentioned output and an undeclared step (not in the numbered Workflow). Either add a workflow step with a precondition, or remove the directive.

### Security

No findings.

## Passing checks

- Description names the artifact produced ("release notes") and the source ("from git history") — `SKILL.md:3`.
- The mandatory template uses placeholder syntax (`{{version}}`, `{{prose summary…}}`) rather than concrete-looking sample values, so no credentials would survive a generated file.
- No prompted secrets, no credential arguments, no URL-form connection strings — the skill stays outside the security objective entirely.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-5/eval-0/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
