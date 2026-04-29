# Reproducibility checks

Two invocations with the same inputs should produce equivalent outputs. Vague language and undeclared environment dependencies are the two biggest threats: vague language because it asks the model to make a judgment call no one can audit, and undeclared deps because they let the run silently depend on state nobody mentioned.

## Checks

### 1. Vague directives without a criterion

Grep for hedge words that ask the model to use judgment without telling it the rubric:

```
grep -niE '(as appropriate|as needed|when relevant|if relevant|when applicable|appropriately|reasonable|reasonably|sensible|properly|where suitable|use (your )?(best )?judgment|carefully|thoroughly|rigorously)' SKILL.md checks/ references/ 2>/dev/null
```

The pattern omits `\b` word boundaries because BSD `grep -E` doesn't support them portably. The grep is a candidate-finder; expect substring matches (e.g. "thoroughfare" would match "thoroughly") and ignore them during triage.

For each hit, check whether the surrounding sentence DEFINES what "appropriate" / "needed" / "relevant" means in this context. If the criterion is named (e.g. "when the file exceeds 500 lines, do X" — "when" + concrete threshold = fine), it's not a finding. If the model is left to guess, it is.

Phrase as: `<file>:<line> — "<exact phrase>" gives no objective criterion; reproducibility requires a stated test (a threshold, a regex, a named condition)`.

Be selective. "Use the appropriate tool" with a concrete table below it is fine. "Audit thoroughly" with no checklist is not.

### 2. Implicit environment dependencies

Search for reads from the runtime environment that aren't declared as inputs:

```
grep -nE '(date|whoami|hostname|pwd|uname|git status|git log|curl|wget|gh (api|pr view|pr list|repo view))' SKILL.md scripts/ 2>/dev/null
```

Same caveat as above: no `\b` boundaries. Expect substring matches (e.g. "update" matches "date"); ignore them during triage.

A skill that uses any of these must list it under "Inputs" or "Preconditions" with the expected shape. Otherwise the same prompt run on a different day, a different working directory, or a different network can silently produce different results.

Raise: `<file>:<line> — depends on <X> (current date / git state / network / cwd) without listing it as a declared input`.

The implicit-input check is strict: even widely-used commands like `date` are findings if undeclared. Reproducibility means *someone reading the skill* should be able to predict the output from the inputs alone.

### 3. Examples that drift from the rule

For every example in `SKILL.md` (look for fenced code blocks following "Example", "For example", or labeled Input/Output pairs):
- Find the rule the example is illustrating (usually the paragraph or bullet immediately above).
- Check the example actually exercises that rule.

If an example shows a different shape than the rule (e.g. rule says "always include a `## When to skip` section", example skill body has no such section), raise: `<file>:<line> — example contradicts the rule stated at <other-line>`.

This check is the one most likely to require careful reading rather than grep. Skim the examples once with the rule fresh in mind.

### 4. Non-deterministic operations without anchors

If the skill calls for sampling, LLM-as-judge, or "pick the best N", check whether there's a tiebreaker, a seed, or a rubric. Subjective LLM judgments are the easiest reproducibility leak — two runs return two different answers and there's no way to tell which was right.

Grep for: `(pick|select|choose|sample|prioritize|rank|score)` (no `\b` boundaries — BSD `grep -E` doesn't support them portably) and look for an accompanying rubric or deterministic rule.

Raise: `<file>:<line> — <verb> is asked of the model without a tiebreaker or rubric; runs will diverge`.

## What is NOT a finding

- "Approximately" or "roughly" attached to a numeric guideline (`~500 lines`) — these are fine because the threshold is named even if soft.
- Hedge words inside a quoted example of bad writing the skill is teaching against.
- Date/time used in a path or filename that's part of the documented output (`audit-<YYYY-MM-DD>.md`) — declared output, not implicit input.
