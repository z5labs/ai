# JSON Specification Reference

## Overview

JSON (JavaScript Object Notation) is a lightweight, text-based, language-independent data interchange format derived from JavaScript object literal syntax. This reference covers the version standardized by **RFC 8259** (December 2017), which obsoletes RFC 7159 and is identical in syntax to ECMA-404. JSON describes a small set of structured values — objects, arrays, numbers, strings, and the three literal names `true`, `false`, and `null` — built from six structural characters and a defined whitespace class. The grammar is given in ABNF (RFC 5234). JSON text exchanged between systems that are not part of a closed ecosystem MUST be encoded in UTF-8.

This reference is intended as input to a Go tokenizer / parser / printer implementation. It deliberately omits topics outside the syntactic scope: the I-JSON interoperability profile (RFC 7493) and security considerations (RFC 8259 §12).

## Lexical Elements (Tokens)

A JSON text is a sequence of tokens. RFC 8259 §2 defines the token set as six structural characters, three literal names, plus `string` and `number` tokens. The grammar wraps each structural character in an optional `ws` (whitespace) class on either side; for tokenization purposes, whitespace is a separate non-emitting token class and the structural characters carry no whitespace of their own.

The implementer should emit the following token classes from the tokenizer.

| Name | Pattern (ABNF / regex) | Examples | Edge cases |
|---|---|---|---|
| `BeginArray` | `%x5B` (`[`) | `[` | A single byte; never combined with whitespace at the token level. |
| `EndArray` | `%x5D` (`]`) | `]` | Same. |
| `BeginObject` | `%x7B` (`{`) | `{` | Same. |
| `EndObject` | `%x7D` (`}`) | `}` | Same. |
| `NameSeparator` | `%x3A` (`:`) | `:` | Used only inside objects (`member = string name-separator value`). |
| `ValueSeparator` | `%x2C` (`,`) | `,` | Used inside arrays and objects; trailing commas are NOT allowed. |
| `True` | `%x74.72.75.65` (`true`) | `true` | Case-sensitive; `True` / `TRUE` are not literals. |
| `False` | `%x66.61.6c.73.65` (`false`) | `false` | Case-sensitive. |
| `Null` | `%x6e.75.6c.6c` (`null`) | `null` | Case-sensitive. |
| `StringLiteral` | `quotation-mark *char quotation-mark` | `""`, `"hello"`, `"\"\\\/\b\f\n\r\t"`, `"é"`, `"𝄞"` | Empty string is valid. Unescaped `"`, `\`, and `U+0000`–`U+001F` MUST be escaped. Non-BMP characters are escaped as a UTF-16 surrogate pair (12 chars). |
| `NumberLiteral` | `[ minus ] int [ frac ] [ exp ]` | `0`, `-0`, `42`, `-3.14`, `2.5e10`, `1E-3`, `0.5` | Leading zeros forbidden (`01` invalid). `+1` invalid (no leading plus). `.5`, `5.`, `Infinity`, `NaN`, `0x1F` all invalid. |
| `Whitespace` | `*( %x20 / %x09 / %x0A / %x0D )` | (single space, tab, LF, CR) | Allowed before or after any structural character; not part of any value token's bytes. |

> **Ambiguity:** RFC 8259 §8.2 acknowledges that the ABNF for `string` permits surrogate-only sequences (e.g. `"\uDEAD"`) that do not encode a valid Unicode scalar. The spec calls behaviour on such inputs "unpredictable" but does not require rejection. The implementer should decide per-call whether to (a) accept the bit sequence, (b) replace with U+FFFD, or (c) error.

### Comments

JSON does not define comments. Any `//`, `/* */`, or `#` sequence outside a string literal is a syntax error. State this to users explicitly: silence on comments has historically led to ad-hoc extensions (JSON5, JSONC) that are not RFC 8259.

### Whitespace and Delimiters

The `ws` production permits four characters and only four:

```abnf
ws = *(
        %x20 /              ; Space
        %x09 /              ; Horizontal tab
        %x0A /              ; Line feed or New line
        %x0D )              ; Carriage return
```

