# Audit: bad-skill

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-0/with_skill/work/skills/bad-skill/`
- Date: 2026-04-30
- Findings: 18  (idempotency: 4, reproducibility: 8, context-management: 1, strict-definitions: 5, security: 0)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses so the model doesn't have to infer it from the workflow.
- `SKILL.md:14` — output `RELEASE_NOTES.md` is written without specifying overwrite/append behavior; a second run silently clobbers the prior file. State the policy (overwrite / refuse-if-exists / append).
- `SKILL.md:15` — `git tag` followed by `gh pr create` are stateful external calls with no precondition; running twice will fail at the tag step or open a duplicate PR. Add a "check for existing tag / existing release PR before creating" step, or document that the duplicate is acceptable.
- `SKILL.md:81` — "also update [CHANGELOG.md]" mutates a file with no statement of overwrite vs append behavior; re-runs will accumulate or destroy entries unpredictably. Specify the merge rule (e.g. prepend a new section above the previous one, refuse if the section already exists).

### Reproducibility

- `SKILL.md:12` — depends on `date` (current date) without listing it as a declared input; the same prompt run on different days produces different release tags.
- `SKILL.md:12` — "figure out an appropriate release tag" gives no objective criterion (semver bump rule? read prior tag?); reproducibility requires a stated test.
- `SKILL.md:13` — depends on `git log` (repo state) without listing it as a declared input or naming the expected branch / range / commit selector.
- `SKILL.md:13` — "summarize the changes as needed" gives no objective criterion for what to include or exclude.
- `SKILL.md:20` — "fill in every section appropriately" gives no objective criterion; "appropriately" is the entire content of the directive.
- `SKILL.md:26` — template field instructs a "reasonable in length" prose summary with no concrete word/sentence bound; two runs will produce different lengths with no way to tell which is right.
- `SKILL.md:35` — "use judgment based on the complexity of the change" is asked of the model with no rubric or tiebreaker; runs will diverge.
- `SKILL.md:79-80` — "works reasonably well for most cases. Use your judgment if something looks off" and "appropriately detailed — neither too sparse nor too verbose" are pure hedge words with no criterion attached. Replace with concrete bounds (e.g. "Highlights: 3–5 sentences", "Features: one paragraph per feature, max 5 sentences") or delete.

### Context management

- `SKILL.md:22-75` — the mandatory release-notes template (~54 lines, much of it dense prose guidance on how to fill each section) is inline; consider moving the template plus the per-section guidance to `references/release-notes-template.md` and referencing it. Keeping it inline is defensible since the model must copy the template literally, but the per-section paragraphs of editorial guidance (Features, Fixes, Breaking changes) are reference material that bloat every invocation.

### Strict definitions

- `SKILL.md:3` — description "Generate release notes from git history." has no "when to use" examples (no concrete user phrasings or contexts that should trigger); the skill is likely to under-trigger.
- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when the user wants a CHANGELOG-only update", "skip when no tags exist yet"); likely to over-trigger on near-misses.
- `SKILL.md:1-10` — no `Inputs` / `Arguments` section. The skill consumes at least: current date, current git branch, range since last tag, optional CHANGELOG.md presence — none are named, sourced, or validated. Add an `Inputs` section listing each, its source, required-ness, and validation.
- `SKILL.md:8` — workflow ends with "Push a tag and create a release PR" but the description never mentions that the skill mutates the remote (pushes a tag, opens a PR). Either narrow the description (only generates the file) or surface the side-effects in the description so triggering decisions account for them.
- `SKILL.md:16` — step 5 "Handle any errors that come up" is a vague verb with no concrete follow-up; the model is given latitude without being told what to actually do. Replace with explicit failure handling (e.g. "if the tag already exists, stop and tell the user; if `gh pr create` fails because a release PR already exists, stop and report its URL").

### Security

No findings.

## Passing checks

- SKILL.md size is well under the 500-line target (92 lines), with no risk of regression.
- No credentials are prompted for, accepted as arguments, or written to disk; the skill leans on `git` and `gh` to manage their own auth out-of-band, which is the correct shape.
- The conditional CHANGELOG.md step (`SKILL.md:81`) does state its predicate ("if the repository has a CHANGELOG.md") rather than being a bare "(optional)".

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-0/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
