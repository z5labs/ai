# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-0/old_skill/work/skills/bad-skill/SKILL.md`
- Date: 2026-04-30
- Findings: 22  (idempotency: 4, reproducibility: 9, context-management: 1, strict-definitions: 8, security: 0)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses, so the model doesn't have to infer it from the workflow.
- `SKILL.md:14` — step 3 writes `RELEASE_NOTES.md` without specifying overwrite/append/refuse behavior when the file already exists.
- `SKILL.md:15` — step 4 calls `git tag` (push a tag) and `gh pr create` — both stateful external calls — with no precondition check (e.g. `git tag -l`, `gh pr list --head $branch`) and no documented duplicate-handling intent. Re-running on the same commit will fail or create a duplicate PR.
- `SKILL.md:81` — note "if the repository has a CHANGELOG.md, also update it" mutates an existing file without specifying overwrite/append/prepend behavior.

### Reproducibility

- `SKILL.md:12` — step 1 says "figure out an appropriate release tag" with no objective criterion (semver bump rule? user-supplied? next-minor?). Two runs on the same history will produce two different tags.
- `SKILL.md:13` — "summarize the changes as needed" gives no rubric for length, grouping, or what counts as worth summarizing.
- `SKILL.md:20` — "fill in every section appropriately" — no test for what "appropriately" means per section.
- `SKILL.md:26` — template instructs "keep it engaging and reasonable in length" with no length bound or tone rubric.
- `SKILL.md:35` — "use judgment based on the complexity of the change" — the "three to five sentences" anchor is fine, but the "use judgment" override has no criterion.
- `SKILL.md:79` — "this skill works reasonably well for most cases. Use your judgment if something looks off" gives no objective criterion for either "reasonably well" or "looks off".
- `SKILL.md:80` — "release notes should be appropriately detailed — neither too sparse nor too verbose" names no threshold.
- `SKILL.md:12` — depends on the current date (`date`) without listing it as a declared input. Same prompt run on a different day produces a different tag.
- `SKILL.md:13` — depends on git state (`git log`, "since the last tag") without listing repository state, "last tag", or commit range as a declared input.

### Context management

- `SKILL.md:22` — inline release-notes template fenced block is 54 lines (lines 22–75), slightly over the ~50-line guideline; consider moving to `references/release-notes-template.md` and referencing it from `SKILL.md`. Soft finding — borderline by length, but the embedded prose instructions inside `{{...}}` placeholders are the bulk of the size and are pure reference content.

### Strict definitions

- `SKILL.md:3` — description "Generate release notes from git history." has no "when to use" examples — likely to under-trigger on phrasings like "draft a changelog", "cut a release", "tag and announce v1.2".
- `SKILL.md:3` — description has no "when to skip" / negative case — likely to over-trigger on near-misses like "summarize recent commits" or "what shipped this week" where the user doesn't want a tag pushed or a PR created.
- `SKILL.md:3` — description claims "generate release notes from git history" but the workflow at `SKILL.md:15` also pushes a git tag, creates a release PR, and optionally edits `CHANGELOG.md`; these state-mutating side effects are not surfaced in the description, so triggering decisions don't account for them. Either narrow the description ("draft release notes") and remove the side-effects, or extend the description to mention the tag/PR/changelog steps.
- `SKILL.md:8` — "Inputs" section is absent. The skill consumes at minimum: a release tag/version (source unstated — user-supplied? inferred?), the current branch, the last tag (assumed to exist), and presence of `CHANGELOG.md`. None are declared with source / required-ness / validation.
- `SKILL.md:18` — outputs section names `RELEASE_NOTES.md` but does not state path location relative to repo root, behavior on pre-existing file, or the path/format of the tag and release PR (the side-effects from step 4) which are also outputs.
- `SKILL.md:14` — step 3 writes `RELEASE_NOTES.md`; step 4 pushes a tag and creates a PR; the PR presumably references the file from step 3, but the dependency is implicit (no statement that the file must be committed before the tag/PR step).
- `SKILL.md:15` — "Handle any errors that come up" is a vague directive with no concrete follow-up — what counts as an error, whether to abort or retry, and what to report to the user are all left to the model.
- `SKILL.md:81` — "If the repository has a CHANGELOG.md, also update it" is an optional step with a predicate but no contract on what "update" means (append? prepend a new section? overwrite?) and no ordering relative to steps 3–5.

### Security

No findings.

## Passing checks

- Security — the skill relies on `gh` and `git`, which manage their own credentials independently of the model; no secrets routed through arguments, prompts, or generated files. Easy to lose if a future revision adds a "tag the release on GitHub via the API" path that handles a token directly.
- Output template is fully inline as a literal "do exactly this" template — appropriate use of inline content for a format the model must reproduce verbatim (would be wrong to move to `references/` for this specific case).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-0/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