- `U+0020` (SPACE), `U+0009` (TAB), `U+000A` (LF), `U+000D` (CR) are the only whitespace characters. Vertical tab (`U+000B`), form feed (`U+000C`), `U+00A0` (NBSP), and other Unicode whitespace are NOT whitespace in JSON.
- Whitespace is allowed before or after any of the six structural characters (`[ ] { } : ,`) and around the top-level value (`JSON-text = ws value ws`).
- Whitespace is NOT allowed inside a number, between the digits of a number, or inside a literal name. `tr ue`, `1 .0`, and `1 e2` are all syntax errors.
- Whitespace is significant inside a string literal — every character between the opening and closing `"` is part of the string's content.
- There is no line-continuation construct. Unescaped LF or CR inside a string literal is forbidden by the `unescaped` production (which excludes `%x00-1F`).
- Trailing newline at end of file is neither required nor forbidden — it falls under `ws` after the top-level value.

### Literals

**String literals.** A string begins and ends with `"` (`U+0022`). Allowed contents are unescaped characters and escape sequences. Multi-line strings are not supported (raw `LF`/`CR` are control characters and forbidden unescaped).

```abnf
string         = quotation-mark *char quotation-mark
char           = unescaped / escape ( %x22 / %x5C / %x2F / %x62 / %x66 / %x6E / %x72 / %x74 / %x75 4HEXDIG )
escape         = %x5C                ; \
quotation-mark = %x22                ; "
unescaped      = %x20-21 / %x23-5B / %x5D-10FFFF
```

Two-character escape sequences:

| Escape | Code point | Meaning |
|---|---|---|
| `\"` | `U+0022` | quotation mark |
| `\\` | `U+005C` | reverse solidus |
| `\/` | `U+002F` | solidus (forward slash) — escaping is optional |
| `\b` | `U+0008` | backspace |
| `\f` | `U+000C` | form feed |
| `\n` | `U+000A` | line feed |
| `\r` | `U+000D` | carriage return |
| `\t` | `U+0009` | horizontal tab |
| `\uXXXX` | `U+XXXX` | four hex digits, case-insensitive (`A–F` or `a–f`) |

Characters outside the Basic Multilingual Plane (above `U+FFFF`) MUST be encoded as a UTF-16 surrogate pair: `\uD800-\uDBFF` (high) followed by `\uDC00-\uDFFF` (low). Example: `U+1D11E` (G clef) is `"𝄞"`.

**Number literals.** Decimal, base-10. Integer part is required; fraction and exponent are optional.

```abnf
number        = [ minus ] int [ frac ] [ exp ]
decimal-point = %x2E                ; .
digit1-9      = %x31-39
e             = %x65 / %x45         ; e or E
exp           = e [ minus / plus ] 1*DIGIT
frac          = decimal-point 1*DIGIT
int           = zero / ( digit1-9 *DIGIT )
minus         = %x2D                ; -
plus          = %x2B                ; +
zero          = %x30                ; 0
```

Constraints:
- No leading `+` on the number itself (only inside `exp`).
- No leading zeros on `int` other than the literal `0` (`01` is invalid; `0` alone is valid; `0.5` is valid; `-0` is valid).
- A decimal point requires at least one digit on each side (`5.` and `.5` are both invalid).
- An `exp` requires at least one digit after `e`/`E` and the optional sign.
- `Infinity`, `-Infinity`, `NaN`, hex (`0x...`), octal, and binary literals are NOT permitted.
- Range and precision are implementation-defined; RFC 8259 §6 notes that integers in `[-(2^53)+1, (2^53)-1]` are interoperable when receivers use IEEE 754 binary64.

**Boolean and null literals.** `true`, `false`, and `null` are the only literal names. They MUST be lowercase; `True`, `TRUE`, `Null`, etc. are syntax errors. They are full-word tokens (no surrounding quotes).

### Keywords and Reserved Words

JSON has exactly three reserved literal names — `true`, `false`, `null` — that act as values, not as identifiers. JSON has no user-defined identifiers, so there is no broader keyword-vs-name conflict. Member names are always strings, never bare words.

### Symbols and Operators

JSON has no operators (no arithmetic, comparison, or logical operators inside the grammar). The structural symbols are:

