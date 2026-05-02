# KVRX Text File Format

## Overview

KVRX (Key-Value Records Extended) is a small text format for declaring typed records, optionally grouped into named blocks, with rich primitives, composite values, conditional sections, type aliases, and imports. A KVRX file is a sequence of top-level statements separated by newlines or semicolons.

A **statement** is one of: a `record` (a typed key-value declaration), a `block` (a named group of statements), an `import` (pull statements from another file), a `type` alias (give a name to a value shape), or an `if` conditional (parse-time-evaluated branching). All five share the same comment-attachment rule: any leading run of comments attaches to the immediately following non-comment statement so they survive a round-trip.

Whitespace between tokens is insignificant except inside quoted strings. Statements at the top level end at the next valid statement opener (one of the five keywords above) or at end-of-file. Statements inside a block end at `;` or at the closing `}`.

```
# the universal greeting
record string GREETING = "hello"

block COLORS {
    record string RED  = "ff0000";
    record string BLUE = "0000ff";
    record bool   DARK = true;
}

import "shared/units.kvrx"

type Port = number

if (FEATURE_FLAG) {
    record Port HTTP = 8080;
}
```

KVRX is designed to round-trip cleanly: leading comments survive parse-then-print, and structural content (records, blocks, imports, type aliases, conditional bodies) is preserved verbatim. The format is consciously line-oriented for human authoring; tools that synthesise KVRX programmatically should still emit one statement per line so the resulting file is reviewable.

## Lexical Elements (Tokens)

KVRX has eight token types. Every token carries a `Pos{Line, Column}` (1-based) marking its first rune in the source.

### Identifier

A sequence of one or more ASCII letters, digits, or underscores, starting with a letter or underscore. Yielded as `TokenIdentifier` with `Value` set to the identifier text.

```
foo  GREETING  user_id  _temp123
```

The reserved words `record`, `block`, `import`, `type`, `if`, `elif`, `else`, `true`, `false`, and `null` are still yielded as `TokenIdentifier`; the parser distinguishes them by value. This keeps the tokenizer deterministic and small — no lookup table on the hot path.

### Symbol

A single punctuation character (or, for `==` / `!=` / `<=` / `>=` / `&&` / `||`, two characters). The KVRX format uses these symbols:

| Symbol | Meaning                            |
|--------|------------------------------------|
| `=`    | record value separator             |
| `{`    | block / map open                   |
| `}`    | block / map close                  |
| `[`    | list open                          |
| `]`    | list close                         |
| `(`    | expression / condition open        |
| `)`    | expression / condition close       |
| `,`    | list / map / argument separator    |
| `;`    | statement separator inside a block |
| `:`    | type annotation                    |
| `&`    | reference prefix                   |
| `==`   | equality                           |
| `!=`   | inequality                         |
| `<`    | less-than                          |
| `<=`   | less-than-or-equal                 |
| `>`    | greater-than                       |
| `>=`   | greater-than-or-equal              |
| `&&`   | logical AND                        |
| `\|\|` | logical OR                         |
| `!`    | logical NOT                        |
| `+`    | numeric add / string concat        |
| `-`    | numeric subtract / unary minus     |
| `*`    | numeric multiply                   |
| `/`    | numeric divide                     |

Yielded as `TokenSymbol` with `Value` set to the symbol text (one or two characters). Two-character symbols are recognized greedily: a `=` followed by `=` becomes `==`, never two `=` tokens. The single-character `&` is distinguishable from `&&` because the second `&` is the *next* rune; a `&` followed by anything else yields `TokenSymbol` with `Value` = `&`.

### String

A run of characters enclosed in double quotes (`"`). Backslash-escapes are recognised for `\\`, `\"`, `\n`, `\t`, `\r`, `\0`, and `\xHH` (two hex digits). The yielded `Value` is the **decoded** string content.

```
"hello"        → Value = `hello`
"a\"b"         → Value = `a"b`
"line1\n"      → Value = `line1` followed by newline
"ctrl-A: \x01" → Value = `ctrl-A: ` followed by byte 0x01
```

> **Ambiguity:** A literal newline (a real `\n` rune) inside a single-line string is rejected with `UnterminatedStringError` carrying the opening-quote position. Use a triple-quoted string for multi-line content.

### Triple-quoted string

A run of characters enclosed in `"""`. Newlines, double quotes, and backslashes are taken literally (no escape processing) so file paths, JSON snippets, and source code can be embedded without escaping. Yielded as `TokenString` with `Value` set to the verbatim content between the opening and closing `"""`.

```
"""
{
    "k": "v"
}
"""           → Value = `\n{\n    "k": "v"\n}\n`
```

The opening `"""` and the closing `"""` are not part of `Value`. A leading newline immediately after the opening `"""` is preserved (so the example above's `Value` begins with `\n`). The tokenizer detects `"""` greedily — three consecutive `"` runes always open a triple-quoted string, never an empty single-quoted string followed by an open quote.

