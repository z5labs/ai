error: refusing --output .git/skills/pg-orders

The path `.git/skills/pg-orders` contains a `.git` segment. This skill's Layer 1 path-safety guard rejects any `--output` value where **any path segment** equals `.git` — not just the leaf. The rule fires here because the first segment of `.git/skills/pg-orders` is literally `.git`.

The reason this guard exists is that the generator's overwrite step runs `rm -rf` on the resolved output directory. If that directory were allowed to sit inside `.git/`, the wipe would destroy part of — or, depending on how the path resolves, all of — the repository's object store, refs, index, and configuration. That loss is irreversible without a remote backup; there is no undo. Protecting `.git/` is therefore a hard stop, not a warning.

This applies to all of the following shapes, not just the one you supplied:

- `.git` (the directory itself)
- `.git/foo` (any subdirectory)
- `subdir/.git` (`.git` as a non-leaf segment)
- `subdir/.git/skills/pg-orders` (`.git` buried in the middle — the case here)

Re-invoke the skill with an `--output` path that does not contain a `.git` segment. For example:

    /postgres-skill-creator --output .claude/skills/pg-orders

Or omit `--output` entirely to accept the default `./.claude/skills/pg-orders/`.
