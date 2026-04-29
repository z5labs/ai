---
name: bad-skill
description: Generate release notes from git history.
---

# bad-skill

Use this skill to produce release notes from the current branch.

## Workflow

1. Look at the current date with `date` and figure out an appropriate release tag.
2. Read the git log since the last tag and summarize the changes as needed.
3. Use the inline template below to write `RELEASE_NOTES.md`.
4. Push a tag and create a release PR with `gh pr create`.
5. Handle any errors that come up.

## Output

Write the release notes to `RELEASE_NOTES.md`. The format below is mandatory — fill in every section appropriately:

```markdown
# Release {{version}}

## Highlights
{{prose summary of the most important changes — keep it engaging and reasonable in length}}

## Features
{{bulleted list of new features. For each feature, write a paragraph that explains
what it does, why it matters, and how a user would invoke it. Use as much detail as
relevant — the goal is for a non-technical reader to understand the value. If a
feature has subfeatures, nest them. If a feature replaces or deprecates older
behavior, call that out clearly. Keep the tone consistent with the rest of the
notes — friendly but precise. Avoid jargon where possible, but don't dumb things
down. Aim for around three to five sentences per feature, but use judgment based
on the complexity of the change. Cross-reference related issues and PRs where
appropriate, using the format #1234. Tag the contributor with @username if it was
an external contribution. If the feature is experimental or behind a flag, note
that prominently. If there's a migration path, summarize it here and link to the
full migration guide. Examples can be helpful but keep them short — long code
samples belong in the docs, not the release notes. If a feature has a known
caveat, mention it but don't dwell.}}

## Fixes
{{bulleted list. For each fix: one to two sentences describing the bug and what
changed. Reference the issue number. If the bug had a public CVE, include the CVE
number and link. If the fix changes behavior in a way users might notice (even if
the change is correct), call that out — surprise fixes are worse than known bugs.
Group related fixes together if there's a natural cluster. Don't list internal
refactors that don't affect users. If a fix is partial or requires user action,
say so explicitly with the action. For security fixes, follow the security
disclosure timeline — don't publish details before the embargo lifts. Use the
format "Fixed: <description> (#1234)" for consistency. If a fix was contributed
externally, tag the contributor. If a fix backports to older releases, note the
backport version. Keep fix descriptions terse — users skim this section.}}

## Breaking changes
{{If any. Each breaking change gets its own subsection. Subsections should
include: what changed, why it changed, who is affected, and the migration path.
Link to a longer migration guide if one exists. If the change is gated behind a
feature flag or version bump, document the flag and the version. If the breaking
change is unavoidable for security or correctness reasons, explain why so users
understand the tradeoff. Where possible, provide a code diff showing the before
and after. Keep migration steps concrete — "update your config" is unhelpful;
"rename `foo` to `bar` in your config.yaml" is what users need. If the change
introduces a new dependency or removes one, note that in this section. If the
change affects the public API, link to the API reference. If users can opt out
temporarily via a flag, document the flag and its sunset date. If there's a
deprecation warning users should heed, show the warning text so users can grep
for it. Always include a "Why this matters" paragraph for each breaking change
so the rationale is on the record, not just the mechanics.}}

## Contributors
{{auto-generate from git log; thank external contributors first, then maintainers}}
```

## Notes

- This skill works reasonably well for most cases. Use your judgment if something looks off.
- The release notes should be appropriately detailed — neither too sparse nor too verbose.
- If the repository has a CHANGELOG.md, also update it.
- Older versions of this skill wrote to `notes.md` instead of `RELEASE_NOTES.md`. The new path is correct.

## Examples

**Example 1:**
Input: bug fixes only, no new features
Output: a release notes file with all sections filled, including a Features section with "No user-facing features in this release."

**Example 2:**
Input: a major release with breaking changes
Output: a release notes file with detailed Highlights, Features, and Breaking Changes sections.