### Number

A run that produces an integer or a floating-point literal. KVRX recognises four integer forms and one float form. The yielded `Value` is the **literal source text**, not a parsed numeric value — the parser converts it into an integer or float when typing the record.

| Form    | Pattern                            | Example   |
|---------|------------------------------------|-----------|
| decimal | `[0-9]+`                           | `42`      |
| hex     | `0x[0-9a-fA-F]+`                   | `0xFF`    |
| octal   | `0o[0-7]+`                         | `0o755`   |
| binary  | `0b[01]+`                          | `0b1010`  |
| float   | `[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?` | `1.5e-3`  |

Yielded as `TokenNumber` with `Value` set to the literal text. The form is recoverable from the prefix (or absence of one).

> **Ambiguity:** A bare leading zero (`042`) is **not** valid. Decimal numbers may not have leading zeros except for the sole literal `0`. Use `0o` for octal, `0x` for hex, or `0b` for binary explicitly.

### Comment

KVRX has both line comments and block comments. A line comment begins with `#` and runs to the next newline (or end-of-file); a block comment is `/* ... */` and may span multiple lines but does **not** nest. The `#`, `/*`, `*/`, and any leading horizontal whitespace are not part of `Value`; the trailing newline (line comments) is also not part of `Value`. Yielded as `TokenComment`.

```
# hello world          → Value = `hello world`
/* multi
   line  */            → Value = `multi\n   line  `
```

### Newline

A `\n` rune. Yielded as `TokenNewline`. Newlines act as soft statement terminators at the top level (after a record or import, the next token starting on a new line ends the prior statement). Inside a block, statements are explicitly terminated with `;` and newlines are insignificant.

### Invalid

The starter `TokenInvalid` exists as the zero value for `TokenType`. Tokens are never emitted with this type — it's a sentinel for "uninitialised" so a stray `Token{}` in a test fails loudly.

## Structure (Grammar)

The grammar is intentionally explicit. Productions are written below in EBNF; terminals are the token types from "Lexical Elements" (or the literal identifier value for the reserved words).

### Grammar conventions

Productions use `=` to introduce a definition and `.` to terminate it. Alternatives are separated by `|`. Optional elements appear in `[ ... ]`; zero-or-more repetitions in `{ ... }`. Literal terminals are written in double quotes (`"record"`, `"="`); token-type terminals are written without quotes (`Identifier`, `TokenString`).

A few grammar choices deserve explicit calling out, because they affect every downstream consumer:

- **Reserved words are terminals defined by literal identifier value.** `"record"` in a production means "a `TokenIdentifier` whose `Value` is exactly the four-letter string `record`". The tokenizer does not classify keywords; the parser does, by matching on `Value`. This keeps the tokenizer state machine small and decouples it from grammar evolution — adding a new keyword to KVRX is a parser change only.

- **Newlines are not in the grammar.** The `TokenNewline` token is consumed silently by the parser when it appears between top-level statements (it acts as a soft terminator). Inside a block, newlines are insignificant — only `;` terminates a statement. The grammar productions therefore don't mention newlines explicitly; the parser's tokenisation layer absorbs them.

- **Comments are not in the grammar's main productions.** The `Comment*` prefix on `Statement` is the only place comments appear; everywhere else, the parser drops `TokenComment` tokens silently. This is so the grammar reads as if comments don't exist (which is the typical authoring view), while the parser still captures them for round-trip via leading-comment attachment.

### Top-level grammar

```
File         = { Statement } .
Statement    = Comment* ( Record | Block | Import | TypeAlias | Conditional ) .
Comment      = TokenComment .
```

The `File` production accepts zero or more statements. The empty file is legal — `Parse("")` yields `&File{}`. A file consisting only of comments is also legal; the comments attach to no statement and are stored as `TrailingComments` on the `File` itself.

### Records

```
Record       = "record" Type Identifier "=" Expression .
```

A record is the literal identifier `record`, followed by a type expression (`Type`), followed by the record's key (an identifier), followed by `=`, followed by an expression. There is no statement terminator at the top level — the next valid token (the keyword of any other statement form, a comment, or end-of-file) ends the record.

