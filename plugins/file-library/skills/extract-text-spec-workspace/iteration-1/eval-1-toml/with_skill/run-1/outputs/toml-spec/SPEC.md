# TOML Specification Reference

## Overview

TOML ("Tom's Obvious, Minimal Language") is a minimal, line-oriented
configuration file format that maps unambiguously onto a hash table. This
reference covers **TOML v1.0.0** (released 2021-01-11), authored by Tom
Preston-Werner, Pradyun Gedam et al.

- Canonical prose specification: <https://toml.io/en/v1.0.0>
- Canonical formal grammar: <https://github.com/toml-lang/toml/blob/1.0.0/toml.abnf>
  (ABNF per RFC 5234)
- File extension: `.toml`
- MIME type: `application/toml`
- Encoding: TOML documents **must** be valid UTF-8.
- Case sensitivity: TOML is case-sensitive (keys, keywords, hex digits in
  strings; hex digits in integer/float literals are case-insensitive — see
  Lexical Elements).

The grammar notation used throughout this document is **ABNF (RFC 5234)**, the
notation used by the canonical grammar. Production names are kebab-case when
quoted from the source ABNF (e.g. `ml-basic-string`); this document also
provides PascalCase names (e.g. `MultiLineBasicString`) intended as AST type /
parser-action identifiers for the implementer. The two naming systems map
1:1 — every PascalCase name in this document is the kebab-case ABNF
production renamed.

Stable terminology used across this document:

- **table** (not "object", "map", or "dictionary") — a collection of key/value
  pairs.
- **array of tables** (not "table array") — the `[[name]]` construct.
- **key** — the left-hand side of a key/value pair; may be bare, quoted, or
  dotted.
- **field** — the keyword used for "a key/value pair belonging to a table" in
  prose; the parser-level term is **key/value pair** (or **keyval**).
- **production** (not "rule" or "nonterminal") — an ABNF grammar production.
- **token class** — a category of terminal the tokenizer emits (PascalCase).

## Lexical Elements (Tokens)

The tokenizer classifies every byte of a UTF-8-encoded TOML document into one
of the token classes below. TOML is line-oriented: newlines are significant
between expressions, and the tokenizer must distinguish `\n` (LF) from
`\r\n` (CRLF) only insofar as both are recognized as a single `Newline`
token; bare `\r` is **not** a valid newline.

For each class: name → ABNF (or pattern fragment) → example → edge cases.

### Comments

- **Name:** `Comment`
- **Pattern (ABNF):**
  ```abnf
  comment-start-symbol = %x23 ; #
  non-ascii            = %x80-D7FF / %xE000-10FFFF
  non-eol              = %x09 / %x20-7F / non-ascii
  comment              = comment-start-symbol *non-eol
  ```
- **Example:** `# this is a comment`
- **Placement:** A `#` outside of a string literal starts a comment that
  extends to (but does not include) the next newline or EOF. Comments may
  appear:
  - On their own line.
  - At the end of a key/value pair line, after the value.
  - Between values inside a multi-line array (within `ws-comment-newline`).
  - Between table headers, inline-table values are **not** allowed to contain
    a comment because newlines are forbidden inside an inline table.
- **Inside strings:** A `#` inside any of the four string literal forms is a
  literal `#`, not a comment.
- **Nesting:** TOML comments do **not** nest. `# # still one comment` is a
  single comment.
- **Edge cases:**
  - Control characters U+0000–U+0008, U+000A–U+001F, and U+007F are
    **forbidden** in comments. (U+000A is the terminating LF — it ends the
    comment but is not part of it; the rule above forbids it appearing
    *inside* the comment body.)
  - U+0009 (tab) is allowed.
  - Surrogate code points (U+D800–U+DFFF) are forbidden by the `non-ascii`
    rule.

### Whitespace and Delimiters

- **Whitespace token class:** `Whitespace`
  ```abnf
  ws     = *wschar
  wschar = %x20  ; Space
  wschar =/ %x09 ; Horizontal tab
  ```
  - Whitespace inside a single line is **insignificant** between tokens
    (around `=`, around `.` in dotted keys, around `[`, `]`, `[[`, `]]`,
    `{`, `}`, and `,`).
  - Inside basic and literal strings (single- and multi-line), whitespace is
    **significant** (it is part of the string value).
  - Inside comments, whitespace is part of the comment body.
  - Indentation has no semantic meaning — it is whitespace and ignored.

- **Newline token class:** `Newline`
  ```abnf
  newline =  %x0A    ; LF
  newline =/ %x0D.0A ; CRLF
  ```
  - `\r\n` is a single `Newline` token (a CR followed by an LF), **not** two
    tokens.
  - A bare `\r` (CR not followed by LF) is **not** a newline and is **not**
    permitted as one. It is permitted inside basic and multi-line basic
    strings only via the `\r` escape (which produces a literal U+000D).
  - A newline (or EOF) is required after every key/value pair and after
    every table header. (See `## Structure` for the formal rule.)

