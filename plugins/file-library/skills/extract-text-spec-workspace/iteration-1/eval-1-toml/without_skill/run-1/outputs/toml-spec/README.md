# TOML v1.0.0 Specification

Extracted from <https://toml.io/en/v1.0.0> on 2026-04-29.

The canonical text version (markdown source) was pulled from
<https://raw.githubusercontent.com/toml-lang/toml.io/main/specs/en/v1.0.0.md>
and the formal grammar from
<https://github.com/toml-lang/toml/blob/1.0.0/toml.abnf>.

## Files

- `full-spec.md` — The complete TOML v1.0.0 spec, as a single document.
- `toml.abnf` — The formal ABNF grammar (RFC 5234) for TOML v1.0.0.
- `00-overview.md` … `21-abnf-grammar.md` — The spec split into one file per
  top-level section, in document order. Useful when you want to load only the
  section relevant to the construct you're implementing.

## Section index

| File | Section | Covers |
|---|---|---|
| `00-overview.md` | (preamble) | Title, authors, publication date |
| `01-objectives.md` | Objectives | Design goals |
| `02-table-of-contents.md` | Table of contents | Spec navigation |
| `03-spec.md` | Spec | Encoding (UTF-8), case sensitivity, whitespace, newline definitions |
| `04-comment.md` | Comment | `#` comments, allowed characters |
| `05-key-value-pair.md` | Key/Value Pair | Basic shape, allowed value types, line rules |
| `06-keys.md` | Keys | Bare keys, quoted keys, dotted keys, redefinition rules |
| `07-string.md` | String | Basic, multi-line basic, literal, multi-line literal; escape sequences; line-ending backslash |
| `08-integer.md` | Integer | Decimal, hex (`0x`), octal (`0o`), binary (`0b`); underscores; 64-bit signed range |
| `09-float.md` | Float | Fractional/exponent forms; `inf`, `nan`; underscores |
| `10-boolean.md` | Boolean | `true`, `false` (lowercase only) |
| `11-offset-date-time.md` | Offset Date-Time | RFC 3339 with offset; `T`/space separator; fractional seconds |
| `12-local-date-time.md` | Local Date-Time | RFC 3339 without offset |
| `13-local-date.md` | Local Date | `YYYY-MM-DD` |
| `14-local-time.md` | Local Time | `HH:MM:SS[.fff]` |
| `15-array.md` | Array | Bracketed, mixed types allowed, trailing commas, multi-line |
| `16-table.md` | Table | `[name]` headers, dotted-name tables, redefinition rules |
| `17-inline-table.md` | Inline Table | `{ k = v, ... }`, single-line, immutable after definition |
| `18-array-of-tables.md` | Array of Tables | `[[name]]`, sub-tables, conflict rules |
| `19-filename-extension.md` | Filename Extension | `.toml` |
| `20-mime-type.md` | MIME Type | `application/toml` |
| `21-abnf-grammar.md` | ABNF Grammar | Pointer to the formal grammar (also bundled as `toml.abnf`) |

## Coverage check (against the user's request)

- Comments — `04-comment.md`
- Strings (all four kinds) — `07-string.md`
- Integers (decimal/hex/octal/binary with underscores) — `08-integer.md`
- Floats — `09-float.md`
- Booleans — `10-boolean.md`
- Datetimes (offset / local datetime / local date / local time) — `11`–`14`
- Arrays — `15-array.md`
- Tables — `16-table.md`
- Inline tables — `17-inline-table.md`
- Arrays of tables — `18-array-of-tables.md`
- Formal grammar — `toml.abnf` + `21-abnf-grammar.md`

Nothing in the published spec was skipped.