The `Type` immediately after `record` may be a single identifier (`record string K = ...`), a list type (`record [number] PORTS = [...]`), or a map type (`record {string:number} LIMITS = {...}`). Type aliases (`type Port = number`) introduce names that can stand in for any of the three forms.

```
record string  GREETING = "hello"
record number  PORT     = 8080
record bool    ENABLED  = true
record [string] HOSTS   = ["a.example", "b.example"]
record {string: number} LIMITS = { "rps" = 100, "burst" = 250 }
```

The expression on the right of `=` is type-checked against `Type` per the rules in `## Semantics`. A `record string K = 42` is a `TypeMismatchError`; a `record number K = "x"` likewise.

### Blocks

```
Block        = "block" Identifier [ ":" Type ] "{" { Statement ";" } "}" .
```

A block is the literal identifier `block`, followed by an identifier (the block name), an optional `:` followed by a type (the block's "shape"), followed by `{`, then zero or more inner statements separated by `;`, then `}`. The `;` is required between statements inside a block and is also required after the last inner statement (i.e. before the closing `}`).

```
block X { record string A = "1"; record string B = "2"; }   # legal
block X { record string A = "1"; record string B = "2" }    # illegal — missing trailing ;
```

A typed block (`block X : Section { ... }`) constrains the inner statements: every record's key must appear in the block's `Section` type alias (see `## Semantics`).

```
block PROD : Section {
    record string HOST = "prod.example.com";
    record number PORT = 443;
    record bool   DARK = true;
}
```

A block whose body contains another block is legal — blocks nest arbitrarily deeply. The implementation does not impose a maximum nesting depth; consumers concerned about pathological inputs should bound recursion on their own AST traversal.

### Imports

```
Import       = "import" StringLiteral .
```

An import statement is the literal identifier `import` followed by a single string literal. The string is the path to another KVRX file relative to the importing file. Importing pulls every top-level statement from the imported file into the importer's scope; circular imports are detected and rejected at parse time.

```
import "shared/units.kvrx"
import "lib/colors/primary.kvrx"
```

The string must be a single-line string (`"..."`), not a triple-quoted string — paths don't contain newlines, and forbidding the triple-quoted form keeps the diagnostic concrete (`UnexpectedTokenError{Want: TokenString}` is the only error variant the parser must emit for a malformed import).

### Type aliases

```
TypeAlias    = "type" Identifier "=" Type .
```

A type alias is the literal identifier `type`, followed by an identifier (the alias name), followed by `=`, followed by a type expression. The alias introduces a name that resolves to its right-hand type wherever a type is expected.

```
type Port  = number
type Hosts = [string]
type Tags  = {string: string}
```

Aliases may reference other aliases declared earlier in the file, but not later ones — forward references are rejected so the parser does not need a multi-pass type resolver. The forbidden pattern is:

```
type A = B   # illegal — B is not yet declared
type B = number
```

### Conditionals

```
Conditional  = "if" "(" Expression ")" "{" { Statement ";" } "}"
               { "elif" "(" Expression ")" "{" { Statement ";" } "}" }
               [ "else" "{" { Statement ";" } "}" ] .
```

A conditional begins with `if (` followed by an expression, followed by `) {`, followed by zero or more inner statements separated by `;`, followed by `}`. Zero or more `elif (...) { ... }` clauses may follow; an optional `else { ... }` clause may close the chain.

```
if (&MODE == "prod") {
    record number PORT = 443;
} elif (&MODE == "stage") {
    record number PORT = 8443;
} else {
    record number PORT = 8080;
}
```

The expression is evaluated at parse time against the values of previously-declared records (see `## Semantics`). The body of the first branch whose expression evaluates to `true` is treated as if its statements appeared at the conditional's position; non-matching branches are discarded from the active program but **preserved in the AST** so they round-trip through the printer.

### Types

```
Type         = NamedType | ListType | MapType .
NamedType    = Identifier .
ListType     = "[" Type "]" .
MapType      = "{" Type ":" Type "}" .
```

A named type is a single identifier. The built-in named types are `string`, `number`, `bool`, and `null`. Any other identifier is treated as a type alias and must resolve at parse time (see `## Semantics`). A list type wraps an inner type in `[ ... ]`; a map type uses `{K: V}` (note the colon, distinguishing it from a value-side map literal which uses `=`).

```
string                 # named type
[string]               # list-of-string
{string: number}       # map-of-string-to-number
[[number]]             # list-of-list-of-number
{string: [number]}     # map-of-string-to-list-of-number
```

Map keys are restricted to types whose values can be compared for equality: `string`, `number`, `bool`. A map type whose key type is a list, map, or `null` is rejected at parse time with `InvalidMapKeyError{Pos, KeyType}`.

### Expressions

```
Expression   = OrExpr .
OrExpr       = AndExpr { "||" AndExpr } .
AndExpr      = NotExpr { "&&" NotExpr } .
NotExpr      = [ "!" ] CompareExpr .
CompareExpr  = AddExpr { ( "==" | "!=" | "<" | "<=" | ">" | ">=" ) AddExpr } .
AddExpr      = MulExpr { ( "+" | "-" ) MulExpr } .
MulExpr      = UnaryExpr { ( "*" | "/" ) UnaryExpr } .
UnaryExpr    = [ "-" ] PrimaryExpr .
PrimaryExpr  = Reference
             | Literal
             | "(" Expression ")"
             | ListValue
             | MapValue .
```

The expression productions encode operator precedence directly in the grammar. The lowest-precedence binary operator (`||`) is the outermost production; the highest-precedence operators (unary `-`, references, literals, parens) are innermost. A consumer building an AST should reflect this nesting — an expression `a + b * c` parses as `Add(a, Mul(b, c))`, never as `Mul(Add(a, b), c)`.

### Comparison chains are illegal

Comparison operators are non-associative — `a < b < c` is rejected with a `ChainedComparisonError`:

```
record bool VALID = 1 < 2 < 3   # illegal
```

The implementation detects this in the `CompareExpr` production: after consuming one comparison operator, encountering another comparison operator before a higher-precedence operator (or end of expression) is a chained-comparison error.

### Primary expressions

```
Reference    = "&" Identifier .
Literal      = StringLiteral | NumberLiteral | BoolLiteral | NullLiteral .
StringLiteral = TokenString .
NumberLiteral = TokenNumber .
BoolLiteral  = "true" | "false" .
NullLiteral  = "null" .
ListValue    = "[" [ Expression { "," Expression } [ "," ] ] "]" .
MapValue     = "{" [ MapEntry { "," MapEntry } [ "," ] ] "}" .
MapEntry     = ( Identifier | StringLiteral ) "=" Expression .
```

A reference is `&NAME`; the identifier after `&` must resolve to a record at parse time. A literal is one of the four primitive forms. A parenthesised expression `( Expression )` is a primary expression and groups the inner expression — this is how the precedence rules are overridden:

```
record number RESULT = (1 + 2) * 3   # 9, not 7
```

A list value contains zero or more expressions separated by commas; a trailing comma is permitted. A map value contains zero or more `key = value` pairs separated by commas; a trailing comma is permitted. Map keys may be unquoted identifiers (`HOST`) or string literals (`"host"`); the parser stores both forms identically — the AST normalises to the string form.

### Comments are statements with attachment

A `TokenComment` is a valid statement on its own (a "free-floating" comment), but the printer's responsibility is to attach a run of comments to the **immediately following** non-comment statement so they survive a round-trip. Implementations should expose a `LeadingComments []string` field on `Record`, `Block`, `Import`, `TypeAlias`, and `Conditional` that the parser populates and the printer emits before the node's own output.

A trailing run of comments at end-of-file (after every non-comment statement) attaches to the `File` itself as `TrailingComments` so it round-trips. Without this, an authored file like:

```
record string A = "1"
# trailing thought
```

would lose the trailing thought through `Parse → Print → Parse`.

### Parser error recovery

When the parser encounters a malformed statement, it does **not** attempt to recover and continue parsing. The first error is returned and the rest of the input is discarded. This is intentional: KVRX files are small (typical files are < 200 lines), and partial-AST recovery would force every consumer to handle the error-tolerance complexity even when their input is well-formed.

Specifically:

- An unexpected token at any position returns `UnexpectedTokenError{Got: <token>, Want: [<token types>]}` and stops parsing. The `Want` slice lists every token type the production at that point would have accepted.
- An unexpected end-of-input returns `UnexpectedEndOfTokensError`.
- A type-checking error (`TypeMismatchError`, `BlockKeyError`, `UndeclaredReferenceError`, etc.) stops parsing at the offending site, even though parsing per se could continue past it. This keeps the rule simple: the parser returns a partial-but-consistent `*File` only if the entire input parses; otherwise it returns `nil, err`.

A consumer that wants tolerance can wrap `Parse` in a custom retry-with-truncation strategy — strip the offending statement and re-parse — but the parser itself does not.

### Source positions on AST nodes

Every AST node produced by the parser carries a `Pos` field identifying the first rune of the construct in the source.

- `Record.Pos` is the position of the `record` keyword.
- `Block.Pos` is the position of the `block` keyword.
- `Import.Pos` is the position of the `import` keyword.
- `TypeAlias.Pos` is the position of the `type` keyword.
- `Conditional.Pos` is the position of the `if` keyword.
- `Reference.Pos` is the position of the `&` rune.
- Literal nodes (`StringLiteral`, `NumberLiteral`, `BoolLiteral`, `NullLiteral`) carry the position of the literal's first rune.
- Composite expressions (`OrExpr`, `AndExpr`, `CompareExpr`, etc.) carry the position of the leftmost operand.
- `ListValue.Pos` and `MapValue.Pos` are the positions of the opening `[` or `{`, respectively.

Positions are not preserved through round-trip — the printer does not have re-position information when it writes to a buffer. A re-parsed `*File` will have `Pos` values reflecting its new printed form, not the original source. Tests that compare a re-parsed AST to an originally-parsed AST should use a pos-stripping helper (`kvrx.StripPositions(file)`) before `require.Equal`.

### Statement disambiguation

Several statement forms begin with identifier tokens (`record`, `block`, `import`, `type`, `if`). The parser disambiguates by peeking at the first token's value:

| First-token value           | Statement form           |
|-----------------------------|--------------------------|
| `record`                    | Record                   |
| `block`                     | Block                    |
| `import`                    | Import                   |
| `type`                      | TypeAlias                |
| `if`                        | Conditional              |
| anything else (Identifier)  | error — not a statement opener |
| `TokenComment`              | leading-comment attachment |
| `TokenNewline`              | absorb (soft terminator) |
| `EOF`                       | end of file              |

A non-keyword identifier at statement position is rejected with `UnexpectedTokenError{Got: <ident>, Want: [<the five keyword forms>]}`. There is no fall-through "expression statement" production — every top-level item is one of the five named statement forms.

### Expression precedence summary

The expression productions above encode the precedence (lowest to highest):

1. `||` (logical OR)
2. `&&` (logical AND)
3. `!` (logical NOT, unary)
4. `==` `!=` `<` `<=` `>` `>=` (comparison)
5. `+` `-` (binary additive)
6. `*` `/` (binary multiplicative)
7. unary `-`
8. references, literals, parenthesised expressions, list values, map values

Comparison operators are non-associative — `a < b < c` is rejected with a `ChainedComparisonError`. All other binary operators are left-associative.

## Semantics

### Identifiers and case sensitivity

- **Identifier case sensitivity**: identifiers are case-sensitive. `RED` and `red` are distinct keys.
- **Reserved words are case-sensitive**: `record`, `block`, etc., are reserved only in their lowercase form. `Record` is a valid identifier (and would be tokenised as `TokenIdentifier{Value: "Record"}`, not as the `record` keyword).
- **Whitespace fidelity**: blank lines, trailing whitespace, and the exact column of a token are not preserved across a round-trip. Only comments and structural content round-trip.

### Record key uniqueness

Keys are not required to be unique within a single scope. A KVRX file with two records sharing a key is well-formed; the consumer of the AST decides which wins. The standard library helper `kvrx.Lookup(file, key)` returns the **last** record with the given key (the most recently declared one).

```
record string A = "first"
record string A = "second"

# Lookup(file, "A") returns the "second" record
```

The same applies to block names — `block X { ... }` followed later by `block X { ... }` is well-formed; `kvrx.LookupBlock(file, "X")` returns the latter. This is intentional — KVRX is designed for layered configuration where later declarations override earlier ones.

### Type-and-value agreement

An expression's value must be assignable to the record's declared type. Assignment rules:

- `string` accepts a string literal or any expression whose static type is `string`.
- `number` accepts a number literal or any expression whose static type is `number`.
- `bool` accepts the `true`/`false` keywords or any expression whose static type is `bool`.
- `null` may be assigned to any named type (it is the bottom value).
- `[T]` (list of T) accepts a list value whose elements are each assignable to T.
- `{K: V}` (map of K→V) accepts a map value whose keys are each assignable to K and whose values to V.
- Named types (`Port`) resolve through their alias and apply the rule for the underlying type.

A mismatch surfaces as `TypeMismatchError{Pos, Want, Got}` carrying the assignment-site position. The error variants for the rule above are summarised in the table:

| Site                              | Error                                        |
|-----------------------------------|----------------------------------------------|
| `record T K = E` where `E :: U`   | `TypeMismatchError{Pos: E.Pos, Want: T, Got: U}` |
| `[T]` populated by `[E1, E2, …]`  | `TypeMismatchError{Pos: Ei.Pos, Want: T, Got: Ui}` for the first offending element |
| `{K: V}` populated by `{ki = Ei}` | `TypeMismatchError` on the offending key or value |
| `&NAME` where `NAME` undeclared   | `UndeclaredReferenceError{Pos, Name}` (not a TypeMismatch) |

### Number type inference

A `TokenNumber` whose source contains a `.` or an `e`/`E` is statically typed `float`; otherwise it is `integer`. Both unify under the named type `number` for record-typing purposes; the inference matters only when the implementation produces typed output (CLI flags, generated code, etc.).

```
record number A = 42        # static type: integer
record number B = 1.5       # static type: float
record number C = 1.5e-3    # static type: float
record number D = 0xFF      # static type: integer
```

### Bool type inference

The keywords `true` and `false` are statically typed `bool`. The result of any comparison or logical operator is statically typed `bool`. Anywhere a `bool` is required (conditional condition, logical operands), a non-`bool` value is rejected with `TypeMismatchError{Want: "bool", Got: <static type>}`.

### Null type inference

The keyword `null` is statically typed `null`. `null` is assignable to any named type — record fields may explicitly be `null`. `null` is **not** assignable to a list or map type — those must be a list or map value (possibly empty). This rules out the ambiguity of "is `null` an empty list or a missing value?".

### Operator typing

Binary operators have the following expected operand types:

| Operator       | Operand types       | Result type  |
|----------------|---------------------|--------------|
| `+`            | number, number      | number       |
| `+`            | string, string      | string (concat) |
| `-` `*` `/`    | number, number      | number       |
| `==` `!=`      | T, T (any matching) | bool         |
| `<` `<=` `>` `>=` | number, number   | bool         |
| `<` `<=` `>` `>=` | string, string   | bool (lexicographic) |
| `&&` `\|\|`    | bool, bool          | bool         |
| `!` (unary)    | bool                | bool         |
| `-` (unary)    | number              | number       |

Comparing values of different types with `==` or `!=` is rejected with `TypeMismatchError{Pos, Want: <left type>, Got: <right type>}`. Comparing `null` with `==` against any value yields a `bool` (the equality is structural — only `null == null` is `true`).

Division by zero in a record's parsed expression is a parse-time error (`DivideByZeroError{Pos}`), not a runtime error — the parser refuses to record a constant whose value cannot be computed.

### References

A reference (`&NAME`) statically resolves to the most recent record with key `NAME` in the enclosing scope chain (innermost block first, then outer blocks, then file scope). The reference's static type is the referenced record's declared type. A reference to an undeclared key fails parse with `UndeclaredReferenceError{Pos, Name}`. A forward reference (referencing a record declared later in the same scope) is rejected the same way — KVRX has no hoisting.

References are evaluated lazily — a reference's runtime value is whatever the consumer of the AST decides (most consumers de-reference recursively at lookup time; some leave references unresolved for templating).

```
record string GREETING = "hello"
record string LOUD = &GREETING + "!"   # &GREETING resolves to the line above, type string
```

### Conditionals at parse time

A conditional's branch expression is evaluated at parse time using the values of previously-declared records and references. The implementation maintains a small environment that maps record keys to their parsed expressions; lookups walk the scope chain. A conditional whose expression cannot be reduced to a literal `bool` (because an operand is a non-`bool` reference, an undefined reference, or a non-comparable type pair) is rejected with `NonStaticConditionalError{Pos}`.

The conditional's branch bodies are **not** type-checked for branches that don't take. This avoids false positives from inactive code that depends on imports the live branch doesn't use.

```
record string MODE = "prod"

if (&MODE == "prod") {
    record number PORT = 443;        # this branch takes
} elif (&MODE == "stage") {
    record number PORT = "8443";     # this would be a type error, but it's inactive
} else {
    record number PORT = 8080;
}
```

### Block typing

A block declared as `block X : Section { ... }` checks each inner record's key against the `Section` type alias. `Section` must resolve to a `MapType` whose keys are `Identifier`s; each inner record's key must appear in the alias's keyset, and each inner record's declared type must be assignable to the alias's value type for that key.

```
type Section = {
    HOST = string,
    PORT = number,
    DARK = bool,
}

block PROD : Section {
    record string HOST = "prod.example.com";
    record number PORT = 443;
    record bool   DARK = true;
}
```

A block typed `: Section` whose inner record has a key not in the alias's keyset is rejected with `BlockKeyError{Pos, Key, Alias}`. A typed block does **not** require every key in the alias to be declared — the alias acts as an upper bound on what may appear, not a contract that every key must appear.

### Imports and circular detection

When a file `A` imports a file `B`, every top-level statement in `B` is parsed and inlined at the `import` statement's position in `A`'s top-level statement list (preserving order). The implementation maintains a stack of currently-being-imported files; if `B` (transitively) imports `A`, the import fails with `CircularImportError{ImportStack}`.

The imported file's filename (the string literal value) is resolved relative to the importing file's directory. Implementations should expose this resolution via a `kvrx.OpenFunc` hook so callers can virtualise the filesystem (in tests, in embedded builds).

Imported records and blocks share a flat top-level scope — there is no namespace prefix. This is deliberate; KVRX is designed for small configuration use cases where hierarchical namespacing is friction. A consumer that wants per-import namespacing can use blocks: `block UNITS { … imported records … }`.

### Lookup precedence

The standard library exposes:

```
kvrx.Lookup(file, key)        — last record with key `key`, or zero Record + ErrNotFound
kvrx.LookupBlock(file, name)  — last block with name `name`, or zero Block + ErrNotFound
kvrx.WithinBlock(file, name)  — kvrx.File-shaped view of a block's inner statements
```

The "last wins" rule applies recursively across imports — an imported record can override a record declared earlier in the importer, and vice versa. Callers that want first-wins semantics can invert the iteration order; the standard library is opinionated on last-wins because it matches how layered configuration typically composes (defaults at the top, overrides at the bottom).

### Order of evaluation

When the parser folds a multi-operand expression, operands are evaluated left-to-right per the grammar's left-associativity. This matters for `+` between mixed literal-and-reference operands:

```
record string MSG = "a" + "b" + &SUFFIX
```

This parses as `Add(Add("a", "b"), &SUFFIX)` (left-associative). The inner `Add("a", "b")` is folded to `"ab"`; the outer `Add("ab", &SUFFIX)` is left as an `AddExpr` because one operand is a reference. A consumer calling `kvrx.Resolve` on this expression sees `"ab" + &SUFFIX` — the resolver does its own folding once references are de-referenced.

Logical operators (`&&`, `||`) are **not** short-circuit during folding — the parser folds both sides regardless of whether one side would short-circuit at runtime. This is acceptable because every operand is type-checked before folding, so folding a `&&` whose left side is `false` does not skip a type error on the right side.

### Round-trip guarantees

- Every record's leading comments survive a round-trip in source order.
- Every record's `LeadingComments` slice is preserved with each comment as one element (the `#` and `/*` `*/` framing is dropped — the printer reapplies it).
- Every block's leading comments survive; the block's inner statements round-trip in declaration order, with their own per-statement leading comments preserved.
- Every import's leading comments survive; the import statement itself round-trips (the imported file's contents are not inlined into the printed output — the import is a structural anchor, not a copy).
- Every conditional's leading comments survive; the conditional's branches round-trip with their declared expressions and bodies (they are **not** evaluated and elided in the printer — print preserves the source).
- Trailing comments at end-of-file round-trip via `File.TrailingComments`.

