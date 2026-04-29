# gitconfig File Format Specification

This document describes the syntax of Git's configuration file format
(`gitconfig`), an INI-dialect used by `~/.gitconfig`, `.git/config`,
`/etc/gitconfig`, and any file referenced through the `include` /
`includeIf` mechanism.

The authoritative reference is the **CONFIGURATION FILE** section of
`git-config(1)` (https://git-scm.com/docs/git-config). This spec is
intended as a parser-implementer's reference and is faithful to that
document; quotes from the manual are used where wording matters.

> **Scope.** This document describes the *file format* only. It does
> not describe the semantics of any particular variable, nor does it
> describe what `git config` does at runtime beyond what is needed to
> understand parsing.

---

## 1. Lexical Structure

### 1.1 Encoding and line endings

* Git config files are byte streams. There is no formal encoding
  declaration; UTF-8 is the de facto standard. The format itself only
  cares about ASCII characters (`[`, `]`, `=`, `"`, `\`, `#`, `;`,
  whitespace, and the newline).
* Lines are terminated by `\n`. Implementations should also accept
  `\r\n` and treat `\r` immediately before `\n` as part of the line
  terminator.
* The file MUST NOT contain a NUL byte (`\0`) inside a subsection
  name; behaviour for NUL bytes elsewhere is unspecified.

> **Ambiguity.** The Git manual does not formally specify line-ending
> handling. Different Git releases and third-party clients (libgit2,
> JGit, Go's `go-git`) have all been observed to differ on whether a
> bare `\r` is treated as whitespace or as a line terminator. A
> conservative parser should accept `\r\n` and `\n`, and treat a bare
> `\r` as ordinary whitespace.

### 1.2 Whitespace

The manual states:

> Whitespace characters, which in this context are the space
> character (SP) and the horizontal tabulation (HT), are mostly
> ignored.

The word *mostly* is doing a lot of work. Concretely:

* Whitespace around section headers, around the `=` separator, and
  at the start/end of an unquoted value is stripped.
* Whitespace **inside** an unquoted value is preserved verbatim
  (so `pager = less -S` produces the value `less -S`).
* A bare variable name (no `=`) is allowed; trailing whitespace on
  that line is discarded.

> **Ambiguity.** "Mostly ignored" is not a precise rule. Implementers
> should treat SP and HT as significant only inside quoted strings
> and inside the body of an unquoted value; everywhere else they are
> token separators / padding.

### 1.3 Comments

Two comment introducer characters are recognised: `#` and `;`.

* A comment begins at `#` or `;` and continues to the end of the
  line.
* Comments are recognised at the top level (between sections), on
  the same line as a section header (after the closing `]`), and on
  the same line as a variable assignment (after the value).
* A `#` or `;` that appears **inside a double-quoted string** is a
  literal character, not a comment introducer.
* A `#` or `;` that appears in an **unquoted** value terminates the
  value. Anything from that character to end-of-line is the comment.

```ini
# top-level comment
; also a comment
[core]            # trailing comment after section header
    pager = less  ; trailing comment after value
    msg = "use # and ; freely inside quotes"
```

> **Ambiguity.** The manual does not explicitly say whether a `#`
> appearing immediately after an unescaped backslash continuation is
> a comment on the *new* line or part of the continued value.
> Practically, line continuation joins the lines first, and only
> then is comment scanning applied; some hand-rolled parsers do this
> in the opposite order. Pick one and document it.

### 1.4 Line continuation

A line ending with an **odd** number of trailing backslashes (i.e. a
single un-escaped `\` immediately before the line terminator) is
continued onto the next line. The backslash and the following
newline are discarded; the next line's contents are appended to the
value.

```ini
[http "https://example.com"]
    cookieFile = \
        /tmp/cookies.txt
```

* Line continuation is meaningful only inside the value of a
  variable assignment. A trailing backslash on a section header,
  comment-only line, or blank line is undefined behaviour.
* Inside a double-quoted string, a backslash before a newline is
  *not* a line-continuation: the newline ends the string, which is
  an error. (Multi-line quoted strings are not supported.)

> **Ambiguity.** The manual is not explicit about whether
> continuation works inside quoted values; the safe rule is "no".

---

## 2. Section Headers

Every variable belongs to a section. A section is introduced by a
section header on its own line; it remains in effect until the next
section header (or end of file).

There are two forms.

### 2.1 Quoted-subsection form (preferred)

```
[<section> "<subsection>"]
```

* `<section>` is a name composed of ASCII alphanumerics, `-`, and
  `.`.
* `<section>` is **case-insensitive** (it is normalised to
  lowercase for matching).
* The subsection is enclosed in double quotes and separated from
  the section name by one or more SP/HT.
* `<subsection>` may contain any byte **except newline and NUL**.
  A literal `"` must be escaped as `\"`; a literal `\` must be
  escaped as `\\`.
* `<subsection>` is **case-sensitive**.
* Backslashes in subsection names that are not part of `\"` or
  `\\` are silently dropped: `\t` becomes `t`, `\0` becomes `0`.
  No escape sequences (other than `\"` and `\\`) are recognised
  inside a subsection name.
* The header must fit on a single line; line continuation is not
  allowed inside a header.

```ini
[remote "origin"]
[branch "feature/login"]
[url "git@github.com:"]
```

### 2.2 Dotted form (deprecated)

```
[<section>.<subsection>]
```

* The subsection is everything after the first `.`.
* It is restricted to the same character set as section names
  (alphanumerics, `-`, `.`).
* The subsection portion is **lower-cased** by Git and then matched
  case-sensitively. (i.e. `[Foo.Bar]` is normalised to `[foo.bar]`
  internally and a key under `Foo.Bar` is found via `foo.bar`.)
* This form is documented as deprecated; emit the quoted form when
  printing.

> **Ambiguity.** The combination "lower-case then match
> case-sensitively" is unusual and easy to get wrong. Some clients
> in the wild treat the dotted form as fully case-insensitive on the
> subsection. Parsers should follow the manual: lowercase on read.

### 2.3 Section without a subsection

```
[<section>]
```

The bare form has no subsection. `[core]` is the canonical example.

### 2.4 Header equivalence

`[Section "Sub"]` and `[section "Sub"]` refer to the same section
(section name is case-insensitive, subsection is case-sensitive).
`[Section.Sub]` refers to `section.sub` (subsection lowercased).

---

## 3. Variable Names

```
<name> = <value>
<name>
```

* A variable name is composed of ASCII alphanumerics and `-`.
* It MUST start with an alphabetic character.
* It is **case-insensitive** (normalised to lowercase for matching).
* A variable may appear multiple times within the same section;
  the format itself does not deduplicate. Whether a particular
  consumer treats the variable as last-wins or as multi-valued is a
  semantic decision made per-key by Git, not by the parser.
* A variable assignment with no `=` sign and no value is shorthand
  for the boolean value `true`:

  ```ini
  [core]
      bare           # equivalent to: bare = true
  ```

  This is the *only* place where the absence of `=` is meaningful.

> **Ambiguity.** Trailing whitespace after a bare variable name is
> ignored. A `=` followed by nothing (e.g. `name =`) means *empty
> string* in general, but is interpreted as boolean *false* when the
> consumer asks for a boolean. The lexical layer should preserve
> "empty string" and let the type layer handle the boolean coercion.

---

## 4. Values

### 4.1 Raw value lexing

After the `=`, the value extends to end-of-line, with the following
transformations applied in order:

1. **Line continuation.** Any line ending in a single trailing `\`
   is joined with the next line; the `\` and following `\n` are
   removed.
2. **Comment stripping.** Outside of a double-quoted region, the
   first un-escaped `#` or `;` and everything after it is dropped.
3. **Whitespace trimming.** Leading and trailing SP/HT outside
   quotes are removed. Internal whitespace is preserved.
4. **Quote processing.** Pairs of unescaped `"` mark
   double-quoted regions. Inside a quoted region, escape sequences
   are interpreted (see below) and characters that would otherwise
   be syntactically significant (`#`, `;`, leading/trailing
   whitespace, `\`-continuation rules above) are taken literally
   except for the listed escapes.

There is no single-quote string. There is no triple-quoted string.
There is no heredoc. The value is a single logical line after
continuation.

### 4.2 Escape sequences

Inside a double-quoted string, the following escapes are valid:

| Sequence | Meaning                  |
|----------|--------------------------|
| `\n`     | Newline (LF, U+000A)     |
| `\t`     | Horizontal tab (U+0009)  |
| `\b`     | Backspace (U+0008)       |
| `\"`     | Literal `"`              |
| `\\`     | Literal `\`              |

The manual says explicitly:

> Other char escape sequences (including octal escape sequences)
> are invalid.

What "invalid" means in practice has varied across Git versions.
Modern Git rejects unknown escapes with an error; older versions
silently dropped the backslash. A new parser SHOULD reject unknown
escapes, but MAY emit a warning and pass through the following
character to be permissive.

> **Ambiguity.** The same escape table is documented for subsection
> names but only `\"` and `\\` are honoured there; other backslash
> sequences are silently de-backslashed. Treat the two contexts
> separately.

### 4.3 Unquoted vs quoted values

Unquoted:

```ini
pager = less -S
```

The value is the literal string `less -S`. Internal whitespace is
preserved; leading/trailing whitespace is trimmed; a trailing
comment is stripped.

Quoted:

```ini
attributesfile = "  ~/.config/git/attrs  "
editor         = "C:\\Program Files\\vim\\vim.exe"
greeting       = "hello\tworld"
```

Quotes preserve leading/trailing whitespace and allow escape
sequences. A value may also be a *mix* of quoted and unquoted
runs on the same line, which are concatenated:

```ini
proxy = http://proxy"."example".com"   # → http://proxy.example.com
```

> **Ambiguity.** The mixed quoted/unquoted form is rarely used and
> rarely tested by other clients. Implementations differ on whether
> whitespace between an unquoted run and the next `"` is preserved
> or trimmed. Prefer to forbid this form when emitting.

### 4.4 Empty values

* `name =`        → empty string.
* `name = ""`     → empty string (explicit).
* `name`          → boolean true (no equals sign).

Distinguishing the first two from the third is required: the
absence of `=` is what triggers the implicit-true interpretation.

---

## 5. Value Types (Interpretation Layer)

Type interpretation is **not part of file parsing** — every value is
a string at the lexical layer. Type coercion happens when a consumer
calls something like `git_config_get_bool()`. A parser library
should expose both the raw string and helpers for these types.

### 5.1 Boolean

The following tokens (case-insensitive) are recognised:

| True                        | False                       |
|-----------------------------|-----------------------------|
| `yes`, `on`, `true`, `1`    | `no`, `off`, `false`, `0`   |
| variable present without `=`| empty string (`name =`)     |

Any other value is an error when interpreted as a boolean.

### 5.2 Integer

Decimal integer with an optional binary-scale suffix:

```
<digits>[k|K|m|M|g|G|t|T|p|P|e|E]
```

* `k` = ×1024, `m` = ×1024², `g` = ×1024³, etc.
* Suffix letters are case-insensitive.
* Sign is permitted (`-1`, `+10`).

> **Ambiguity.** The manual examples show `k`, `M`, `G` only.
> Practical Git also accepts `t`, `p`, `e`. Document only `k/m/g`
> if conservative; accept all when permissive.

### 5.3 Color

A color value is a sequence of up to two color tokens followed by
zero or more attribute tokens, all whitespace-separated.

* **Color tokens (each):** one of
  * a basic name: `normal`, `default`, `black`, `red`, `green`,
    `yellow`, `blue`, `magenta`, `cyan`, `white`;
  * a *bright* variant: `brightred`, `bright-red`, etc.;
  * a 0–255 ANSI integer;
  * a 24-bit hex `#rrggbb`;
  * a 12-bit hex `#rgb` (which expands to `#rrggbb`).
* **Attributes:** `bold`, `dim`, `ul` (underline), `blink`,
  `reverse`, `italic`, `strike`. Each may be prefixed with `no` or
  `no-` to turn it off (`nobold`, `no-reverse`).
* The pseudo-attribute `reset` clears prior color/attribute state.

Order: `[<fg> [<bg>]] [<attr>...]`. Either color may be omitted by
writing `normal` or by reordering, but two colors at most.

> **Ambiguity.** The `bright-` variants vs concatenated form
> (`brightred` vs `bright red` vs `bright-red`) are accepted by Git
> but third-party tools sometimes only accept one form. Be liberal
> on input.

### 5.4 Path (path-expandable)

A path value receives the following expansions when interpreted by
Git:

* Leading `~/`     → `$HOME/`.
* Leading `~user/` → home directory of `user`.
* Leading `%(prefix)/` → Git's compiled-in prefix (install dir).

A leading `:` (colon) is an *option marker* meaning "treat as
optional" — used primarily by `include.path`. The parser should
preserve the marker and let the include layer act on it.

Path values that do not begin with one of the above prefixes are
taken literally; tilde inside the path (not at the start) is
**not** expanded.

> **Ambiguity.** Whether path expansion happens at parse time or at
> use time is implementation-defined. The reference implementation
> expands lazily.

### 5.5 String

The fallback type. The raw post-quote-processing value is the
string. Strings may contain embedded newlines (via `\n`) and tabs.

---

## 6. The `include` Mechanism

The variable `include.path` causes the named file to be read as
though its contents were spliced in at the location of the
`include` directive.

```ini
[include]
    path = /etc/gitconfig.d/work.inc      ; absolute
    path = work.inc                        ; relative to *this* file
    path = ~/extra.gitconfig                ; tilde expansion
    path = %(prefix)/etc/gitconfig.shared
```

Rules:

* Relative paths are resolved relative to the **directory of the
  including file**, not the current working directory.
* `~/`, `~user/`, and `%(prefix)/` expansions all apply.
* Multiple `include.path` entries are processed in order.
* Includes nest; an included file may itself contain `include`
  directives.
* A missing file referenced by `include.path` is an **error**
  unless the path begins with `:` (the optional marker), in which
  case it is silently skipped.

> **Ambiguity.** The manual does not explicitly bound include
> recursion depth. The reference implementation has a hard cap
> (currently 10 levels) to prevent loops. Parsers SHOULD enforce a
> finite recursion depth and surface a clear error when exceeded.

---

## 7. The `includeIf` Mechanism

```ini
[includeIf "<condition>"]
    path = <file>
```

The subsection holds the condition string. If the condition
evaluates to true, the file is included exactly as for `include`;
otherwise the directive is silently ignored.

Four condition prefixes are defined.

### 7.1 `gitdir:<pattern>`

True when the absolute path of the current repository's `.git`
directory matches `<pattern>` as a glob.

Glob conveniences specific to this matcher:

* `~/` at the start expands to `$HOME/`.
* `./` at the start expands to the directory containing the
  including config file.
* If `<pattern>` does not start with `/`, `~/`, or `./`, an
  implicit `**/` is prepended (so `foo/` matches any directory
  named `foo` anywhere).
* If `<pattern>` ends with `/`, an implicit `**` is appended (so
  the pattern matches the directory and everything inside).

Glob metacharacters supported: `*`, `?`, `[...]`, `**`.

> **Ambiguity.** "Absolute path of the `.git` directory" needs
> care:
> * For a linked worktree, this is the `gitdir` line inside the
>   `.git` file, not the directory containing it.
> * For a submodule, this is the path under the superproject's
>   `.git/modules/...`.
> * Symlinks inside `$GIT_DIR` are **not** resolved before
>   matching; symlinks **outside** are matched both as the
>   symlink path and as the realpath. Implementations SHOULD
>   document which they do.

### 7.2 `gitdir/i:<pattern>`

Identical to `gitdir:` but matches case-insensitively. Intended for
case-insensitive filesystems (Windows, macOS HFS+/APFS-default).

### 7.3 `onbranch:<pattern>`

True when HEAD points to a branch whose short name matches
`<pattern>` as a glob. The same `**` and trailing-`/` conveniences
as `gitdir:` apply.

* When HEAD is detached (no current branch), the condition is
  false.
* The pattern matches the *short* branch name (e.g. `main`,
  `feature/login`), not the full ref (`refs/heads/main`).

### 7.4 `hasconfig:remote.*.url:<url-pattern>`

True when *any* remote URL in the in-progress config matches the
glob `<url-pattern>`.

This condition is special:

* It must be evaluated **after** at least an initial scan of the
  config has produced remote URLs. Git solves this by reading the
  config twice: first to gather `remote.*.url` values, then to
  process `includeIf` directives.
* The manual states that files included via `hasconfig:` MUST NOT
  themselves declare `remote.*.url` — doing so risks a
  chicken-and-egg situation. Parsers SHOULD detect and reject this.

> **Ambiguity.** Older Git versions did not implement
> `hasconfig:`. Tools that need to consume configs written for
> modern Git but parsed by older logic may need to fall back to
> `gitdir:`-based gates.

---

## 8. Grammar Summary (informal EBNF)

```
file              = { line } ;
line              = section_header | assignment | comment_line | blank_line ;

section_header    = "[" section_name [ ws subsection ] "]" [ comment ] EOL
                  | "[" section_name "." dotted_subsection "]" [ comment ] EOL ;
section_name      = alpha { alpha | digit | "-" | "." } ;
subsection        = '"' { subsection_char | "\\" '"' | "\\" "\\" } '"' ;
dotted_subsection = section_name ;     (* lowercased on read *)

assignment        = ws variable_name [ ws "=" ws value ] [ comment ] EOL ;
variable_name     = alpha { alpha | digit | "-" } ;
value             = { value_run } ;
value_run         = unquoted_run | quoted_run ;
unquoted_run      = { any char except '"', '#', ';', NL, unescaped \ at EOL } ;
quoted_run        = '"' { quoted_char | escape } '"' ;
escape            = "\\n" | "\\t" | "\\b" | "\\\"" | "\\\\" ;

comment_line      = ws ( "#" | ";" ) { not NL } EOL ;
comment           = ( "#" | ";" ) { not NL } ;
blank_line        = ws EOL ;

ws                = { SP | HT } ;
EOL               = "\n" | "\r\n" ;
```

The grammar above is informal — line continuation, value
concatenation across mixed quoted/unquoted runs, and comment
stripping interact with whitespace handling in ways that the EBNF
abstracts over. See section 4.1 for the precise lexing pipeline.

---

## 9. Implementer Checklist

A parser is conformant if it:

1. Accepts both `[section "subsection"]` and the deprecated
   `[section.subsection]` forms.
2. Lower-cases section names and variable names before matching;
   preserves subsection case for the quoted form; lower-cases
   subsection case for the dotted form.
3. Strips comments introduced by `#` and `;` outside double-quoted
   strings.
4. Handles trailing-`\` line continuation only in unquoted-value
   context.
5. Preserves internal whitespace in unquoted values; trims
   leading/trailing.
6. Recognises exactly the escape set `\n \t \b \" \\` inside
   double-quoted strings, and rejects (or at minimum warns on)
   other backslash sequences.
7. Treats a bare variable name (no `=`) as the *string with no
   value* and supplies a boolean-true coercion to consumers.
8. Distinguishes `name =` (empty string) from `name` (no value).
9. Supports `include.path` with relative-to-including-file
   resolution and `~`/`%(prefix)` expansion.
10. Supports `includeIf "<cond>"` with the four condition kinds
    `gitdir`, `gitdir/i`, `onbranch`, `hasconfig:remote.*.url`,
    including the implicit `**/` prefix and `**` suffix glob
    conveniences.
11. Bounds include recursion to prevent loops.

A parser is **non-conformant** if it silently accepts:
* unterminated `[` headers,
* unterminated `"` strings,
* a section header that spans lines,
* unknown escape sequences without a warning,
* values containing a literal newline outside of `\n`.

---

## 10. Open Ambiguity Index

This list collects the places where the manual is loose or where
real-world clients have been observed to disagree. Implementers
should make an explicit decision for each and document it.

| # | Topic                                     | Section |
|---|-------------------------------------------|---------|
| 1 | `\r\n` vs bare `\r` handling              | 1.1     |
| 2 | Meaning of "whitespace mostly ignored"    | 1.2     |
| 3 | Comment-vs-continuation interaction       | 1.3     |
| 4 | Continuation inside quoted strings        | 1.4     |
| 5 | Dotted subsection lowercase + cs match    | 2.2     |
| 6 | `name =` empty string vs boolean false    | 3       |
| 7 | Mixed quoted/unquoted runs                | 4.3     |
| 8 | Modern reject vs legacy silent-pass for   |         |
|   | unknown escapes                           | 4.2     |
| 9 | Integer suffixes beyond k/M/G             | 5.2     |
| 10 | `bright-` color naming variants          | 5.3     |
| 11 | Path expansion timing (parse vs use)     | 5.4     |
| 12 | Include recursion depth limit            | 6       |
| 13 | Symlink resolution in `gitdir:` matching | 7.1     |
| 14 | Detached HEAD with `onbranch:`           | 7.3     |
| 15 | `hasconfig:` two-pass evaluation         | 7.4     |

When in doubt, follow what the reference Git implementation does at
the version you are targeting; cross-check against libgit2 if you
need to interoperate with non-Git clients.
