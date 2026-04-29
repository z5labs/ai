# Audit Report: `bad-skill`

**Target:** `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-0/without_skill/work/skills/bad-skill/SKILL.md`
**Auditor:** Claude (no checklist; ad hoc judgment-based review)
**Verdict:** **Not ready to ship.** Multiple high-severity issues across triggering, safety, and template usability. Recommend a substantial rewrite before this skill is used on a real repo.

---

## Summary of Severity

| Severity | Count |
|----------|-------|
| Critical (must fix)   | 4 |
| High (should fix)     | 4 |
| Medium                | 5 |
| Low / nit             | 4 |

---

## Critical Issues

### C1. Workflow performs destructive/irreversible side effects without consent
The `Workflow` section instructs the skill to:
- Step 4: "Push a tag and create a release PR with `gh pr create`."

A skill named "Generate release notes from git history" should **not** be cutting tags, pushing, or opening PRs as a default behavior. Tag pushes are effectively irreversible on shared remotes (force-deletion is destructive and breaks downstream consumers). PR creation has lower blast radius but is still surprising. Users invoking a "generate notes" skill expect a file, not a release.

**Fix:** Remove tagging/pushing/PR creation from the default workflow. If the skill should support a "publish" mode, gate it behind an explicit user instruction and require confirmation; never make it the default path.

### C2. Skill `description` is too thin to trigger correctly
```
description: Generate release notes from git history.
```
A skill description is the primary signal Claude uses to decide when to invoke the skill. This one is a single fragment with no triggering examples, no SKIP conditions, and no third-person framing. It will both over-trigger (any time a user mentions git history or notes) and under-trigger (when a user says "draft a changelog entry" or "summarize what shipped this week").

**Fix:** Rewrite as a richer description following the conventions used elsewhere in this codebase (see e.g. the `claude-api` skill description pattern: explicit TRIGGER and SKIP clauses, concrete examples, third-person voice).

### C3. Step 1 picks the version tag heuristically — silently wrong is the default
> "Look at the current date with `date` and figure out an appropriate release tag."

Release tags are not a date-driven decision in any project that uses semver. They are derived from the previous tag plus the nature of the changes (patch/minor/major). Using `date` will produce nonsense like `v2026.04.29` for a project on `v1.4.x`, and worse, the skill will then push that tag (see C1). This is the load-bearing first step of the workflow and it is wrong by construction.

**Fix:** Read the latest tag (`git describe --tags --abbrev=0`), inspect the change set, and either propose a version with rationale or ask the user. Never invent a tag from the calendar.

### C4. "Handle any errors that come up" is not an instruction
> Step 5: "Handle any errors that come up."

This is a placeholder, not behavior. It tells Claude nothing about which errors are recoverable vs. fatal, what state the repo should be left in on failure, or how to report problems back. Combined with C1, an error mid-workflow could leave the repo with a pushed tag but no PR, or a PR with no notes file, etc.

**Fix:** Either remove the step or enumerate the failure modes that matter (no prior tag exists; remote rejects push; `gh` not authed; uncommitted changes in working tree) and specify the recovery for each.

---

## High-Severity Issues

### H1. The mandatory template is a maintenance and quality trap
The `Output` template under "Features" is a single placeholder block containing roughly **20 inline directives** (~250 words of guidance crammed into one `{{...}}`):

> "...Aim for around three to five sentences per feature ... Cross-reference related issues and PRs ... Tag the contributor with @username ... If the feature is experimental or behind a flag, note that prominently. If there's a migration path, summarize it here and link to the full migration guide. Examples can be helpful but keep them short..."

Two problems:
1. **Authoring guidance is mixed into the output template.** A template should describe the shape of the output. Authoring rules belong above it, in prose, where they can be edited without breaking the template's structure.
2. **The directives contradict the section's own framing.** The section is introduced as a "bulleted list of new features", but the directives demand multi-paragraph treatment per feature, contributor tagging, migration links, code examples, caveats, etc. That's a long-form section, not a bulleted list.

**Fix:** Split into (a) a clean structural template with short placeholders and (b) a separate "Authoring guidance" section with the rules.

### H2. The same template-as-prose problem exists in `Fixes` and `Breaking changes`
Both follow the same anti-pattern as `Features` — long instructional blobs inside `{{...}}`. The `Breaking changes` block in particular embeds rules about CVE handling, embargo timelines, deprecation warnings, opt-out flags with sunset dates, and "Why this matters" paragraphs. These are policy-level instructions, not template content.

**Fix:** Same as H1 — separate structure from guidance.

### H3. Mandatory format vs. variable inputs creates a contradiction
> "The format below is mandatory — fill in every section appropriately"

Then Example 1 shows the output for "bug fixes only, no new features" producing a Features section that says "No user-facing features in this release." So the rule is actually: every section is present, but sections without content get a stock filler line. That's reasonable, but it isn't what the skill says, and the discrepancy will lead to either omitted sections or empty sections depending on which instruction Claude latches onto.