### Constant folding

The parser folds constant expressions at parse time so the AST stores fully-reduced literal values where possible. This applies to:

- Arithmetic on literal numbers (`1 + 2 * 3` becomes `NumberLiteral{Value: "7"}` in the AST).
- String concatenation with `+` between two string literals (`"a" + "b"` becomes `StringLiteral{Value: "ab"}`).
- Boolean reductions of constant operands (`true && false` becomes `BoolLiteral{Value: "false"}`).
- Comparison of two constants (`1 < 2` becomes `BoolLiteral{Value: "true"}`).
- Unary negation of a number literal (`-5` becomes `NumberLiteral{Value: "-5"}`).
- Parenthesised expressions whose body is a constant (the parens are dropped from the AST).

Folding does **not** apply across references. `&PORT + 1` is left as an `AddExpr` in the AST, even when `PORT` resolves statically to `8080`. The reason is round-trip — a printed `&PORT + 1` should re-emit as `&PORT + 1`, not as the resolved `8081`. The consumer that wants the resolved value uses `kvrx.Resolve(file, expr)`.

Folding is also skipped when it would change observable behaviour: division that would cause `DivideByZeroError` is **not** folded (the error is raised); operations on `null` are not folded (they would change the static type from a more specific type to `null`).

### Diagnostics and error positions