- **Structural delimiters:**

  | Token class           | Lexeme  | Codepoint | Role                                      |
  |-----------------------|---------|-----------|-------------------------------------------|
  | `Equals`              | `=`     | U+003D    | Separator between key and value           |
  | `Dot`                 | `.`     | U+002E    | Separator between simple-keys in a dotted key, **and** the decimal point in floats and dotted parts in dates/times |
  | `Comma`               | `,`     | U+002C    | Separator between array elements / inline-table entries |
  | `LeftBracket`         | `[`     | U+005B    | Opens an array literal **or** a standard table header |
  | `RightBracket`        | `]`     | U+005D    | Closes an array literal **or** a standard table header |
  | `DoubleLeftBracket`   | `[[`    | 2× U+005B | Opens an array-of-tables header           |
  | `DoubleRightBracket`  | `]]`    | 2× U+005D | Closes an array-of-tables header          |
  | `LeftBrace`           | `{`     | U+007B    | Opens an inline table                     |
  | `RightBrace`          | `}`     | U+007D    | Closes an inline table                    |

  > **Ambiguity:** `[` may begin either an array literal (in value position)
  > or a standard table header (in expression position). The two are
  > distinguished by *parser context*, not by the tokenizer. Likewise the
  > tokenizer must decide whether `[[` is two `LeftBracket` tokens (the start
  > of a nested array literal `[[1,2]]`) or a single `DoubleLeftBracket`
  > token (an array-of-tables header). The simplest rule: in expression
  > position (start-of-line after optional whitespace), `[[` is the
  > array-of-tables opener; in value position, `[[` is two `LeftBracket`
  > tokens. The implementer should treat `[[`/`]]` as table-header tokens
  > only at the top of an expression.

### Literals

#### Strings

There are four string forms; each is its own token class.

##### `BasicString` (single-line, double-quoted, escapes interpreted)

```abnf
basic-string    = quotation-mark *basic-char quotation-mark
quotation-mark  = %x22  ; "
basic-char      = basic-unescaped / escaped
basic-unescaped = wschar / %x21 / %x23-5B / %x5D-7E / non-ascii
escaped         = escape escape-seq-char
escape          = %x5C  ; \
escape-seq-char =  %x22         ; \"   U+0022
escape-seq-char =/ %x5C         ; \\   U+005C
escape-seq-char =/ %x62         ; \b   U+0008  backspace
escape-seq-char =/ %x66         ; \f   U+000C  form feed
escape-seq-char =/ %x6E         ; \n   U+000A  line feed
escape-seq-char =/ %x72         ; \r   U+000D  carriage return
escape-seq-char =/ %x74         ; \t   U+0009  tab
escape-seq-char =/ %x75 4HEXDIG ; \uXXXX     unicode BMP scalar
escape-seq-char =/ %x55 8HEXDIG ; \UXXXXXXXX unicode scalar
```

- **Example:** `"I'm a string. \"You can quote me\". Name\tJosé\nLocation\tSF."`
- **Edge cases:**
  - Must not contain a literal newline (single-line only).
  - Must not contain unescaped control characters U+0000–U+001F or U+007F
    (excluding U+0009 — tab is allowed unescaped).
  - All escape sequences not listed above are reserved and **must** error.
    In particular, `\x`, `\0`, `\'`, `\v`, `\a` are invalid in TOML.
  - `\uXXXX` and `\UXXXXXXXX` must reference a valid Unicode scalar value
    (U+0000–U+D7FF or U+E000–U+10FFFF — surrogates are rejected).
  - Empty string `""` is valid.

##### `MultiLineBasicString` (triple-double-quoted, multi-line, escapes interpreted)

```abnf
ml-basic-string       = ml-basic-string-delim [ newline ] ml-basic-body
                        ml-basic-string-delim
ml-basic-string-delim = 3quotation-mark
ml-basic-body         = *mlb-content *( mlb-quotes 1*mlb-content ) [ mlb-quotes ]
mlb-content           = mlb-char / newline / mlb-escaped-nl
mlb-char              = mlb-unescaped / escaped
mlb-quotes            = 1*2quotation-mark
mlb-unescaped         = wschar / %x21 / %x23-5B / %x5D-7E / non-ascii
mlb-escaped-nl        = escape ws newline *( wschar / newline )
```

- **Example:**
  ```
  str = """
  Roses are red
  Violets are blue"""
  ```
