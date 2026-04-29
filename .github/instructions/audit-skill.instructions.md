---
applyTo: "skills/**,plugins/**/skills/**,agents/**,!**/skills/*-workspace/**"
---

# Skill / agent review checklist

Apply these checks when reviewing a PR that touches a `SKILL.md`, an agent definition, or files under a skill's bundled `scripts/` or `references/`. They mirror the four objectives the `audit-skill` plugin enforces. Raise inline review comments citing the objective name in bold, e.g. **Reproducibility — ...**.

The canonical guidance for each objective lives at `plugins/audit-skill/skills/audit-skill/checks/<objective>.md`. Read those files for the full check list when reviewing — the summaries below are pointers, not replacements.

## Objectives

### Idempotency — `plugins/audit-skill/skills/audit-skill/checks/idempotency.md`

- Does `SKILL.md` declare up front whether re-running is safe? If not, that absence is itself a finding.
- Are mutating commands (`rm`, `mv`, `gh pr create`, `git push`, `curl -X POST`, etc.) preceded by an existence/state check OR an explicit "overwrite" intent?
- Do generated output paths state what happens when the destination already exists?

### Reproducibility — `plugins/audit-skill/skills/audit-skill/checks/reproducibility.md`

- Vague directives ("as appropriate", "as needed", "if relevant", "appropriately", "use judgment") without a concrete criterion (a threshold, a regex, a named condition).
- Implicit environment dependencies (`date`, `git status`, `pwd`, network calls) used without being declared as inputs.
- Examples that don't actually exercise the rule stated above them.

### Context management — `plugins/audit-skill/skills/audit-skill/checks/context-management.md`

- `SKILL.md` over ~500 lines (soft target).
- Long inline blocks (>50 lines) or large tables that should be in a sibling `references/<topic>.md`.
- Heavy lifts (multi-file reads, item-by-item processing) done in main context without delegating to subagents.
- Reference files >300 lines without a TOC.

### Strict definitions — `plugins/audit-skill/skills/audit-skill/checks/strict-definitions.md`

- Description must answer all three: what it does, when to use, when to skip.
- Inputs declared with name, source (CLI / env / prompt / file), required vs optional, validation.
- Outputs declared with path, format, and behavior on pre-existing files.
- Step ordering states preconditions and dependencies.

## Style for the review

- One finding per inline comment. Lead with `**<Objective>** — `.
- Cite `file:line`. Don't paraphrase the offending line; quote it when short.
- Suggest a fix only when it's short and obvious. Don't editorialize ("this is bad practice").
- No severity tiers — the four objectives ARE the categorization. Leave it to the author to decide what blocks the merge.
- If everything passes, leave one short approving comment naming the checks that passed.

## What this checklist does NOT apply to

- Sibling `<name>-workspace/` trees under `skills/` — they are frozen audit-trail snapshots from skill-creator iterations, so findings belong on the source skill, not on the captured outputs.

## What this checklist does NOT replace

- Spelling, grammar, and link rot — handle as usual.
- Behavioral correctness when running the skill against real prompts. That's runtime evaluation (`skill-creator`), not static audit. If the diff suggests a behavioral regression, flag it as a non-audit comment.