Every parser-time error carries a `Pos` so consumers can render it against the source. The convention is that `Pos` points at the **first offending rune** of the error site — not the surrounding statement, not the leading comment, not the prior token.

- `UnexpectedTokenError.Got.Pos` is the offending token's position.
- `UnexpectedEndOfTokensError` has no `Pos` — by definition, the offense is at end-of-file.
- `TypeMismatchError.Pos` is the value-expression's position (the right-hand side of `=`, or the offending element of a list/map).
- `UndeclaredReferenceError.Pos` is the `&` rune of the reference.
- `CircularImportError` carries an `ImportStack []string` of file paths; each path is the source-resolved file that participated in the cycle, in the order they were entered.
- `BlockKeyError.Pos` is the offending record's `Key` token position; the `Alias` field names the type alias the block was declared against.
- `InvalidMapKeyError.Pos` is the `{` rune of the map type expression; the `KeyType` field is the rejected key type as a printable string.
- `ChainedComparisonError.Pos` is the second comparison operator's position (the one that violated non-associativity).
- `NonStaticConditionalError.Pos` is the `(` rune after `if` / `elif` (i.e. the start of the offending expression).
- `DivideByZeroError.Pos` is the `/` operator's position.

Consumers rendering errors with surrounding context should fetch the source line at `Pos.Line` and underline the rune at `Pos.Column` (1-based). The standard library exposes a `kvrx.FormatError(err, source)` helper that does this rendering for the typed errors above.

