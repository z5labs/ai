# gitconfig Specification Reference

## Overview

The Git configuration file (`gitconfig`) is the line-oriented, INI-dialect text
format used by Git to store per-user, per-repository, per-worktree, and
system-wide settings. Files following this format live at well-known locations
(`/etc/gitconfig`, `$XDG_CONFIG_HOME/git/config`, `~/.gitconfig`,
`$GIT_DIR/config`, `$GIT_DIR/config.worktree`) and are also reachable through
`include.path` / `includeIf.*.path` directives. This document covers the
syntax described under "CONFIGURATION FILE" → "Syntax", "Includes",
"Conditional includes", and "Values" of the `git-config(1)` manual page
(<https://git-scm.com/docs/git-config>). The grammar below is paraphrased into
EBNF — the upstream documentation is informal prose, not a formal grammar.
Encoding is not formally specified; in practice contents are read as bytes and
treated as ASCII/UTF-8 with no BOM (see Appendix and Ambiguity callouts).

> **Ambiguity:** "gitconfig" is not standardised by an external body. The
> `git-config(1)` manual is the authoritative reference, but the canonical
> implementation (`config.c` in the Git source tree) accepts a few constructs
> the prose does not describe (notably the case of `\` immediately before EOF
> with no newline, NUL bytes inside double-quoted values, and the precise set
> of valid bytes in subsection names). Where the prose is silent or
> contradictory, this document calls it out and the implementer should match
> upstream `config.c` semantics.

This document refers to top-level groupings as **sections**, nested groupings
as **subsections**, and individual key/value declarations as **variables**.
The lexical elements produced by the tokenizer are called **tokens**; grammar
rules are called **productions**.

## Lexical Elements (Tokens)

The gitconfig tokenizer is line-oriented. Whitespace inside a logical line is
generally insignificant except inside double-quoted strings, inside subsection
names, and as a separator between a section name and its quoted subsection.
A line ending in an unquoted, unescaped backslash is continued onto the next
line (the backslash and the line terminator are consumed by the tokenizer
before the value token is finalised).

The token classes below cover every byte the format recognises.

### Comments

- **Name:** `Comment`
- **Pattern:** `( "#" | ";" ) <any byte except LF>* LF`
- **Where allowed:** A comment may appear at the start of a line (the line is
  then a comment-only line and is treated as blank for grammar purposes), at
  the end of a section header line, or at the end of a variable line. The
  introducer (`#` or `;`) must occur outside any double-quoted string.
- **Examples:**
  - `# this is a comment`
  - `; also a comment`
  - `[core] ; trailing comment`
  - `gitproxy = default-proxy ; for all the rest`
- **Edge cases:**
  - `#` and `;` inside a double-quoted value are literal characters, not a
    comment introducer.
  - Comments do not nest. A second `#` or `;` within a comment is part of the
    comment text.
  - A comment runs to the end of line; a backslash immediately before LF
    inside a comment does NOT cause line continuation (continuation only
    applies to value tokens, see `BackslashLineContinuation`).
  - The empty comment `#` followed immediately by LF is valid.

> **Ambiguity:** The spec says "blank lines are ignored" and "comments…​ are
> discarded". It does not state explicitly whether a `#` or `;` flush against
> the start of a line that has leading whitespace before any other content
> still introduces a comment — by convention `   # foo` is a comment. The
> implementer should treat any line whose first non-`SP`/non-`HT` character
> is `#` or `;` as a comment line.

### Whitespace and Delimiters

- **Name:** `Whitespace`
- **Pattern:** `( SP | HT )+`  where `SP = U+0020` and `HT = U+0009`.
- **Significance:**
  - Surrounding whitespace around the variable name, the `=` sign, and the
    value is discarded.
  - Whitespace internal to an unquoted value is preserved verbatim, except
    that trailing whitespace on a value line is stripped before the comment
    introducer is examined.
  - Whitespace inside a double-quoted string is part of the value verbatim.
  - At least one whitespace character is required between a section name
    and its quoted subsection inside `[ ... ]`.
- **Line terminator token:** `Newline`
  - Pattern: `LF` (`U+000A`) or `CRLF` (`U+000D U+000A`).
  - Significance: terminates a logical line. Section headers and variable
    declarations are line-scoped.
- **Blank line token:** `BlankLine` — a line consisting only of `Whitespace`
  (or empty) followed by `Newline`. Blank lines are ignored by the parser.

> **Ambiguity:** The spec says only "space (SP) and horizontal tabulation
> (HT)" are whitespace. `\f`, `\v`, and Unicode whitespace are not recognised
> as whitespace by `git-config(1)`; in practice they are passed through as
> ordinary value bytes. The implementer should follow the spec strictly:
> only `SP` and `HT` are whitespace.

### Line Continuation

- **Name:** `BackslashLineContinuation`
- **Pattern:** an unescaped, unquoted `\` immediately followed by `LF` (or
  `CRLF`) inside a variable's value. The backslash and the line terminator
  are both consumed and replaced by nothing — the next line is read as a
  continuation of the current value.
- **Examples:**
  - ```
    [user]
        bio = This value spans \
              two lines
    ```
    yields a single value `This value spans       two lines` (the leading
    whitespace of the continuation line is preserved verbatim because internal
    whitespace inside an unquoted value is retained).
- **Edge cases:**
  - Continuation does not apply inside a section header (`Section headers
    cannot span multiple lines.`).
  - Continuation does not apply inside a comment line.
  - Inside a double-quoted string, `\` before LF is invalid (the only
    recognised escape sequences are `\"`, `\\`, `\n`, `\t`, `\b`); other
    escapes including `\<LF>` are an error.

> **Ambiguity:** The prose only describes line continuation for value lines.
> Whether a backslash immediately before EOF (no LF) is an error or is
> silently dropped is not stated. Upstream `config.c` treats trailing `\`
> with no following byte as an unterminated value error. The implementer
> should treat it as an error.

### Section Header

- **Name:** `SectionHeader`
- **Pattern:** `[` `Whitespace`? `SectionName` ( `Whitespace`+ `SubsectionLiteral` )? `Whitespace`? `]`
- **Examples:**
  - `[core]`
  - `[remote "origin"]`
  - `[includeIf "gitdir:/path/to/group/"]`
- **Edge cases:**
  - The whole header must lie on a single physical line — no continuation.
  - Whitespace between `[` and the section name is allowed.
  - Whitespace between the closing `]` and a trailing comment is allowed.

### Section Name

- **Name:** `SectionName`
- **Pattern:** `ALPHA ( ALPHA | DIGIT | "-" | "." )*`
  where `ALPHA = %x41-5A / %x61-7A`, `DIGIT = %x30-39`.
- **Examples:** `core`, `branch`, `includeIf`, `url.https://github`,
  `merge-tool`.
- **Edge cases:**
  - Section names are case-insensitive (`[Core]` and `[core]` refer to the
    same section).
  - The dot (`.`) is allowed and historically appears in the deprecated
    `[section.subsection]` syntax (see below). Whether a dot is permitted in
    a "modern" section name (e.g. `[my.section]`) is observed in real configs
    such as `url.https://example`. The spec phrases this as "alphanumeric, -
    and . are allowed in section names".

### Subsection Literal

- **Name:** `SubsectionLiteral`
- **Pattern:** `DQUOTE SubsectionByte* DQUOTE`
  where `SubsectionByte = ( "\" "\"" ) | ( "\" "\\" ) | ( "\" any-other-byte ) | <any byte except DQUOTE, LF, NUL, BACKSLASH>`.
- **Escape semantics inside a subsection literal:**
  - `\"` → `"`
  - `\\` → `\`
  - Any other `\X` (for example `\t`, `\n`, `\0`) → the literal `X`. Backslash
    before any character other than `"` or `\` is dropped; the following
    character is taken verbatim. Note this is **different** from value
    escapes (see `StringValue`), where `\n`/`\t`/`\b` are recognised.
- **Examples:**
  - `[remote "origin"]` → subsection name `origin`.
  - `[includeIf "gitdir:/path/with \"quotes\"/and \\backslashes\\"]` →
    subsection name `gitdir:/path/with "quotes"/and \backslashes\`.
- **Edge cases:**
  - Subsection names are **case-sensitive** (whereas section names are
    case-insensitive).
  - `LF` (`U+000A`) and `NUL` (`U+0000`) are explicitly forbidden inside the
    subsection literal.
  - The empty subsection `""` is syntactically valid; whether it is
    semantically distinct from no subsection is unspecified.

> **Ambiguity:** The spec says subsection names "can contain any characters
> except newline and the null byte" but says nothing about CR (`U+000D`) or
> control characters generally. Upstream `config.c` accepts them literally.
> The implementer should accept any byte other than LF, NUL, and the
> unescaped DQUOTE / unescaped BACKSLASH classes.

### Deprecated Dotted Subsection

- **Name:** `DeprecatedDottedSection`
- **Pattern:** `[` `SectionName` "." `LegacySubsectionName` `]`
  where `LegacySubsectionName = ALPHA ( ALPHA | DIGIT | "-" )*`.
- **Examples:** `[core.gui]`, `[branch.devel]`.
- **Semantics:**
  - The portion after the first dot is the subsection name.
  - The dotted subsection name is **lower-cased** at parse time and is
    matched case-sensitively against other dotted-form references.
  - Same restrictions as `SectionName` apply (alphanumeric, `-`, must start
    with a letter).
- **Edge cases:**
  - Marked deprecated in the spec; new files should use the quoted form.
  - The two forms `[foo.bar]` and `[foo "bar"]` produce the same
    `(section, subsection)` key after lowercasing the legacy form's
    subsection.

> **Ambiguity:** Whether `[foo.Bar]` and `[foo "bar"]` are the same key is
> stated by the spec only indirectly: the dotted form is lowercased, and
> the quoted form is case-sensitive — so `[foo "Bar"]` is distinct from
> `[foo.Bar]` (which becomes `[foo.bar]`). The implementer must lowercase
> only the legacy dotted form.

### Variable Name

- **Name:** `VariableName`
- **Pattern:** `ALPHA ( ALPHA | DIGIT | "-" )*`
- **Case rule:** case-insensitive (canonicalised to lower case for lookup).
- **Examples:** `filemode`, `gitProxy`, `sslVerify`, `auto-crlf`.
- **Edge cases:**
  - Must start with an alphabetic character (no leading digit, no leading
    `-`).
  - Period (`.`) is **not** allowed in a variable name (unlike section names).
  - The maximum length is unspecified.

### Equals Sign

- **Name:** `Equals`
- **Pattern:** `=`
- **Significance:** separates a `VariableName` from its value. Whitespace on
  either side is discarded.

### String Value (unquoted)

- **Name:** `UnquotedValue`
- **Pattern:** the byte stream from the first non-whitespace byte after `=`
  up to the end of the logical line, after backslash continuations have been
  applied and after a trailing comment (introduced by an unquoted `#` or
  `;`) has been stripped.
- **Allowed bytes:** any byte except `LF`, the unescaped comment introducers
  (`#`, `;`), and the unquoted/unescaped `\` (which either continues the
  line or introduces an escape inside a quoted span).
- **Recognised in-value escapes (apply both inside and outside double quotes
  in the value, except as noted):**
  - `\n` → `LF` (`U+000A`)
  - `\t` → `HT` (`U+0009`)
  - `\b` → `BS` (`U+0008`)
  - `\"` → `"`
  - `\\` → `\`
  - Any other `\X` is an error. **Octal escape sequences are explicitly not
    valid.**
- **Examples:**
  - `filemode = false` → value `false`.
  - `external = /usr/local/bin/diff-wrapper` → value
    `/usr/local/bin/diff-wrapper`.
- **Edge cases:**
  - A value with leading or trailing whitespace must be enclosed in double
    quotes. Surrounding whitespace outside the quotes is discarded.
  - A bare `name` with no `=` is a shorthand for boolean true (see
    `Variable` production); it has no value token.

### String Value (quoted span)

- **Name:** `QuotedValueSpan`
- **Pattern:** `DQUOTE ( <any byte except DQUOTE, LF, BACKSLASH> | "\\" Escape )* DQUOTE`
  where `Escape ∈ { "\"", "\\", "n", "t", "b" }`.
- **Examples:**
  - `gitProxy = "ssh" for "kernel.org"` (the value contains two quoted spans
    and unquoted text mixed together — see edge cases).
  - `bio = "  trimmed only when unquoted  "` → preserves leading/trailing
    spaces.
- **Edge cases:**
  - A value may mix quoted spans with unquoted bytes. The whole value is the
    concatenation of unquoted runs and the (escape-decoded, quote-stripped)
    contents of each quoted span on the same logical line.
    Example: `name = "Hello "world` parses to `Hello world`.
  - Inside a quoted span, `#` and `;` are literal — they do not introduce a
    comment.
  - A quoted span may not contain a literal newline; line continuation
    (`\` before LF) is also not allowed inside a quoted span.

> **Ambiguity:** Whether `\` before any byte other than `"`, `\`, `n`, `t`,
> `b` inside a `QuotedValueSpan` should be an error or should silently drop
> the backslash is contradicted between the prose ("Other char escape
> sequences (including octal escape sequences) are invalid.") and historical
> implementations that have at times silently accepted them. The implementer
> should treat `\X` for any other `X` as a parse error inside the value.

### Symbols

- **Name:** `LBracket` — `[` — opens a section header.
- **Name:** `RBracket` — `]` — closes a section header.
- **Name:** `DoubleQuote` — `"` — delimits a `SubsectionLiteral` or a
  `QuotedValueSpan`.
- **Name:** `Backslash` — `\` — introduces escapes (`\"`, `\\`, `\n`, `\t`,
  `\b`) and, at end-of-line outside quotes, line continuation.
- **Name:** `Dot` — `.` — separator inside `SectionName` and the
  `DeprecatedDottedSection` form.

### Keywords / Reserved Names

The format has no reserved keywords at the lexical level. The conditional
include keywords are reserved at the **semantic** level:

- `gitdir`, `gitdir/i`, `onbranch`, `hasconfig:remote.*.url` — recognised
  values inside an `[includeIf "<keyword>:<data>"]` subsection literal. They
  are matched as a literal byte prefix of the subsection name up to (and
  including) the first `:`. Other keywords are forbidden in `includeIf`
  subsection names and cause the include to be ignored.

These tokens are not enforced by the lexer — they are checked by the
semantics layer.

### Type-Tagged Values (parsed, not lexed)

The lexer produces a single string value per variable. The five "types"
described in the manual page are not lexical token classes — they are
secondary parses applied to the string value when the consumer asks for a
typed value (`--type=bool`, `--type=int`, `--type=color`, `--type=path`).
They are listed here for completeness and detailed under `## Semantics` →
"Typed values".

| Type | Synonyms / shape |
|---|---|
| boolean | `true`, `yes`, `on`, `1`, plus *no value at all* → true; `false`, `no`, `off`, `0`, empty string → false; case-insensitive |
| integer | optional sign, decimal digits, optional unit suffix `k`/`K`, `m`/`M`, `g`/`G`, `t`/`T`, `p`/`P` for `*1024^n` |
| color | one or two color tokens (`black`, `red`, …, `default`, `normal`, `bright<color>`, `0`–`255`, `#RRGGBB`, `#RGB`) plus zero or more attribute tokens (`bold`, `dim`, `ul`, `blink`, `reverse`, `italic`, `strike`, `reset`, `no<attr>` / `no-<attr>`) separated by spaces |
| pathname | string with optional `~/` / `~user/` / `%(prefix)/` prefix, optionally tagged with `:(optional)` |
| string | the raw value, no transformation |

## Structure (Grammar)

The grammar below uses EBNF with the conventions:

- `?` — zero or one
- `*` — zero or more
- `+` — one or more
- `|` — alternation
- terminals appear in `"double quotes"` or as token-class names from
  `## Lexical Elements`
- character-class shorthand: `ALPHA = "A".."Z" | "a".."z"`,
  `DIGIT = "0".."9"`, `SP = " "`, `HT = "\t"`, `LF = "\n"`,
  `DQUOTE = "\""`.

### Top-Level Structure

A gitconfig file is a (possibly empty) sequence of lines. Three line shapes
exist:

1. blank/comment lines (ignored),
2. section-header lines (open a new active section),
3. variable lines (declare a key/value within the currently active section).

Variable lines may only appear after a section-header line — there is no
implicit "preamble" section. The file ends at end-of-input; the final line
need not be terminated by a newline (this is unspecified by the manual page,
see Ambiguity below).

> **Ambiguity:** The spec does not say whether a file must end with a
> trailing newline. Upstream `config.c` accepts a final variable line
> without trailing LF; the printer in this library should always emit a
> trailing LF for normalisation.

### Grammar Productions

```
ConfigFile      = Line*
Line            = SectionLine | VariableLine | BlankOrCommentLine
BlankOrCommentLine = Whitespace? ( Comment )? Newline

SectionLine     = Whitespace? SectionHeader Whitespace? ( Comment )? Newline
SectionHeader   = ModernSectionHeader | DeprecatedSectionHeader
ModernSectionHeader
                = "[" Whitespace? SectionName ( Whitespace+ SubsectionLiteral )? Whitespace? "]"
DeprecatedSectionHeader
                = "[" Whitespace? SectionName "." LegacySubsectionName Whitespace? "]"

SectionName     = ALPHA ( ALPHA | DIGIT | "-" | "." )*
LegacySubsectionName
                = ALPHA ( ALPHA | DIGIT | "-" )*
SubsectionLiteral
                = DQUOTE SubsectionChar* DQUOTE
SubsectionChar  = <any byte except LF, NUL, DQUOTE, "\">
                | "\" DQUOTE
                | "\" "\"
                | "\" <any-other-byte>          ; backslash dropped, byte kept

VariableLine    = Whitespace? VariableDeclaration Whitespace?
                  ( Comment )? Newline
VariableDeclaration
                = VariableName ( Whitespace? "=" Whitespace? Value )?

VariableName    = ALPHA ( ALPHA | DIGIT | "-" )*

Value           = ValuePart+
ValuePart       = QuotedValueSpan | UnquotedValueRun | LineContinuation
LineContinuation
                = "\" Newline                  ; backslash + LF, both consumed
UnquotedValueRun
                = ( <any byte except LF, "#", ";", "\", DQUOTE>
                  | "\n"                       ; → LF
                  | "\t"                       ; → HT
                  | "\b"                       ; → BS
                  | "\\"                       ; → "\"
                  | "\""                       ; → DQUOTE
                  )+
QuotedValueSpan = DQUOTE
                  ( <any byte except LF, DQUOTE, "\">
                  | "\""                       ; → DQUOTE
                  | "\\"                       ; → "\"
                  | "\n"                       ; → LF
                  | "\t"                       ; → HT
                  | "\b"                       ; → BS
                  )*
                  DQUOTE

Comment         = ( "#" | ";" ) <any byte except LF>*
Newline         = LF | CR LF
Whitespace      = ( SP | HT )+
```

### Members / Fields per Production

- `ConfigFile`
  - `lines` — ordered list of `Line`. Order is significant (later assignments
    override earlier ones; multivalued variables retain all values in order).

- `SectionLine`
  - `header` — exactly one `SectionHeader`.
  - `trailingComment` — optional `Comment`.

- `SectionHeader`
  - `name` — `SectionName`, lowercased for canonical comparison.
  - `subsection` — optional string. From `SubsectionLiteral`: case-sensitive,
    raw decoded bytes. From `DeprecatedSectionHeader`: lowercased.
  - `dialect` — `modern` or `deprecated` (drives printer output).

- `VariableLine`
  - `name` — `VariableName`, lowercased for canonical comparison.
  - `value` — optional. Absent ⇒ implicit boolean true.
  - `trailingComment` — optional.

- `Value`
  - `parts` — ordered list of `ValuePart`. The string value is the
    concatenation of part contents after escape decoding and continuation
    splicing.

### Ordering and Optionality

- A `VariableLine` is only valid when at least one `SectionLine` has appeared
  earlier in the same file (or in a transitively-included file already
  parsed).
- A `SectionHeader` may be repeated. A second `[core]` re-opens the same
  logical section; variables defined under the re-opened header append to
  the same `(section, subsection)` namespace.
- A `VariableName` may be repeated within the same `(section, subsection)`.
  Same-name variables are either *single-valued* (last write wins, i.e. the
  reader returns the last occurrence) or *multivalued* (all values are
  returned in source order). Whether a particular variable is multivalued
  is **not** declared syntactically — it is determined by the consumer's
  expectation.
- Quoted and unquoted runs may concatenate within a single value. There must
  be no whitespace between them outside the quotes (otherwise the second run
  is a separate value-byte sequence joined by the intervening internal
  whitespace).
- Line continuation may appear anywhere a `Value` token may appear, except
  inside a `QuotedValueSpan`.
- A `SectionHeader` whose name does not satisfy the `SectionName` rule, or
  whose subsection literal contains a forbidden byte, is a parse error and
  the whole file is rejected.

## Semantics

### Section / variable identity

The canonical identity of a setting is the triple
`(lower(section), subsection-or-null, lower(variable))`. Section names are
case-insensitive; subsection names from the modern quoted form are
case-sensitive; subsection names from the deprecated dotted form are
lowercased at parse time. Variable names are case-insensitive.

### Multivalued variables

A variable that appears more than once in the same section is "multivalued"
when the reader requests all values. Otherwise the last occurrence wins.
The grammar does not distinguish — both shapes produce the same AST; the
difference is in how the consumer queries the AST.

### File precedence (when read in scope)

When git assembles configuration from multiple files, the order is:

1. system (`/etc/gitconfig`)
2. global (`$XDG_CONFIG_HOME/git/config`, then `~/.gitconfig`)
3. local (`$GIT_DIR/config`)
4. worktree (`$GIT_DIR/config.worktree`, only if
   `extensions.worktreeConfig` is present in local)
5. command (`-c key=value`, `GIT_CONFIG_KEY_n` / `GIT_CONFIG_VALUE_n`)

Last value found takes precedence for single-valued reads; for multivalued
reads, all values are concatenated in source order. This precedence is
**outside** the file format itself but is required for correct semantic
interpretation across files. A library that only parses one file at a time
need not implement it; a library that materialises the effective
configuration must.

### Includes

The variables `include.path` and `includeIf.<condition>.path` cause another
gitconfig file to be parsed and its lines spliced in at the position of the
`path` directive.

- The value of `include.path` is a pathname (subject to `~` expansion).
- A relative path is resolved relative to the file in which the `include`
  directive appears (not the current working directory, not the file being
  effectively included by).
- `include.path` and `includeIf.*.path` may be set multiple times; each
  occurrence is processed in source order.
- Cycles (file A includes file B which includes file A) are not addressed by
  the spec. Upstream `config.c` detects and breaks them.

> **Ambiguity:** The spec does not specify a maximum include depth or define
> what happens when a referenced file does not exist. Upstream `config.c`
> silently skips a missing include file (so writers can use `include.path`
> defensively). The implementer should mirror this lenient behaviour and
> document any deviation.

### Conditional includes

`includeIf.<condition>.path` is processed exactly like `include.path`,
except its `path` is only effective when the `<condition>` resolves to true.
The condition is the full subsection name, of the form
`<keyword>:<data>`. Recognised keywords:

| Keyword | Data | Match rule |
|---|---|---|
| `gitdir` | glob pattern | Matches the canonical filesystem path of the active `.git` directory. Case-sensitive. |
| `gitdir/i` | glob pattern | Same as `gitdir`, case-insensitive. |
| `onbranch` | glob pattern | Matches the currently checked-out branch name. |
| `hasconfig:remote.*.url` | glob pattern | True if any `remote.*.url` value (across all files read so far, plus a forward-scan of remaining files) matches the pattern. |

Glob extensions (over POSIX glob): `**/` matches any sequence of full path
components at the start of a segment; `/**` matches any sequence of full
path components at the end. Patterns starting with `~/` get `~` expanded;
patterns starting with `./` are anchored to the directory of the current
config file; patterns not starting with `~/`, `./`, or `/` get an implicit
`**/` prefix; patterns ending with `/` get an implicit `**` suffix.

> **Ambiguity:** "the canonical filesystem path of `.git`" is murky in the
> presence of symlinks. The spec says: symlinks inside `$GIT_DIR` are NOT
> resolved before matching, but symlink and realpath versions of paths
> outside `$GIT_DIR` are both matched. The v2.13.0 release matched only the
> realpath version — configurations that need cross-version compatibility
> must declare both. The implementer should match upstream's current
> behaviour and surface this as a known compatibility note.

### Typed values

When a variable's value is read with a type request, the raw string is
re-interpreted:

- **boolean:**
  - True synonyms (case-insensitive): `true`, `yes`, `on`, `1`. A bare
    `name` with no `=` is also true.
  - False synonyms (case-insensitive): `false`, `no`, `off`, `0`, the empty
    string.
  - Any other value is an error.
- **integer:** decimal integer with optional sign, optional unit suffix.
  Suffixes (case-insensitive) and their multipliers:
  - `k` → 1024
  - `m` → 1024² (1 048 576)
  - `g` → 1024³
  - `t` → 1024⁴
  - `p` → 1024⁵

  > **Ambiguity:** The spec gives the `k`, `M`, "…​" hand-wave but does not
  > exhaustively list the suffixes. Upstream `config.c`'s
  > `git_parse_signed()` accepts `k`, `m`, `g`, `t`, `p`. The implementer
  > should accept this set and reject other letters.

- **color:** at most two color terms (foreground, then background) and
  any number of attribute terms, in any interleaving, separated by `SP`.
  - Color terms: `normal`, `default`, `black`, `red`, `green`, `yellow`,
    `blue`, `magenta`, `cyan`, `white`, `bright<color>` for any of those
    eight basic colors except `normal`/`default`, an integer `0`–`255`,
    or `#RRGGBB` (24-bit hex) or `#RGB` (12-bit hex, expanded by repeating
    each nibble).
  - Attribute terms: `bold`, `dim`, `ul`, `blink`, `reverse`, `italic`,
    `strike`, plus the negated forms `no<attr>` and `no-<attr>`, plus the
    pseudo-attribute `reset`.
  - The empty string is a valid color value and produces no effect.
- **pathname:** a string with one of the optional prefixes:
  - `~/` → expanded against `$HOME`.
  - `~user/` → expanded against the named user's home directory.
  - `%(prefix)/` → expanded against Git's compile-time runtime prefix.
  - `./` (literal) → preserved verbatim, used to opt out of `%(prefix)`
    expansion.

  Optionally a path value may be prefixed with `:(optional)`; this marker
  tells consumers that a missing target should be treated as if the variable
  were unset rather than as an error.
- **string:** raw value, no transformation.

These typed parses are NOT performed by the gitconfig parser — they are the
responsibility of the consumer that asks for a typed read. A library that
materialises the AST should retain the raw string and expose typed parsers
as separate functions.

### Equivalence and uniqueness

- `[Foo "Bar"] x = 1` and `[foo "Bar"] x = 1` set the same key
  (`foo.Bar.x`).
- `[Foo "Bar"] x = 1` and `[foo "bar"] x = 1` set **different** keys
  (subsections are case-sensitive).
- `[foo.Bar] x = 1` after lowercasing becomes `[foo.bar] x = 1`, distinct
  from `[foo "Bar"] x = 1`.
- Setting a boolean by bare-name shorthand and setting it explicitly are
  semantically equivalent for boolean reads, but textually distinct in
  the AST (the printer must round-trip what was written).

## Examples

### Minimal Valid File

```
[core]
	filemode = false
```

A single section with a single key/value. This is the smallest non-empty
gitconfig that defines a setting. An entirely empty file is also a valid
gitconfig (it defines no settings).

### Typical File

```
# user identity
[user]
	name = Pat Example
	email = pat@example.com

[core]
	autocrlf = input
	editor = vim

[alias]
	st = status
	co = checkout

[color "diff"]
	meta = yellow bold
	frag = magenta bold
	old = red
	new = green
```

This file shows: comment lines, multiple sections, a quoted subsection
(`color "diff"`) holding scoped variables, color-typed values, and the
section-then-variable line shape.

### Complex File

```
; ---------------------------------------------------------------
; gitconfig with every wrinkle of the syntax
; ---------------------------------------------------------------

# section with mixed-case name (case-insensitive)
[Core]
	# implicit boolean true (no = sign)
	bare
	# explicit boolean false
	filemode = false
	# integer with unit suffix
	packedGitLimit = 256m
	# pathname with tilde expansion
	excludesFile = ~/.config/git/ignore
	# pathname opting out of %(prefix) expansion
	hooksPath = ./%(prefix)/hooks

# quoted subsection, case-sensitive: "Origin" ≠ "origin"
[remote "Origin"]
	url = https://example.com/repo.git
	fetch = +refs/heads/*:refs/remotes/Origin/*

# multivalued variable: two values for core.gitProxy
[core]
	gitProxy = "ssh" for "kernel.org"
	gitProxy = default-proxy ; comment after value

# value with leading/trailing whitespace requires quoting
[user]
	signature = "  -- Pat Example  "

# value with embedded escapes
[advice]
	statusHints = "first line\nsecond line\twith tab"

# line continuation across multiple physical lines
[merge "smartmerge"]
	driver = /usr/local/bin/smartmerge \
	         --left %A \
	         --base %O \
	         --right %B \
	         --output %A

# deprecated dotted-subsection form (lowercased at parse time)
[branch.devel]
	remote = origin
	merge  = refs/heads/devel

# unconditional include
[include]
	path = ~/.gitconfig.shared        ; tilde-expanded
	path = ./fragments/aliases.inc    ; relative to this file

# conditional includes
[includeIf "gitdir:~/work/"]
	path = ~/.gitconfig.work
[includeIf "gitdir/i:C:/Users/Pat/"]
	path = ~/.gitconfig.windows
[includeIf "onbranch:release/"]
	path = ./release.inc
[includeIf "hasconfig:remote.*.url:https://github.com/**"]
	path = ./gh.inc

# subsection name with embedded escapes
[example "weird \"quoted\" \\name"]
	value = ok

# color-typed value
[color "decorate"]
	branch = green bold reverse
	# integer color (256-color mode) and attribute negation
	HEAD   = 14 noreverse
	# 24-bit RGB color, two-color form (fg bg) plus attributes
	tag    = #ff0ab3 #102030 italic
```

This document exercises: comments at top of file and end of line; mixed-case
section names; implicit-boolean (bare-name) variables; integer suffix `m`;
tilde and `%(prefix)` paths; quoted and case-sensitive subsections;
multivalued variables; quoted values with leading/trailing whitespace; in-
value escape sequences; backslash line continuation; the deprecated dotted
form; the four `includeIf` keywords (`gitdir`, `gitdir/i`, `onbranch`,
`hasconfig:remote.*.url`); subsection names containing escaped quotes and
backslashes; color values in name, integer, and `#RRGGBB` forms.

## Appendix

- **Encoding.** Not formally specified by `git-config(1)`. In practice files
  are read as bytes; ASCII and UTF-8 work and are common on real systems.
  A leading UTF-8 BOM is **not** part of the format and is treated as
  garbage bytes that will fail to lex. Implementations should reject a BOM.
- **Line endings.** LF and CRLF are accepted (CRLF is normalised to LF as
  the line terminator). Lone CR is unspecified; treat it as a value byte.
- **NUL bytes.** Forbidden inside subsection literals; behaviour elsewhere
  is unspecified. An implementation should reject any embedded NUL outside
  a quoted span and, conservatively, also inside one.
- **File size limits.** None specified.
- **Related references.**
  - `git-config(1)` man page (canonical), <https://git-scm.com/docs/git-config>
  - `gitignore(5)` for the glob extensions used by `includeIf`
  - `git-worktree(1)` for `extensions.worktreeConfig` and
    `$GIT_DIR/config.worktree`
- **Implementation notes for a Go parser.**
  - The lexer is line-oriented; `bufio.Scanner` with a large enough buffer
    is sufficient.
  - Quoted-span recognition must be done before comment recognition on the
    same logical line, so that `;` inside quotes is not mistaken for a
    comment.
  - Backslash line continuation must be handled before comment stripping,
    so that a backslash at end-of-line is not consumed by a stray
    `Whitespace`/`Comment` rule.
  - Subsection literals and value escape rules are *different* (subsection
    drops unknown `\X` to `X`; value rejects unknown `\X`). Two distinct
    state-machine actions, not one shared escape decoder.
- **Known compatibility caveats.**
  - `gitdir` matching changed between v2.13.0 and later; configs that need
    to work on both should specify both the symlink and realpath patterns.
  - The `[section.subsection]` deprecated form is still accepted by all
    current Git versions but should not be emitted by new printers.
  - The `hasconfig:remote.*.url` keyword exists for forwards compatibility
    with a future `hasconfig:<variable>` family; only the `remote.*.url`
    spelling is currently supported.