- **Edge cases:**
  - A newline that immediately follows the opening `"""` is **trimmed**.
  - All other newlines and whitespace are preserved verbatim.
  - The implementation **may** normalize newlines to the platform default
    (LF on Unix, CRLF on Windows). This is observable round-trip behavior;
    the printer should document its choice.
  - **Line-ending backslash:** A `\` as the last non-whitespace char on a
    line, followed by zero or more whitespace, the line terminator, and any
    further whitespace/newlines, is consumed entirely (the `mlb-escaped-nl`
    production). This lets a long string be split across lines without
    introducing whitespace into the value.
  - 1 or 2 quotation marks may appear anywhere inside the body, including
    just inside the closing delimiter (so `""""..."""""""` is parseable but
    confusing). 3+ consecutive `"` characters delimit the end; to encode 3+
    consecutive `"` in the value, escape one or more of them (e.g.
    `""\"`).
  - All `BasicString` escape sequences are valid here.
  - Forbidden raw control characters: U+0000–U+0008, U+000B, U+000C,
    U+000E–U+001F, U+007F. (U+0009 tab, U+000A LF, U+000D CR are allowed
    raw — though CR raw is only allowed as part of CRLF.)

##### `LiteralString` (single-line, single-quoted, no escapes)

```abnf
literal-string = apostrophe *literal-char apostrophe
apostrophe     = %x27 ; '
literal-char   = %x09 / %x20-26 / %x28-7E / non-ascii
```

- **Example:** `winpath = 'C:\Users\nodejs\templates'`
- **Edge cases:**
  - **No escapes whatsoever.** The bytes between the surrounding `'` are the
    string value verbatim.
  - Cannot contain a `'` (U+0027) — there is no escape mechanism, so a
    literal apostrophe inside a literal string requires the multi-line
    form.
  - Cannot contain a literal newline.
  - Cannot contain control characters U+0000–U+0008, U+000A–U+001F, U+007F
    (tab U+0009 is allowed).

##### `MultiLineLiteralString` (triple-single-quoted, multi-line, no escapes)

```abnf
ml-literal-string       = ml-literal-string-delim [ newline ] ml-literal-body
                          ml-literal-string-delim
ml-literal-string-delim = 3apostrophe
ml-literal-body         = *mll-content *( mll-quotes 1*mll-content ) [ mll-quotes ]
mll-content             = mll-char / newline
mll-char                = %x09 / %x20-26 / %x28-7E / non-ascii
mll-quotes              = 1*2apostrophe
```

- **Example:**
  ```
  regex2 = '''I [dw]on't need \d{2} apples'''
  ```
- **Edge cases:**
  - No escape processing — literal bytes between the delimiters.
  - A newline immediately following the opening `'''` is **trimmed**.
  - 1 or 2 single quotes anywhere in the body are fine. 3+ single quotes are
    not representable inside a multi-line literal string at all (you must
    use a basic string with `'` escapes, or split the string).
  - Forbidden raw control characters: same exclusions as `MultiLineBasicString`
    — only `\t`, `\n`, `\r` (the latter only as part of CRLF) and printable
    bytes are allowed.

#### Integer

- **Token class:** `Integer`
- **Pattern (ABNF):**
  ```abnf
  integer          = dec-int / hex-int / oct-int / bin-int

  minus            = %x2D                   ; -
  plus             = %x2B                   ; +
  underscore       = %x5F                   ; _
  digit1-9         = %x31-39                ; 1-9
  digit0-7         = %x30-37                ; 0-7
  digit0-1         = %x30-31                ; 0-1

  hex-prefix       = %x30.78                ; 0x
  oct-prefix       = %x30.6F                ; 0o
  bin-prefix       = %x30.62                ; 0b

  dec-int          = [ minus / plus ] unsigned-dec-int
  unsigned-dec-int = DIGIT / digit1-9 1*( DIGIT / underscore DIGIT )

  hex-int          = hex-prefix HEXDIG     *( HEXDIG   / underscore HEXDIG   )
  oct-int          = oct-prefix digit0-7   *( digit0-7 / underscore digit0-7 )
  bin-int          = bin-prefix digit0-1   *( digit0-1 / underscore digit0-1 )

  HEXDIG           = DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
  ```
- **Examples:**
  - Decimal: `+99`, `42`, `0`, `-17`, `1_000`, `5_349_221`, `1_2_3_4_5`
  - Hex: `0xDEADBEEF`, `0xdead_beef`
  - Octal: `0o01234567`, `0o755`
  - Binary: `0b11010110`
- **Edge cases:**
  - Decimal: leading zeros forbidden (so `01` is **not** a valid integer).
    `-0` and `+0` are valid and equal to `0`.
  - Hex digits A–F are case-insensitive (`0xdead`, `0xDEAD`, `0xDeAd` all
    valid and equal); the prefix letters `x`, `o`, `b` themselves must be
    lowercase.
  - Underscores must each be flanked by at least one digit on each side —
    `_1`, `1_`, `1__2`, `0x_FF`, `0xFF_` are all invalid.
  - For `hex-int`, `oct-int`, `bin-int`: a leading sign (`+`/`-`) is **not**
    permitted, and leading zeros after the prefix **are** permitted.
  - Range: implementations must accept and losslessly handle the full signed
    64-bit range −2^63 .. 2^63−1. Values outside that range must error.

#### Float

- **Token class:** `Float`
- **Pattern (ABNF):**
  ```abnf
  float                = float-int-part ( exp / frac [ exp ] )
  float                =/ special-float

  float-int-part       = dec-int
  frac                 = decimal-point zero-prefixable-int
  decimal-point        = %x2E ; .
  zero-prefixable-int  = DIGIT *( DIGIT / underscore DIGIT )

  exp                  = "e" float-exp-part
  float-exp-part       = [ minus / plus ] zero-prefixable-int

  special-float        = [ minus / plus ] ( inf / nan )
  inf                  = %x69.6e.66 ; inf
  nan                  = %x6e.61.6e ; nan
  ```
- **Examples:**
  - Fractional: `+1.0`, `3.1415`, `-0.01`
  - Exponent: `5e+22`, `1e06`, `-2E-2`
  - Combined: `6.626e-34`
  - Underscored: `224_617.445_991_228`
  - Special: `inf`, `+inf`, `-inf`, `nan`, `+nan`, `-nan`
- **Edge cases:**
  - Decimal point must be flanked by at least one digit on each side.
    `.7`, `7.`, and `3.e+20` are all invalid.
  - Exponent character is `e` or `E` (case-insensitive — note the ABNF
    literal `"e"` is by RFC 5234 case-insensitive).
  - Exponent body may include leading zeros (it is `zero-prefixable-int`),
    unlike a bare decimal integer.
  - `+0.0` and `-0.0` are both valid; they map per IEEE 754.
  - `inf`, `nan` keywords are always lowercase. `Inf`, `NAN`, etc. are
    invalid.
  - The sign of `nan` is implementation-defined (the specific NaN payload
    is also implementation-defined).
  - Floats are IEEE 754 binary64 (`double`).
  > **Ambiguity:** ABNF case-insensitivity (`"e"` matches `e` and `E`) is a
  > standard RFC 5234 convention but is not always obvious. Implementers
  > should accept both `e` and `E` as the exponent introducer; the prose
  > spec confirms this.

#### Boolean

- **Token class:** `Boolean`
- **Pattern (ABNF):**
  ```abnf
  boolean = true / false
  true    = %x74.72.75.65    ; true
  false   = %x66.61.6C.73.65 ; false
  ```
- **Examples:** `true`, `false`
- **Edge cases:** Always lowercase. `True`, `TRUE`, `FALSE` are not
  booleans — they would be invalid identifiers in value position.

#### Date and Time (four token classes)

All TOML date-time literals are **subsets of RFC 3339**.

```abnf
date-fullyear  = 4DIGIT
date-month     = 2DIGIT  ; 01-12
date-mday      = 2DIGIT  ; 01-28, 01-29, 01-30, 01-31 by month/year
time-delim     = "T" / %x20  ; T, t, or SPACE
time-hour      = 2DIGIT  ; 00-23
time-minute    = 2DIGIT  ; 00-59
time-second    = 2DIGIT  ; 00-58, 00-59, 00-60 (leap second)
time-secfrac   = "." 1*DIGIT
time-numoffset = ( "+" / "-" ) time-hour ":" time-minute
time-offset    = "Z" / time-numoffset

partial-time   = time-hour ":" time-minute ":" time-second [ time-secfrac ]
full-date      = date-fullyear "-" date-month "-" date-mday
full-time      = partial-time time-offset
```

- **`OffsetDateTime`** (`offset-date-time = full-date time-delim full-time`)
  - Examples: `1979-05-27T07:32:00Z`, `1979-05-27T00:32:00-07:00`,
    `1979-05-27T00:32:00.999999-07:00`, `1979-05-27 07:32:00Z`.
  - The `T` between date and time may be a literal `T`/`t` or an ASCII
    space.

- **`LocalDateTime`** (`local-date-time = full-date time-delim partial-time`)
  - Example: `1979-05-27T07:32:00`, `1979-05-27T00:32:00.999999`.
  - No offset and no `Z`. Represents a calendar moment with no timezone
    information.

- **`LocalDate`** (`local-date = full-date`)
  - Example: `1979-05-27`.

- **`LocalTime`** (`local-time = partial-time`)
  - Example: `07:32:00`, `00:32:00.999999`.

- **Edge cases (all four):**
  - Millisecond precision is required (i.e. tokenizer must accept at least
    3 fractional-second digits when present).
  - More precision is implementation-defined; excess precision must be
    **truncated**, not rounded.
  - Leap seconds (`:60`) are syntactically valid.
  - `time-delim` matches both `T` (uppercase) and `t` (lowercase) per
    RFC 3339; the ABNF literal `"T"` is case-insensitive. A space is also
    permitted as the delimiter.
  - `time-offset` `Z` likewise matches both `Z` and `z` per ABNF
    case-insensitivity.

  > **Ambiguity:** The ABNF allows a leap-second `:60`, but a parser
  > targeting a non-leap-aware time library may need to either reject
  > `:60`, fold it to `:59.999999`, or carry it through verbatim. The spec
  > does not mandate any of these — pick a documented behavior.

  > **Ambiguity:** Bare 2-digit-month / 2-digit-day fields in the ABNF say
  > `01-28, 01-29, 01-30, 01-31 based on month/year`. The grammar itself
  > does not enforce the calendar — `2021-02-30` is syntactically accepted
  > by the ABNF but semantically invalid. The implementer must validate
  > calendar values during parsing.

### Keywords and Reserved Words

TOML's only reserved bareword tokens are the following. They are recognized
**only in value position**; the same character sequences are valid bare keys
in key position.

| Token class | Lexeme    | Notes                                     |
|-------------|-----------|-------------------------------------------|
| `Boolean`   | `true`    | Always lowercase.                         |
| `Boolean`   | `false`   | Always lowercase.                         |
| `Float`     | `inf`     | Always lowercase. May be prefixed `+`/`-`. |
| `Float`     | `nan`     | Always lowercase. May be prefixed `+`/`-`. |

There are no other reserved words. In particular, `null`, `none`, `nil`,
`undefined`, `yes`, `no`, `on`, `off`, `True`, `TRUE` are **not** TOML
literals — TOML has no null type.

### Symbols and Operators

TOML has no operators (no arithmetic, comparison, or assignment beyond `=`
itself). All "symbols" are the structural delimiters listed under
"Whitespace and Delimiters" above.

### Bare Keys

- **Token class:** `BareKey`
- **Pattern (ABNF):**
  ```abnf
  unquoted-key = 1*( ALPHA / DIGIT / %x2D / %x5F ) ; A-Z / a-z / 0-9 / - / _
  ALPHA        = %x41-5A / %x61-7A
  DIGIT        = %x30-39
  ```
- **Examples:** `key`, `bare_key`, `bare-key`, `1234`
- **Edge cases:**
  - Must be non-empty.
  - Composed only of ASCII letters, ASCII digits, `-`, and `_`.
  - All-digit bare keys (`1234`) are valid and are interpreted as **strings**
    — they are keys, not integers.

  > **Ambiguity:** A bare key composed of digits and a `.` (e.g.
  > `3.14159 = "pi"`) is a **dotted key** of two simple keys (`3` and
  > `14159`), **not** a single float. The tokenizer must decide based on
  > position: at the start of an expression (key position), `3.14` is a
  > dotted key; in value position, `3.14` is a float. The prose spec makes
  > this explicit.

### Quoted Keys

A `BasicString` or `LiteralString` token used in key position. Same lexical
rules as the corresponding string literal forms; semantics differ — see
`## Structure`.

## Structure (Grammar)

The parser consumes the token stream above and produces an AST whose nodes
correspond to the productions below. Every production is given in **ABNF**
(matching the canonical `toml.abnf`); each production carries a PascalCase
AST type name for the implementer.

### Top-Level Structure

A TOML document is a sequence of expressions separated by newlines.

```abnf
toml       = expression *( newline expression )

expression =  ws [ comment ]
expression =/ ws keyval ws [ comment ]
expression =/ ws table  ws [ comment ]
```

- **AST root:** `Document`
  - Fields:
    - `expressions` — ordered list of `Expression` nodes (multiplicity `*`).
- **Document-level constraints:**
  - The document must be valid UTF-8.
  - There is exactly one implicit root table; key/value pairs and tables
    encountered before any `[header]` belong to it.
  - A trailing newline at end of file is **not** required; the final
    expression need not be followed by a newline (the ABNF allows
    `*( newline expression )` so the last expression has no trailing
    newline). EOF terminates the last expression.
  - No BOM. The spec mandates UTF-8 but does not specify a BOM; a leading
    `EF BB BF` byte sequence would not match `expression` and is therefore
    invalid input.
  > **Ambiguity:** The spec and ABNF do not explicitly mention BOM
  > handling. Most TOML parsers strip an optional UTF-8 BOM at the very
  > start of the input as a courtesy; this is not mandated and is not
  > required by the grammar. Document the parser's choice.

### Grammar Productions

#### `Expression`

```abnf
expression =  ws [ comment ]
expression =/ ws keyval ws [ comment ]
expression =/ ws table  ws [ comment ]
```

- **AST:** `Expression` is a sum type:
  - `BlankExpression { trailingComment: Comment? }`
  - `KeyvalExpression { keyval: Keyval, trailingComment: Comment? }`
  - `TableExpression  { table: TableHeader, trailingComment: Comment? }`
- **Notes:**
  - Whitespace tokens around the body and after the final body token are
    consumed as `ws` and not represented in the AST.
  - A trailing comment on the same line attaches to the expression for
    round-trip preservation.

#### `Keyval`

```abnf
keyval     = key keyval-sep val
keyval-sep = ws %x3D ws ; =
```

- **AST:** `Keyval { key: Key, value: Value }`
- **Constraints:**
  - The key, the `=`, and the value must all reside on a single line
    (newlines inside `Value` are only permitted within multi-line strings,
    multi-line arrays, or value subtrees that themselves accept newlines).
  - The same key (after dotted-key normalization) cannot be defined twice
    in the same scope. (Enforced semantically — see `## Semantics`.)

#### `Key`

```abnf
key        = simple-key / dotted-key
simple-key = quoted-key / unquoted-key
quoted-key = basic-string / literal-string
dotted-key = simple-key 1*( dot-sep simple-key )
dot-sep    = ws %x2E ws ; . Period
```

- **AST:** `Key` is a sum type:
  - `SimpleKey { kind: BareKey | BasicString | LiteralString, raw: string, value: string }`
  - `DottedKey { parts: [SimpleKey, SimpleKey, ...] }` — at least two parts.
- **Notes:**
  - `value` of a `SimpleKey` is the **decoded** key string (e.g. `"a.b"` for
    a literal-string key spelled `'a.b'`); `raw` is the source form.
  - Whitespace around `dot-sep` is permitted but discouraged. The whitespace
    is discarded.
  - A bare key composed only of ASCII digits is still a **string** key, not
    a number — see `## Semantics`.

#### `Val`

```abnf
val = string / boolean / array / inline-table / date-time / float / integer
```

- **AST:** `Value` is a sum type:
  - `StringValue  { kind: BasicString | MultiLineBasicString | LiteralString | MultiLineLiteralString, value: string, raw: string }`
  - `IntegerValue { value: i64, raw: string, base: Decimal | Hex | Octal | Binary }`
  - `FloatValue   { value: f64, raw: string }`
  - `BooleanValue { value: bool }`
  - `OffsetDateTimeValue { … }`
  - `LocalDateTimeValue  { … }`
  - `LocalDateValue      { … }`
  - `LocalTimeValue      { … }`
  - `ArrayValue       { array: Array }`
  - `InlineTableValue { inlineTable: InlineTable }`

#### `Array`

```abnf
array              = array-open [ array-values ] ws-comment-newline array-close
array-open         = %x5B ; [
array-close        = %x5D ; ]
array-values       = ws-comment-newline val ws-comment-newline array-sep array-values
array-values       =/ ws-comment-newline val ws-comment-newline [ array-sep ]
array-sep          = %x2C ; , Comma
ws-comment-newline = *( wschar / [ comment ] newline )
```

- **AST:** `Array { elements: [Value, Value, ...] }`
- **Constraints:**
  - Elements are separated by `,`.
  - A trailing `,` is permitted after the last element.
  - An empty array `[]` is valid.
  - Comments and newlines may freely appear between elements (via
    `ws-comment-newline`) — arrays may span lines.
  - Element types are independent — TOML arrays are heterogeneous.

#### `Table` (sum type wrapping the two table-header forms)

```abnf
table = std-table / array-table
```

#### `StdTable` (standard table header)

```abnf
std-table       = std-table-open key std-table-close
std-table-open  = %x5B ws ; [
std-table-close = ws %x5D ; ]
```

- **AST:** `StdTableHeader { key: Key }`
- **Effect:** Subsequent key/value expressions belong to the table named by
  `key` (which may be dotted), until the next table header or EOF.
- **Constraints (semantic, see `## Semantics`):**
  - Every dotted-key prefix is implicitly created as a table if not already
    defined.
  - A table cannot be redefined.
  - A table cannot collide with an array-of-tables of the same name.

#### `ArrayTable` (array-of-tables header)

```abnf
array-table       = array-table-open key array-table-close
array-table-open  = %x5B.5B ws ; [[
array-table-close = ws %x5D.5D ; ]]
```

- **AST:** `ArrayTableHeader { key: Key }`
- **Effect:** Each occurrence of `[[name]]` appends a new empty table to
  the array at `name` (creating the array on first occurrence). Following
  key/value expressions belong to that newly appended table until the next
  table header or EOF.

#### `InlineTable`

```abnf
inline-table         = inline-table-open [ inline-table-keyvals ] inline-table-close
inline-table-open    = %x7B ws ; {
inline-table-close   = ws %x7D ; }
inline-table-sep     = ws %x2C ws ; ,
inline-table-keyvals = keyval [ inline-table-sep inline-table-keyvals ]
```

- **AST:** `InlineTable { entries: [Keyval, Keyval, ...] }`
- **Constraints:**
  - Newlines inside the braces are **not** permitted (except as part of an
    inner multi-line string value).
  - **No trailing comma** after the last `keyval` (unlike arrays).
  - Comments inside an inline table are **forbidden**, because comments
    must be terminated by a newline.
  - Inline tables are fully self-contained — no key may be added to an
    inline table from outside its braces.

#### `Comment`

```abnf
comment = comment-start-symbol *non-eol
```

- **AST:** `Comment { text: string }` (text excludes the leading `#`).
- A comment's content extends from the `#` (exclusive) to the next newline
  or EOF (exclusive).

### Ordering and Optionality

- **Required vs optional document elements:**
  - The empty document is valid (zero expressions).
  - Each expression is one of: blank/comment-only, key/value pair, or
    table header.
- **Ordering rules the grammar cannot express:**
  - **Single definition:** a given key (after dotted-key normalization)
    cannot be defined more than once in the same table scope. Both
    `[fruit]` followed by `[fruit]` and `name = "Tom"` followed by
    `name = "Pradyun"` are errors.
  - **Open vs closed tables:**
    - A table opened with a `[header]` is **closed** for direct redefinition
      and for reopening with a duplicate `[header]`. Sub-tables of the same
      parent may still be opened with `[parent.child]` later.
    - A table created **implicitly** (as a parent of a dotted key, or as a
      super-table inferred from a deeper `[a.b.c]` header) **may** be
      closed later with an explicit `[a]` header. The reverse is also
      true: `[a]` followed by `[a.b.c]` is fine.
    - An inline table is **always closed** the moment its `}` is seen. No
      later expression may add a key to it.
    - An array of tables created with `[[name]]` is closed for use as a
      normal table — `[[name]]` followed by `[name]` (single-bracket) is
      an error, and vice-versa.
  - **Out-of-order tables / dotted keys** are valid but discouraged
    stylistically; this affects style guides, not parser acceptance.
  - **Array element parent must exist:** `[fruit.physical]` followed by
    `[[fruit]]` is an error — the implicit `fruit` table created by
    `[fruit.physical]` cannot be retroactively reinterpreted as an array.
  - **Static array immutability:** `fruits = []` followed by `[[fruits]]`
    is an error — a value-position array cannot be appended to via the
    array-of-tables syntax.

## Semantics

Beyond syntactic acceptance, TOML imposes the following semantic rules. These
shape the AST and the parser's error reporting; the printer must respect
them when emitting.

### Mapping to a hash table

A TOML document is interpreted as a **hash table** (the root table). Each
table is itself a hash table; each key's value is one of: string, integer,
float, boolean, four flavors of date-time, array, or table.

- Key uniqueness within a table is **strict** — no duplicate keys, no
  silent overwriting.
- Key/value ordering is **not significant** at the data-model level. A
  conformant parser may expose document-order metadata, but two TOML
  documents with the same keys in different orders represent the same
  data.

### Key normalization

- Bare keys and quoted keys with the same character content are
  **equivalent**: `spelling = "x"` and `"spelling" = "x"` collide.
- Quoted keys decode escape sequences (basic-string keys) or take the
  contents verbatim (literal-string keys). Two keys are equal iff their
  decoded character sequences are equal.
- Bare keys that are all-digits (`1234`) are **strings**, not numbers.

### Dotted keys define implicit tables

Writing `fruit.apple.color = "red"` implicitly creates `fruit` and
`fruit.apple` as tables, then sets `color = "red"` inside
`fruit.apple`. These implicit tables may later be **opened** with their own
header (e.g. `[fruit.apple]`) provided no direct value collision exists.

- A dotted key cannot redefine a value:
  ```toml
  fruit.apple = 1
  fruit.apple.smooth = true   # error: cannot turn an integer into a table
  ```

### Table-header semantics

- `[a.b.c]` opens the table at path `a.b.c` for subsequent key/value
  pairs. It implicitly creates `a` and `a.b` if needed.
- A given header path may be opened **at most once** (with the exception
  noted below for arrays of tables and previously-implicit tables).
- A table may not be reopened by a *dotted-key prefix* if it has already
  been closed by an explicit header — see the "Open vs closed tables"
  rules above.

### Array-of-tables semantics

- `[[name]]` appends a new empty table to the array at `name`.
- The first `[[name]]` creates the array.
- Subsequent dotted references (e.g. `[name.child]` or
  `[[name.children]]`) refer to the **most recently appended** array
  element.
- A normal table `[name]` and an array of tables `[[name]]` at the same
  path **conflict** — pick one or the other.

### Inline-table semantics

- An inline table is **closed** at its `}` — no later expression can add a
  key to it, and no other inline-table appearance at the same path can
  merge into it.
- `name = { first = "Tom", last = "Preston-Werner" }` is exactly
  equivalent to `[name]` then `first = "Tom"` then `last = "Preston-Werner"`,
  except that the `[name]` form leaves `name` open to additional keys, and
  the inline form does not.

### Number semantics

- Decimal integers parse to a signed 64-bit value. `+0`, `-0`, `0` all map
  to the integer 0.
- Hex/oct/bin integers parse to non-negative values; the prefix denotes
  base. (A `+`/`-` sign is forbidden on these forms.)
- Floats parse to IEEE 754 binary64. `+0.0` and `-0.0` are distinct under
  IEEE 754 but compare equal.
- `inf`, `+inf` map to positive infinity; `-inf` maps to negative
  infinity. `nan`, `+nan`, `-nan` map to a NaN; the sign and payload of
  the NaN are implementation-defined.
- Underscores in numeric literals are stripped before interpretation —
  `1_000` is the integer 1000.

### Date-time semantics

- All date-time literals are local-aware: `OffsetDateTime` carries a UTC
  offset; the three "local" forms do not.
- Excess fractional-seconds precision is **truncated**, not rounded,
  before storage.
- Calendar validity (e.g. `2021-02-30`) is **enforced** at parse time
  beyond the ABNF — implementations must reject invalid calendar dates.

### Equivalence rules

- `1.0` is a `Float`; `1` is an `Integer`; they are **not equal** under
  TOML's data model (different types).
- `0xFF` and `255` represent the same `Integer` value 255.
- `0xff`, `0xFF`, `0xFf` are all the same value (hex digits are
  case-insensitive).
- Two keys are equal iff their decoded character sequences are equal,
  regardless of whether each is bare/basic-string/literal-string.
- Two strings are equal iff their decoded code-point sequences are
  equal — `"abc"` and `'abc'` and `"abc"` all denote the same value.

### Comments are non-semantic

Comments do not appear in the data model. A round-trip-preserving printer
**may** keep them as auxiliary metadata, but two TOMLs differing only in
comments represent the same data.

## Examples

### Minimal Valid File

The empty file is valid (zero expressions, root table empty). The smallest
non-empty document is a single key/value pair:

```toml
title = "TOML Example"
```

### Typical File

A realistic configuration that exercises the common features a real user
would write — string, integer, boolean, date-time, array, nested table,
comment.

```toml
# Application config
title = "Example Server"
version = 1

[server]
host = "0.0.0.0"
port = 8080
enabled = true
started_at = 2024-04-01T09:00:00Z

[database]
host = "db.example.com"
port = 5432
credentials = { username = "admin", password = "s3cret" }

[logging]
level = "info"
sinks = [ "stdout", "file:///var/log/app.log" ]
```

### Complex File

A document that exercises every token class and nearly every grammar
production: dotted keys, quoted keys, all four string forms (with
line-ending backslash), every integer base, special floats, all four
date-time forms, mixed-type arrays, multi-line arrays with comments,
implicitly-created tables, super-tables defined out-of-order, an inline
table, and arrays of tables with nested sub-tables.

```toml
# Comprehensive TOML v1.0.0 example.
title = "TOML ✨ example"           # unicode escape in basic string
literal_path = 'C:\Users\alice'         # literal string, no escapes
multiline_basic = """
The quick brown fox \
    jumps over the lazy dog.\