### What does not round-trip

KVRX explicitly does not preserve:

- The exact horizontal whitespace between tokens within a statement.
- Blank lines between statements (the printer emits one blank line between top-level statements regardless of source spacing).
- The order of map entries (the printer emits them in source order if available, but does not guarantee preservation across construct-then-print).
- The presence of optional trailing commas in list and map values (the printer always emits them).
- The choice of identifier vs. quoted-string for map keys (the printer always emits them as quoted strings).
- The choice of single-quoted vs. triple-quoted string for content that fits in either (the printer chooses single-quoted unless the value contains a literal newline).

The non-preserved choices were dropped intentionally: each one would force the parser to retain extra information (raw bytes, position spans) the AST does not otherwise carry, and the cost is paid in every consumer's memory and traversal time.

## Examples

### Minimal: empty file

An empty input is a legal KVRX file. `Parse("")` yields `&File{}`; `Print(&File{})` writes nothing.

### Typical: two records with a comment

```
# the universal greeting
record string GREETING = "hello"
record number ANSWER  = 42
```

The AST contains two `Record` values. The first has `LeadingComments: []string{"the universal greeting"}`.

### Composite: a list and a map record

```
record [number] PORTS = [80, 443, 8080]
record {string:string} HEADERS = {
    "Content-Type" = "application/json",
    "Accept" = "*/*",
}
```

