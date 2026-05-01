# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-30
- Findings: 23  (idempotency: 4, reproducibility: 10, context-management: 1, strict-definitions: 8, security: 0)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses.
- `SKILL.md:14` — `RELEASE_NOTES.md` is written without specifying overwrite/append behavior when the file already exists.
- `SKILL.md:15` — `git tag` push and `gh pr create` are stateful external calls with no precondition (no `gh pr list --head $branch` check, no policy on existing tag, no documented "duplicates acceptable").
- `SKILL.md:81` — instructing the workflow to "also update `CHANGELOG.md`" does not state whether the update appends a new section, replaces the file, or refuses if a section for this version already exists.

### Reproducibility

- `SKILL.md:12` — depends on `date` (current calendar date) to derive a release tag without listing it as a declared input; the same prompt run on a different day yields a different tag.
- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion (semver bump rule? read VERSION file? infer from commit messages?); runs will diverge.
- `SKILL.md:13` — depends on `git log` and on the current branch's commit history without listing it as a declared input or stating what "since the last tag" means when no tag exists.
- `SKILL.md:13` — "summarize the changes as needed" gives no criterion for what summarization is required.
- `SKILL.md:26` — "keep it engaging and reasonable in length" — no length threshold or tone rubric.
- `SKILL.md:33` — "as much detail as relevant" and "use judgment based on the complexity of the change" leave the model to guess.
- `SKILL.md:36` — "Cross-reference related issues and PRs where appropriate" — no rule for what makes a cross-reference appropriate.
- `SKILL.md:79` — "This skill works reasonably well for most cases. Use your judgment if something looks off." asks for judgment without a rubric.
- `SKILL.md:80` — "appropriately detailed — neither too sparse nor too verbose" gives no objective test.
- `SKILL.md:88` — Example 1 output specifies a Features section reading `"No user-facing features in this release."`; the rule at line 28 mandates a bulleted list shape for that section. Example contradicts the rule.

### Context management

- `SKILL.md:22` — the inline release-notes template (lines 22-75, ~53 lines) is mostly per-section prose guidance rather than a literal template the model must copy verbatim; consider moving the per-section guidance to `references/release-notes-template.md` and keeping only the skeleton inline.

### Strict definitions

- `SKILL.md:3` — description has no "when to use" examples (concrete user phrasings or contexts that should trigger); likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when the user wants a CHANGELOG entry only" or "skip when no commits since the last tag"); likely to over-trigger.
- `SKILL.md:3` — description says "from git history" but the workflow also pushes a tag and creates a PR (line 15) and updates CHANGELOG.md (line 81); the description should surface those side-effect classes so triggering decisions account for them.
- `SKILL.md:3` — no Inputs / Arguments section and no `argument-hint`; the model has no declared way to know what is supplied (target version? base branch? scope?).
- `SKILL.md:20` — outputs are partially documented (`RELEASE_NOTES.md` named, `CHANGELOG.md` only mentioned at line 81 inside a Notes bullet); add an Outputs section listing each artifact's path, format, and pre-existing-file behavior.
- `SKILL.md:15` — step 5 "Handle any errors that come up" uses the vague verb "Handle" with no concrete follow-up; specify which errors and what to do (refuse and report? retry? roll back the tag?).
- `SKILL.md:12` — step 1 doesn't state preconditions (must there be a prior tag to diff from? must the working tree be clean? must the branch be up to date with the remote?).
- `SKILL.md:81` — the "also update CHANGELOG.md" instruction is a state-mutating step that lives in Notes rather than as a numbered workflow step; either move it into the workflow with a precondition or drop it.

### Security

No findings.

## Passing checks

- Security — the skill does not prompt the model for credentials, accept credentials as arguments, embed credentials in URLs, or write credentials to disk; `gh` and `git` manage their own credentials independently of the model.
- SKILL.md size is well under the 500-line target (92 lines), leaving room to grow as the workflow tightens.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-0/with_skill/work/audit-bad-skill-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