**Fix:** Replace "mandatory — fill in every section appropriately" with explicit conditional rules ("If there are no breaking changes, omit the section" OR "Always include the section; if empty, write `None.`"). Pick one and be consistent.

### H4. CHANGELOG.md update is a footnote, not a workflow step
> Notes: "If the repository has a CHANGELOG.md, also update it."

A CHANGELOG update is a real edit to a real file, sometimes with a strict format (Keep a Changelog, conventional commits, etc.). Burying it in a Notes bullet means it will be skipped or done inconsistently. Also, updating both `RELEASE_NOTES.md` and `CHANGELOG.md` raises questions about which is canonical and how they should be kept in sync — none of which the skill addresses.

**Fix:** Decide whether CHANGELOG is in scope. If yes, make it a workflow step with explicit format guidance. If no, drop the bullet.

---

## Medium-Severity Issues

### M1. "Since the last tag" assumes a tag exists
Step 2 says read the git log "since the last tag." If the repo has never been tagged, `git log <tag>..HEAD` will fail. The skill should detect this and fall back (e.g., entire history, or first commit). This intersects with C3 — the tag selection logic is unspecified throughout.

### M2. No specification of which branch to operate on
> "produce release notes from the current branch"

Releases are usually cut from `main`/`master` or a release branch, not whatever the user happens to be on. The skill should at minimum warn if the current branch isn't a typical release branch, or ask which range to summarize.

### M3. Scope of "Contributors" auto-generation is underspecified
> "auto-generate from git log; thank external contributors first, then maintainers"

How does the skill know who is "external" vs. "maintainer"? GitHub orgs? CODEOWNERS? A maintainers file? Email domain? Unspecified. Without a clear rule, this will either be skipped or guessed.

### M4. Stale-skill warning is concerning
> "Older versions of this skill wrote to `notes.md` instead of `RELEASE_NOTES.md`. The new path is correct."

If this is documentation for human maintainers, it's fine but belongs in a CHANGELOG for the skill itself, not in the runtime instructions. If it's a hint to Claude to "ignore older invocations," it's likely to confuse rather than help. Either way it leaks implementation history into runtime context.

### M5. No mention of file overwrite / append behavior
If `RELEASE_NOTES.md` already exists (e.g., from a prior release), the skill silently overwrites it. For a release-notes file that is typically appended to (or rotated per-version), this is wrong behavior. Specify: append a new section at the top, or write a per-version file (`RELEASE_NOTES_v1.4.0.md`), or confirm with the user.

---

## Low-Severity / Nits

### L1. Skill name is `bad-skill`
Presumably a placeholder, but worth flagging — the `name` frontmatter and the H1 heading both say `bad-skill`. Real users will see this name surfaced in skill listings.

### L2. "Use your judgment if something looks off" is non-actionable
The Notes section opens with "This skill works reasonably well for most cases. Use your judgment if something looks off." This is filler. Claude does not need to be told to use judgment, and "reasonably well for most cases" is a self-undermining claim that gives no useful signal.

### L3. Examples are sparse
The two examples describe inputs and outputs in a single sentence each, with no actual rendered output. For a skill whose value is mostly in the formatting of the output, showing a real (small) rendered example would be far more useful than the abstract description.

### L4. Tone instructions inside the template pull two ways
The Features template asks for "friendly but precise", "engaging", non-jargon prose, while also requiring contributor tagging, issue cross-refs, CVE links (in Fixes), and migration tables. The voice/structure tradeoffs aren't reconciled.

---

## Recommendations (priority order)

1. **Strip the workflow back to "generate a notes file."** Remove tagging, pushing, and PR creation from the default path. (C1)
2. **Rewrite `description`** with explicit triggers and skip conditions, modeled on the existing skills in this repo. (C2)
3. **Replace the date-based tag heuristic** with a real version-bump logic step (or ask the user). (C3)
4. **Separate template structure from authoring guidance.** Move all the prose-style rules out of `{{...}}` blocks into a dedicated section. (H1, H2)
5. **Decide CHANGELOG scope** and either spec it or remove it. (H4)
6. **Pick one rule for empty sections** ("always present with `None.`" vs. "omit entirely") and apply it everywhere. (H3)
7. **Specify branch and prior-tag handling**, including the no-prior-tag case. (M1, M2)
8. **Specify the Contributors heuristic** or drop the "external first" rule. (M3)
9. Polish: rename the skill, tighten Notes, add at least one rendered example. (L1-L4)

After (1)-(6) the skill would be roughly fit for purpose. Items (7)-(9) are quality-of-life.

---

## Did anything work well?

A few things, briefly, so the rewrite doesn't lose them:
- The overall section list (Highlights / Features / Fixes / Breaking changes / Contributors) is a reasonable shape for release notes.
- Splitting the file into Workflow / Output / Notes / Examples is the right top-level structure for a SKILL.md, even if the content within those sections is off.
- The instinct to call out backports, CVEs, embargoes, and migration paths is correct — those are real concerns. They just belong in authoring guidance, not in the template body.
