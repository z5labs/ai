# graphql package

This package reads and writes GraphQL schema documents. It follows the
**Tokenizer -> Parser -> AST -> Printer** pipeline pattern shared by all text
file libraries in this repo.

## State Machine Pattern

Each pipeline component is implemented as a state machine of recursive action
functions. Each action takes the component's state, performs one step of work,
and returns the next action to run. Returning `nil` (and `nil` error where
applicable) terminates the loop.

### Tokenizer (`tokenizer.go`)

```go
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction
```

The tokenizer wraps a `*bufio.Reader` and tracks the current source position.
The entry-point action is `tokenize`. Sub-actions read additional runes from
`t.next()`, optionally `t.backup()` if they read one too many, and yield a
completed `Token` via the `yieldToken` helper. The closure pattern is used
when an action needs to capture state (such as the start position of the
current token) before recursing.

`Tokenize(r io.Reader) iter.Seq2[Token, error]` is the public entry point. It
constructs a tokenizer and drives the action loop, surfacing errors through
the iterator's error channel.

### Parser (`parser.go`)

```go
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)
```

The parser pulls tokens through a `next func() (Token, error, bool)` callback
backed by `iter.Pull2`. Actions are generic over the AST node currently being
populated, so the same shape can be reused for `*File`, individual definitions,
and nested types.

Return values:
- `(action, nil)` — continue with `action`.
- `(nil, nil)` — finished successfully.
- `(nil, err)` — terminate with `err`.

For complex types (those with nested members like object types, input objects,
unions, enums, fields with arguments), use the inner action loop pattern. Do
**not** mix inline `for` loops with direct logic for these constructs.

`Parse(r io.Reader) (*File, error)` is the public entry point.

### Printer (`printer.go`)

```go
type printerAction func(pr *printer, f *File) printerAction
```

The printer wraps an `io.Writer` and records the first I/O error in `pr.err`.
Every helper (`pr.write`, `pr.writef`) is a no-op once `err` is set, and the
`Print` driver checks `pr.err` after each action so the loop short-circuits
cleanly.

`Print(w io.Writer, f *File) error` is the public entry point.

## Helpers

| Helper | Signature | Purpose |
| --- | --- | --- |
| `yieldErr` | `func(yield, err) tokenizerAction` | Emit an error and stop the tokenizer. |
| `yieldToken` | `func(yield, tok, next) tokenizerAction` | Emit a token; respect the consumer's stop signal. |
| `(*tokenizer).skipWhitespace` | `() (rune, error)` | Consume whitespace; return the next significant rune. |
| `(*parser).expect` | `(want TokenType, description string) (Token, error)` | Pull a token and require its type. |
| `(*printer).write` | `(s string)` | Emit a literal string; no-op once `pr.err` is set. |
| `(*printer).writef` | `(format string, args ...any)` | Emit a formatted string. |
| `writeThen` | `(s string, next printerAction) printerAction` | Helper action that writes a string and transitions. |

## Testing

- Use table-driven tests with a `testCases` slice.
- Call `t.Parallel()` at both the top-level test and inside each subtest.
- Run subtests with `t.Run(tc.name, ...)`.
- Use `github.com/stretchr/testify/require` for assertions (not `assert`).
- Name test cases descriptively in lowercase.
- Parser tests must call `Parse()` against real source strings rather than
  constructing AST nodes manually.
- Printer tests come in two flavors: direct print tests (AST in -> string
  out) and round-trip tests (Parse -> Print -> Parse -> compare).
- The `collect` helper in `tokenizer_test.go` drains an `iter.Seq2[Token, error]`
  into a `[]Token` for assertion.

Run `go test -race ./...` after each change.

## Errors

- `*UnexpectedCharacterError` — tokenizer saw a rune it could not classify.
- `*UnexpectedEndOfTokensError` — parser ran out of tokens mid-rule.
- `*UnexpectedTokenError` — parser saw a token whose type did not match the
  current grammar rule.

All error types include enough context (position, expected description, or the
offending token) to produce a useful diagnostic.
