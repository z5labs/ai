---
name: audit-skill
description: Statically audit a Claude Code skill against four quality objectives — idempotency, reproducibility, context management, and strict definitions — and either post findings as PR review comments or write a report file the author can hand back to skill-creator. Use whenever the user asks to audit, review, lint, sanity-check, or critique a skill (their own or someone else's), or asks "is this skill any good", "what's wrong with this skill", "review my SKILL.md". Read-only — never modifies the target skill.
---

# audit-skill

Read-only static analysis of a target skill. Audits its `SKILL.md`, bundled `scripts/`, and `references/`. Produces flat findings (no severity tiers — the author decides what matters) that each cite `file:line` and the objective they threaten. Output goes to a PR review when a PR is open on the current branch; otherwise to a report file at a path the user can hand to `skill-creator`.

This skill is itself bound by the same four objectives. If you find yourself wanting to add a vague directive ("audit thoroughly", "use good judgment"), turn it into a concrete check in the appropriate `checks/<objective>.md`.

## Inputs

- **Target skill** — a name (e.g. `extract-text-spec`) or absolute path. If a name, resolve in this order, first match wins:
  1. `~/.claude/skills/<name>/`
  2. `./.claude/skills/<name>/`
  3. `./plugins/*/skills/<name>/`
  
  If multiple paths match, list them all and ask the user which to audit. If none match, stop and report the search paths tried.

- **Output mode** — auto-detected. Run `gh pr view --json number,headRefName,baseRefName 2>/dev/null`. Exit 0 with JSON ⇒ PR mode. Anything else ⇒ file mode.

## Workflow

1. **Resolve the target.** Confirm `SKILL.md` exists at the resolved path. List bundled files: `find <skill-dir> -type f \( -name '*.md' -o -name '*.sh' -o -name '*.py' \)`.

2. **Run each objective in sequence.** For each of the four below, read the matching `checks/<objective>.md` only when starting that objective (progressive disclosure — don't load all four upfront). Each check file lists concrete patterns to look for and how to phrase the finding. Append findings to an in-memory list as you go.

   - **Idempotency** — `checks/idempotency.md`. Skip with a note in the report if `SKILL.md` declares the skill non-idempotent up front (e.g. "this skill is intentionally non-idempotent because…"). Otherwise the absence of that declaration is itself a finding.
   - **Reproducibility** — `checks/reproducibility.md`. Always run.
   - **Context management** — `checks/context-management.md`. Always run.
   - **Strict definitions** — `checks/strict-definitions.md`. Always run. Absorbs trigger-quality (description, when-to-use, when-to-skip).

3. **Emit output.** Read `report-template.md` for the exact format.
   - **PR mode**: For each finding whose `file:line` falls on a line modified in the PR diff, post an inline review comment. Findings on unchanged lines, plus an audit summary, go in a single top-level review comment. Use `gh pr review --comment -F -` for the summary and `gh api repos/{owner}/{repo}/pulls/{n}/comments` for inline comments. Determine modified lines via `gh pr diff --patch` — parse `@@` hunks. If no findings, post a single summary comment saying "audit clean — N checks passed".
   - **File mode**: Write to `./audit-<skill-name>-<YYYY-MM-DD>.md` using the report template. Tell the user the path so they can pass it to `skill-creator` for revision (`/skill-creator <path>` or just attach the file in conversation).

4. **Report back.** Always tell the user: target audited, output mode used, finding count by objective, and where the output went.

## Findings format

Every finding cites `file:line` and the objective it threatens. No severity tiers — the four objectives ARE the categorization. See `report-template.md` for the exact schema and example findings.

A finding is worth raising only if it's specific enough that the author could fix it without guessing what you meant. "Description is vague" is not a finding; "description doesn't say when NOT to trigger (line 3)" is.

## What this skill does NOT do

- Run the target skill or evaluate runtime behavior — that's `skill-creator` with eval prompts.
- Modify the target skill — read-only, always.
- Score or rank skills — there's no aggregate grade, just findings.
- Apply severity tiers — the author judges what matters.
