# KVR Text File Format

## Overview

KVR (Key-Value Records) is a small text format for declaring typed records, optionally grouped into named blocks, with line-comment support. A KVR file is a sequence of top-level statements separated by newlines. Two statement kinds exist: a **record** (a single typed key-value declaration) and a **block** (a named group of records).

Whitespace between tokens is insignificant except inside quoted strings. End-of-line is just whitespace — statements end implicitly at the next valid statement opener or at end-of-file.

```
# a leading comment attached to the first record
record string GREETING = "hello"

block COLORS {
    record string RED  = "ff0000";
    record string BLUE = "0000ff";
}
```

## Lexical Elements (Tokens)

KVR has six token types. Every token carries a `Pos{Line, Column}` (1-based) marking its first rune in the source.

### Identifier

A sequence of one or more ASCII letters, digits, or underscores, starting with a letter or underscore. Yielded as `TokenIdentifier` with `Value` set to the identifier text.

```
foo  GREETING  user_id  _temp123
```

The reserved words `record` and `block` are still yielded as `TokenIdentifier`; the parser distinguishes them by value.

### Symbol

A single punctuation character. The KVR format uses these symbols:

| Symbol | Token value | Meaning                       |
|--------|-------------|-------------------------------|
| `=`    | `=`         | record value separator        |
| `{`    | `{`         | block open                    |
| `}`    | `}`         | block close                   |
| `;`    | `;`         | block-internal record separator |

Yielded as `TokenSymbol` with `Value` set to the symbol character.

### String

A run of characters enclosed in double quotes (`"`). Backslash-escapes are recognised for `\\`, `\"`, `\n`, `\t`. The yielded `Value` is the **decoded** string content (the surrounding quotes and escape sequences are not part of `Value`).

```
"hello"      → Value = `hello`
"a\"b"       → Value = `a"b`
"line1\n"    → Value = `line1` followed by newline
```

> **Ambiguity:** A literal newline (a real `\n` rune, not the two-character escape) inside a quoted string is rejected with `UnterminatedStringError` carrying the opening-quote position.

### Number

A run of one or more ASCII digits. Yielded as `TokenNumber` with `Value` set to the digit text (no parsing into a numeric type — the parser does that).

```
0  42  1234567
```

KVR does not currently support negative numbers, decimals, or exponents.

### Comment

A line comment begins with `#` and runs to the next newline (or end-of-file). The `#` and any leading horizontal whitespace are not part of `Value`; the trailing newline is also not part of `Value`. Yielded as `TokenComment`.

```
# hello world      → Value = `hello world`
#tight             → Value = `tight`
```

### Invalid

The starter `TokenInvalid` exists as the zero value for `TokenType`. Tokens are never emitted with this type — it's a sentinel for "uninitialised" so a stray `Token{}` in a test fails loudly.

## Structure (Grammar)

The grammar is intentionally small. Productions are written below in EBNF; terminals are the token types from "Lexical Elements".

```
File        = { Statement } .
Statement   = Comment* ( Record | Block ) .
Comment     = TokenComment .
Record      = "record" Type Identifier "=" Value .
Block       = "block" Identifier "{" { Statement ";" } "}" .
Type        = "string" | "number" .
Value       = TokenString | TokenNumber .
```

### Records

A record is the literal identifier `record`, followed by a type name (`string` or `number`), followed by the record's key (an identifier), followed by `=`, followed by a value of the matching type. There is no statement terminator at the top level — the next valid token (an identifier `record`, an identifier `block`, a comment, or end-of-file) ends the record.

### Blocks

A block is the literal identifier `block`, followed by an identifier (the block name), followed by `{`, then zero or more inner statements separated by `;`, then `}`. The `;` is required between statements inside a block and is also required after the last inner statement (i.e. before the closing `}`).

```
block X { record string A = "1"; record string B = "2"; }   # legal
block X { record string A = "1"; record string B = "2" }    # illegal — missing trailing ;
```

### Comments are statements with attachment

A `TokenComment` is a valid statement on its own (a "free-floating" comment), but the printer's responsibility is to attach a run of comments to the **immediately following** non-comment statement so they survive a round-trip. Implementations should expose a `LeadingComments []string` field on `Record` and `Block` that the parser populates and the printer emits before the node's own output.

## Semantics

- **Identifier case sensitivity**: identifiers are case-sensitive. `RED` and `red` are distinct keys.
- **Record key uniqueness**: keys are not required to be unique. A KVR file with two records sharing a key is well-formed; the consumer of the AST decides which wins.
- **Block uniqueness**: the same applies to block names.
- **Whitespace fidelity**: blank lines, trailing whitespace, and the exact column of a token are not preserved across a round-trip. Only comments and structural content round-trip.
- **Type/value agreement**: `record string K = 42` (number value with `string` type) is rejected at parse time with a typed `TypeMismatchError{Type, Got}` carrying the value-token position.

## Examples

### Minimal: empty file

An empty input is a legal KVR file. `Parse("")` yields `&File{}`; `Print(&File{})` writes nothing.

### Typical: two records, one comment

```
# the universal greeting
record string GREETING = "hello"
record number ANSWER  = 42
```

The AST should contain two `Record` values. The first has `LeadingComments: []string{"the universal greeting"}`.

### Complex: a block with two records and a comment inside

```
block COLORS {
    # primary red
    record string RED = "ff0000";
    record string BLUE = "0000ff";
}
```

The block's `Records[0]` carries the comment; `Records[1]` does not.

### Round-trip: comments on top-level records

```
# greet
record string A = "1"

# answer
record number B = 2
```

After `Parse → Print → Parse`, both records still carry their leading comment.
