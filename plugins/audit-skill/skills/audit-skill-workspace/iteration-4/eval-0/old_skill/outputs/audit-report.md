# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-0/old_skill/work/skills/bad-skill`
- Date: 2026-04-30
- Findings: 22  (idempotency: 5, reproducibility: 8, context-management: 1, strict-definitions: 8)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses.
- `SKILL.md:14` — `RELEASE_NOTES.md` is written without specifying overwrite/append behavior when the destination already exists.
- `SKILL.md:15` — `git tag` / "Push a tag" mutates state without a precondition (e.g. checking the tag doesn't already exist) or documented overwrite intent.
- `SKILL.md:15` — `gh pr create` is stateful and not naturally idempotent; no `gh pr list --head $branch` precondition and no documented "duplicates are acceptable" stance.
- `SKILL.md:81` — "If the repository has a CHANGELOG.md, also update it" mutates a file without specifying append vs. overwrite vs. insert-at-top behavior.

### Reproducibility

- `SKILL.md:12` — `date` is read from the environment but is not listed under any Inputs / Preconditions section; the same prompt run on a different day will produce different output.
- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion; reproducibility requires a stated rule (semver bump policy, regex for the tag format, or a named source for the version).
- `SKILL.md:12` — "summarize the changes as needed" — "as needed" gives no test for when a change gets a line vs. is dropped.
- `SKILL.md:13` — `git log` is read but not declared as an input (which range, which format); reads against a different cwd or branch silently change the output.
- `SKILL.md:26` — "keep it engaging and reasonable in length" — "reasonable" gives no length criterion (sentences, words, lines).
- `SKILL.md:36` — "Aim for around three to five sentences per feature, but use judgment based on the complexity of the change" — the named range is fine, but "use judgment based on the complexity" gives the model latitude with no rubric.
- `SKILL.md:79` — "This skill works reasonably well for most cases. Use your judgment if something looks off" — no criterion for what "off" means or what to do when judgment is invoked.
- `SKILL.md:80` — "should be appropriately detailed — neither too sparse nor too verbose" — no objective threshold; two runs will diverge.

### Context management

- `SKILL.md:22` — the release-notes template is ~54 lines of fenced inline content (lines 22–75); consider moving to `assets/release-notes-template.md` and referencing it, so the template only loads when step 3 runs.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (no concrete user phrasings or contexts that should trigger); likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when there is no prior tag", "skip for non-release branches"); likely to over-trigger on near-misses.
- `SKILL.md:3` — description claims "Generate release notes from git history" but the workflow at `SKILL.md:15` also pushes a tag and creates a PR; either narrow the description's scope or surface the state-mutating side-effects in it so triggering decisions account for them.
- `SKILL.md:1` — no Inputs / Arguments / Preconditions section: the skill consumes a version/tag, the current branch, the last tag, and the presence of `CHANGELOG.md`, none of which are declared with source / required-ness / validation.
- `SKILL.md:14` — output `RELEASE_NOTES.md` has a path but no documented behavior for a pre-existing file (overwrite / append / error).
- `SKILL.md:81` — `CHANGELOG.md` is a second output mentioned only as a parenthetical note; it has no documented path-format rule and is not surfaced in the Output section.
- `SKILL.md:15` — step 4 bundles two state-mutating actions ("push a tag" AND "create a release PR") into one numbered step, with no preconditions stated for either; split the steps and state what must be true before each runs (tag does not exist; no open release PR for this version).
- `SKILL.md:17` — step 5 "Handle any errors that come up" is a vague-verb instruction with no concrete follow-up; specify which errors (push rejected, tag exists, PR already open) and what to do for each.

## Passing checks

- SKILL.md is 93 lines, well under the 500-line target — context-management size budget is healthy.
- Description names the artifact produced ("release notes") and a verb ("Generate"), so the "what it does" element of the trigger contract is present.
- The inline template (lines 22–75) is concrete and prescriptive about section structure — the model has a literal shape to match rather than a vague request for "release notes".

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-0/old_skill/work/audit-bad-skill-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