| Symbol | Code point | Role |
|---|---|---|
| `[` | `U+005B` | begin array |
| `]` | `U+005D` | end array |
| `{` | `U+007B` | begin object |
| `}` | `U+007D` | end object |
| `:` | `U+003A` | name separator (between member name and value) |
| `,` | `U+002C` | value separator (between array elements; between members) |
| `"` | `U+0022` | string quotation mark (token-internal — not emitted alone) |
| `\` | `U+005C` | escape introducer (token-internal — not emitted alone) |

## Structure (Grammar)

### Top-Level Structure

A JSON text is exactly one value, optionally surrounded by whitespace.

```abnf
JSON-text = ws value ws
```

Constraints:
- Exactly one root value. Multiple top-level values (`1 2`) or trailing garbage after a complete value are syntax errors.
- The root value MAY be of any type: object, array, number, string, `true`, `false`, or `null`. (RFC 4627 required object or array; RFC 8259 relaxes this.)
- Implementations MUST NOT emit a UTF-8 BOM (`U+FEFF`); parsers MAY ignore one if present.
- No declared file extension or magic number is required at the syntax level (see Appendix).

### Grammar Productions

All productions below are taken verbatim from RFC 8259 §§2–7 in ABNF (RFC 5234). Production names are kept as-is so the implementer can map them onto AST node types.

**JSON-text** — the document root.
```abnf
JSON-text = ws value ws
```
- Members: optional leading `ws`, exactly one `value`, optional trailing `ws`.
- Constraints: exactly one root; nothing after it.

**value** — a discriminated union of seven variants.
```abnf
value = false / null / true / object / array / number / string
```
- AST: typically a `Value` interface implemented by `Object`, `Array`, `Number`, `String`, and three nullary singletons (`True`, `False`, `Null`).

**object** — zero or more members between `{` and `}`.
```abnf
object = begin-object [ member *( value-separator member ) ] end-object
member = string name-separator value
```
- Members: a (possibly empty) list of `member`s, comma-separated.
- A `member` is a (string name, value) pair joined by `:`.
- Multiplicity: `*` members allowed; the empty object `{}` is valid.
- No trailing comma after the last member.
- The names within an object SHOULD be unique. See Semantics for behaviour on duplicates.

**array** — zero or more values between `[` and `]`.
```abnf
array = begin-array [ value *( value-separator value ) ] end-array
```
- Members: a (possibly empty) list of `value`s, comma-separated.
- Multiplicity: `*` values allowed; the empty array `[]` is valid.
- No trailing comma after the last element.
- Element types may be heterogeneous: an array MAY mix numbers, strings, objects, etc.

**number** — see Lexical Elements > Literals for the full ABNF. The parser receives a single `NumberLiteral` token.

**string** — see Lexical Elements > Literals. The parser receives a single `StringLiteral` token whose value is the decoded sequence of Unicode code points (escapes resolved).

**Structural character productions** — wrap whitespace around each structural byte.
```abnf
begin-array     = ws %x5B ws
end-array       = ws %x5D ws
begin-object    = ws %x7B ws
end-object      = ws %x7D ws
name-separator  = ws %x3A ws
value-separator = ws %x2C ws
```
- For tokenizers that emit whitespace as a separate (skipped) token, the parser sees only the bare structural token; the `ws` allowance is satisfied by the tokenizer skipping over whitespace between tokens.

### Ordering and Optionality

- **Object members are unordered.** JSON parsing libraries differ on whether they preserve member order; RFC 8259 §4 explicitly says implementations whose behaviour does not depend on member ordering are interoperable. Printers that need deterministic output MAY sort members lexicographically by name (the implementer should expose this as a printer option).
- **Array elements are ordered.** Their textual order is the array's logical order; printers MUST preserve it.
- **Required vs. optional:** every grammar element above is either fully required (e.g. the matching `}` for an `object`) or fully absent (no member or element appears zero times by being implicit — emptiness is expressed by the optional `[ ... ]` group around the comma-separated list). There is no concept of "default values" at the syntax level.
- **Whitespace is always optional** wherever the grammar permits `ws`. Compact (`{"k":1}`) and pretty-printed JSON parse identically.

## Semantics

JSON is a syntactic format; most semantics belong to the consuming application. The following rules nevertheless affect the AST shape, the printer's output, and equality checks, and the implementer must decide them up front.

- **Number representation.** RFC 8259 §6 does not require a particular numeric type. Implementations that target IEEE 754 binary64 are interoperable for integers in `[-(2^53)+1, (2^53)-1]` and for finite doubles. The Go implementer should decide per-call between (a) decoding to `float64` (lossy for large integers), (b) decoding to a string-backed `Number` type (lossless), or (c) decoding integers to `int64` and fractions to `float64`.
- **Number equivalence.** `1`, `1.0`, `1e0`, and `10e-1` all represent the same mathematical value but have different textual forms. Whether a printer round-trips the original textual form or normalises it is implementation-defined. The grammar treats them as distinct tokens; a value-level comparison must compare numerically.
- **Negative zero.** `-0` is grammatically valid and distinct from `0` in IEEE 754; equality semantics depend on the chosen number type.
- **Duplicate member names.** RFC 8259 §4: "The names within an object SHOULD be unique." Behaviour on duplicates is unspecified and observed implementations vary: last-wins, first-wins, error, or all-pairs preserved. The implementer should pick one and document it. Last-wins is the most common.
- **String comparison.** RFC 8259 §8.3: equality should be tested on the decoded Unicode code-unit sequences, not the source bytes. `"a\\b"` and `"a\b"` are equal strings even though their source forms differ.
- **String content.** A `StringLiteral` token's value is the decoded sequence of Unicode code points after escapes are resolved and surrogate pairs are combined.
- **Cross-references.** JSON has no `$ref`, anchor, or include mechanism — those are extensions (JSON Schema, JSON Pointer, JSON-LD) and are out of scope for the core format.
- **Validation beyond syntax.** None at the format level. Any `value` produced by the grammar is a valid JSON value.

## Examples

The three examples below all conform to the grammar above and are intended as round-trip fixtures (tokenizer → parser → AST → printer → tokenizer …).

### Minimal Valid File

A single literal value as the entire document. Note the absence of any structure — RFC 8259 explicitly permits a bare value as a JSON text.

```json
null
```

Equally minimal alternates: `true`, `false`, `0`, `""`, `[]`, `{}`.

### Typical File

A flat object with the four most common scalar types — strings, numbers, booleans, and null — plus one nested structure. Whitespace is light but present.

```json
{
  "name": "ada",
  "age": 36,
  "active": true,
  "manager": null,
  "tags": ["engineer", "founder"]
}
```

### Complex File

A nested document that exercises every token class and most grammar productions: nested objects and arrays, the empty object and empty array, a heterogeneous array, all three literal names, every numeric form (integer, negative, fraction, exponent, negative zero), and a string with each kind of escape including a surrogate pair.

```json
{
  "empty_object": {},
  "empty_array": [],
  "nested": {
    "level": 2,
    "children": [
      {"id": 1, "value": -0},
      {"id": 2, "value": 3.14},
      {"id": 3, "value": 6.022e23},
      {"id": 4, "value": 1E-9}
    ]
  },
  "literals": [true, false, null],
  "mixed": [1, "two", 3.0, [4, [5]], {"six": 6}],
  "escapes": "quote=\" backslash=\\ slash=\/ bs=\b ff=\f lf=\n cr=\r tab=\t bmp=é astral=𝄞",
  "duplicate_keys_note": {"k": 1, "k": 2}
}
```

> **Ambiguity:** The final member `"duplicate_keys_note"` contains a duplicate key. RFC 8259 SHOULDs uniqueness but does not forbid duplicates; this fixture exists specifically to exercise the implementer's chosen policy (typically last-wins).

## Appendix

- **Character encoding.** UTF-8 is required for JSON exchanged outside a closed ecosystem (RFC 8259 §8.1). Implementations MUST NOT emit a BOM; parsers MAY ignore a leading BOM.
- **Media type.** `application/json` (RFC 8259 §11). File extension `.json`. No `charset` parameter is defined; encoding is fixed to UTF-8 by §8.1.
- **Size limits.** None at the spec level. Parsers MAY impose limits on text size, nesting depth, number range / precision, and string length (RFC 8259 §9).
- **Surrogate handling.** ABNF permits unpaired surrogates inside strings; behaviour on such inputs is unpredictable. Implementations targeting strict Unicode should reject or replace them.
- **Version history.** This reference covers RFC 8259 (December 2017), which obsoletes RFC 7159 (March 2014) and RFC 4627 (July 2006). RFC 8259 is identical in syntax to ECMA-404 (2nd edition, December 2017).
- **Out of scope for this reference.** The I-JSON interoperability profile (RFC 7493) and security considerations (RFC 8259 §12) were excluded by user request.
- **Related standards.** RFC 5234 (ABNF), RFC 3629 (UTF-8), Unicode 10.0+, IEEE 754-2008 binary64, ECMA-404.