"""
multiline_literal = '''
First newline trimmed.
   Indentation preserved.
'''

# Integers in every base.
[numbers]
dec       = -1_000_000
zero      = 0
hex       = 0xDEAD_BEEF
oct       = 0o755
bin       = 0b1010_0101

# Floats including specials.
[numbers.floats]
fract     = 3.141_592
exp       = 6.022e23
both      = -1.5E-10
pos_inf   = +inf
neg_inf   = -inf
not_num   = nan

# Booleans.
[flags]
ready     = true
debug     = false

# All four date-time forms.
[times]
odt = 1979-05-27T07:32:00.999-07:00
ldt = 1979-05-27T07:32:00
ld  = 1979-05-27
lt  = 07:32:00.999

# Quoted and dotted keys; super-tables created implicitly.
[site]
"home page".url = "https://example.com"
"home page".visits = 12_345

# Array of tables, with a sub-table on each element.
[[fruits]]
name = "apple"

  [fruits.physical]
  color = "red"
  shape = "round"

  [[fruits.varieties]]
  name = "red delicious"

  [[fruits.varieties]]
  name = "granny smith"

[[fruits]]
name = "banana"

  [[fruits.varieties]]
  name = "plantain"

# Mixed-type array spanning multiple lines, with comments and a trailing comma.
[misc]
contributors = [
  "Foo Bar <foo@example.com>",                              # plain string
  { name = "Baz Qux", email = "baz@example.com" },          # inline table
  1979-05-27T07:32:00Z,                                     # offset date-time
  0xCAFE,                                                   # hex int
]
```

## Appendix

### Character Encoding

- TOML documents **must** be valid UTF-8.
- No BOM is specified by the standard. Most parsers, as a courtesy, accept
  and skip a leading UTF-8 BOM (`EF BB BF`); a strict implementation may
  reject it. Document the choice.
- Non-ASCII characters are permitted everywhere `non-ascii` appears in the
  ABNF (comments, strings, etc.). The `non-ascii` class is U+0080–U+D7FF
  and U+E000–U+10FFFF — surrogate code points are excluded.

### Newline Handling

- LF (`\n`, U+000A) and CRLF (`\r\n`, U+000D U+000A) are both newlines.
- Bare CR (U+000D not followed by LF) is **not** a newline anywhere in
  TOML and is invalid as a line terminator.
- Implementations **may** normalize newlines inside multi-line strings.

### Size Limits and Numeric Limits

- Integers must be at least the full signed 64-bit range
  (−2^63 … 2^63−1). Outside that range, parsers must error.
- Floats are IEEE 754 binary64.
- The spec sets no maximum on document size, key length, string length, or
  nesting depth. Implementers should pick conservative limits documented
  in their implementation notes.

### Version History (TOML 1.0.0 vs earlier)

This document covers **TOML 1.0.0** (2021-01-11). Earlier 0.x releases
differed in date-time syntax (no local-only forms), in arrays
(homogeneous-only), and in dotted-key behavior. The 1.0.0 grammar is the
first stable release, frozen for backwards-compatible 1.x evolution.

### Related Standards and References

- **RFC 5234** — Augmented BNF (ABNF), the grammar notation used in
  `toml.abnf`.
- **RFC 3339** — Date and Time on the Internet, the basis for TOML's four
  date-time forms.
- **IEEE 754** — binary64 (`double`), the floating-point format for TOML's
  `Float`.
- **Unicode 13.0+ / UTF-8** — encoding requirement.
- Canonical sources:
  - Prose: <https://toml.io/en/v1.0.0>
  - ABNF: <https://github.com/toml-lang/toml/blob/1.0.0/toml.abnf>
  - GitHub: <https://github.com/toml-lang/toml>