Both records round-trip; trailing commas are preserved in the AST but normalised on print.

### Conditional: feature-flagged record

```
record bool ENABLE_TLS = true

if (&ENABLE_TLS) {
    record number PORT = 443;
} else {
    record number PORT = 80;
}
```

The conditional resolves at parse time. The resulting `File` has a `Conditional` AST node whose taken branch produced a `Record` with key `PORT` and value `443`.

### Typed block with an alias

```
type Endpoint = {
    HOST = string,
    PORT = number,
}

block API : Endpoint {
    record string HOST = "api.example.com";
    record number PORT = 443;
}
```

The block's inner records are validated against the `Endpoint` alias. A `record bool DARK = true;` inside `API` would be rejected with `BlockKeyError{Key: "DARK", Alias: "Endpoint"}`.

### Imports composing a configuration

```
# defaults.kvrx
record string MODE = "dev"
record number PORT = 8080
```

```
# main.kvrx
import "defaults.kvrx"

# override the port for production
record number PORT = 443
```

After parsing `main.kvrx`, `kvrx.Lookup(file, "PORT")` returns the `443` record (last-wins).

### Type alias chain

```
type Tag      = string
type Tags     = [Tag]
type TagsByEnv = {string: Tags}

record TagsByEnv ROUTING = {
    "prod"  = ["frontend", "backend"],
    "stage" = ["frontend"],
}
```

The alias chain `TagsByEnv → {string: Tags} → {string: [Tag]} → {string: [string]}` resolves at parse time so the record's value can be type-checked against the fully-resolved type.

### Folded constant arithmetic

```
record number TIMEOUT_SECONDS = 60 * 5      # folded to 300
record number BACKOFF_FLOOR   = 2 * 1000    # folded to 2000
record string GREETING        = "hi" + ", " + "world"   # folded to "hi, world"
```

The folded values appear in the AST as literal nodes; the original expression structure is not preserved. A round-trip prints the folded literals, not the original arithmetic.
