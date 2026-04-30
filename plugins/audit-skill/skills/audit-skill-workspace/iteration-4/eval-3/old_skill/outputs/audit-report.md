# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-3/old_skill/work/skills/mongo-explorer/SKILL.md`
- Date: 2026-04-30
- Findings: 9  (idempotency: 2, reproducibility: 2, context-management: 0, strict-definitions: 5)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare a top-level idempotency stance. Step 3 says "overwriting any prior version" of the skill directory (line 20), but the preamble never tells the model whether re-invoking `mongo-explorer` itself is safe, what happens to a `.env` already containing different credentials, or whether a partial run can be resumed. State the stance once, up top.
- `SKILL.md:48` — output `./.claude/skills/mongo-<dbname>/.env` is written without specifying overwrite/append behavior. The step says "Generate a `.env` file" but doesn't say whether an existing `.env` (with possibly different credentials) is overwritten, merged, or refused.

### Reproducibility

- `SKILL.md:12` — connection-string parsing is described prose-only ("looks like `mongodb://user:password@host:27017/dbname`") with no validation rule (regex, allowed schemes, required fields). Two runs given subtly different forms (`mongodb+srv://`, missing port, trailing `/?authSource=admin`) will diverge silently. State a parser/regex or reject inputs that don't match a named shape.
- `SKILL.md:44` — step 3's `SKILL.md` substep says "A standard schema-context skill" without naming the template, schema, or section list. The contents of the generated SKILL.md are left to the model, so two runs against the same database will produce different generated skills.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case. With `disable-model-invocation: true` this is less load-bearing for triggering, but humans reading the description still get no guidance on when `mongo-explorer` is the wrong tool (e.g. read-only Atlas clusters, non-Mongo schema introspection, CI environments without `mongo` CLI installed).
- `SKILL.md:3` — description claims the skill "bakes its collection schema into a reference," but the workflow at lines 38–69 also writes a `.env` file and a `scripts/query.sh` wrapper — material side effects on the user's `.claude/skills/` tree that the description never surfaces. Either narrow the description (drop "schema into a reference," say "generates a query-capable skill including credentials and a wrapper script") or narrow the workflow.
- `SKILL.md:11` — input "connection string" is referenced but its required-ness (line 12 says it MAY be omitted, then prompts), validation rule (no regex), and the optional `MONGO_PASSWORD` env-var fallback (line 14) are scattered across three paragraphs rather than declared in one Inputs section.
- `SKILL.md:42` — output "generated `SKILL.md`" has no documented path/format beyond "A standard schema-context skill." No template referenced, no required sections, no example. The generated artifact's contract is undefined.
- `SKILL.md:71` — step 4 ("Verify") doesn't state preconditions: it depends on step 3 having written `scripts/query.sh` and a populated `.env`, but the dependency is implicit. State it (e.g. "After step 3 has written both files…").

## Passing checks

- Frontmatter declares `argument-hint: "[connection-string]"` — slash-style invocation has a documented calling form.
- `SKILL.md` is 80 lines, comfortably under the 500-line target; no inline content needs to move out.
- Workflow is broken into numbered steps with sub-headings, and step 3 enumerates each generated file under its own `###` heading — easy for the model to locate substeps.
- Description's "what" and "when" elements are both present (verb + artifact + a triggering phrase).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-3/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
