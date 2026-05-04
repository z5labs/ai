**Error: unsafe `--output` path rejected.**

The path `../../etc` contains `..` segments and cannot be used as the output directory. This skill refuses to proceed.

**Why this matters:** The output directory is wiped with `rm -rf` before being recreated (overwrite is intentional — schemas drift, and stale references mislead). A `..`-containing path like `../../etc` resolves to a location entirely outside the intended directory tree. `rm -rf /tmp/out/..` resolves to `/tmp` (or worse); `rm -rf ../../etc` from the current working directory could delete arbitrary system or project directories. A literal-string check on the leaf alone would not catch this — `../../etc` passes leaf-safety but fails path-component safety.

**The rule that fired (layer 2 of 3):** Reject paths whose *any* component includes `..`. This matches `..` at any position: leading (`../foo`), trailing (`foo/..`), middle (`foo/../bar`), or lone (`..`). The value `../../etc` has two leading `..` components and fails immediately.

**No further steps were taken.** Introspection, container-runtime detection, and all disk writes are skipped — path-safety guards fire first, before anything else.

**To proceed, supply a safe `--output` value** — one with no `..` segments, no empty or whitespace-only components, and a non-empty leaf that is not `.`, `..`, `~`, or `*`. For example:

```
/postgres-skill-creator --output plugins/my-plugin/skills/pg-orders/
```

or omit `--output` entirely to use the default `./.claude/skills/pg-orders/`.
