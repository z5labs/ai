# toml

A Go library for parsing and formatting [TOML](https://toml.io) documents.

The package implements the canonical text-file-library pipeline:

```
Tokenizer -> Parser -> AST -> Printer
```

## State machine pattern

Each stage of the pipeline is implemented as a state machine of recursive
action functions. An action consumes some input and returns the next
action to run. Returning `nil` terminates the loop.

| Stage     | Action type                                                                  | Driver  |
|-----------|------------------------------------------------------------------------------|---------|
| Tokenizer | `tokenizerAction = func(t *tokenizer, yield func(Token, error) bool) tokenizerAction` | `Tokenize`|
| Parser    | `parserAction[T] = func(p *parser, t T) (parserAction[T], error)`            | `Parse` |
| Printer   | `printerAction = func(pr *printer, f *File) printerAction`                   | `Print` |

Implementations should prefer recursive action dispatch over inline
for-loops. When a single action needs to scan a run of related runes or
tokens (a string body, a multi-line value, etc.), introduce an inner action
loop in a helper function rather than nesting `for` statements inside the
action.

## Helpers

### Tokenizer

- `newTokenizer(io.Reader) *tokenizer` — constructs a tokenizer over a
  buffered reader with 1-indexed position tracking.
- `(*tokenizer).next() (rune, error)` — reads the next rune and advances
  the position.
- `(*tokenizer).backup()` — undoes the most recent `next()`. May only be
  called once between calls to `next()`.
- `yieldErr(yield, err) tokenizerAction` — emits an error and stops the
  state machine.
- `yieldToken(yield, tok, next) tokenizerAction` — emits a token and
  returns `next` (or `nil` if the consumer requested early termination).
- `skipWhitespace(*tokenizer) error` — consumes spaces, tabs, CRs, and
  newlines until a non-whitespace rune or EOF.

### Parser

- `iter.Pull2(Tokenize(r))` is used to drive the parser; the parser stores
  the returned pull function in `parser.next`.
- `(*parser).expect(TokenType) (Token, error)` — pulls the next token and
  asserts its type, returning a typed error if the stream is exhausted or
  the token does not match.

### Printer

- `(*printer).write(s string)` — writes `s` if no prior error has been
  recorded.
- `(*printer).writef(format string, args ...any)` — Printf-style write.
- `writeThen(s string, next printerAction) printerAction` — convenience
  for "write `s`, then continue with `next`".
- The `Print` driver checks `pr.err` between every action so a single
  failed write short-circuits the rest of the document.

## Errors

| Type                         | Meaning                                                |
|------------------------------|--------------------------------------------------------|
| `*UnexpectedCharacterError`  | Tokenizer hit a rune that isn't valid in this context. |
| `*UnexpectedEndOfTokensError`| Parser ran out of tokens while still expecting input.  |
| `*UnexpectedTokenError`      | Parser got a token of the wrong type.                  |

All error types are pointer-receivers and embed enough context (position,
token, expected production) to render a useful message via `Error()`.

## Testing

- Tests are table-driven and call the public surface (`Tokenize`, `Parse`,
  `Print`) with real source strings rather than constructing AST nodes by
  hand. This keeps tests resilient to AST refactors.
- Both the outer `Test*` and the inner `t.Run` subtests call
  `t.Parallel()`.
- Assertions use `github.com/stretchr/testify/require`.
- The `collect` helper in `tokenizer_test.go` drains an
  `iter.Seq2[Token, error]` into a `[]Token` plus a single error.
- `TestPrinterRoundTrip` exercises `Parse` -> `Print` and asserts the
  output matches the original input.
- Run `go test -race ./...` before committing.
