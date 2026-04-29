# toml

A Go text file library for parsing and formatting TOML files.

This package follows the **tokenizer / parser / printer** pipeline pattern
documented in the repo-level `references/architecture.md`. Each component
owns one concern and can be tested in isolation.

## Pipeline

```
text  ─── Tokenize ─►  iter.Seq2[Token, error]  ─── Parse ─►  *File (AST)  ─── Print ─►  text
                │                                       │                       │
            tokenizer.go                             parser.go               printer.go
```

- `tokenizer.go` — bytes in, lazy `iter.Seq2[Token, error]` out. Owns the
  `*bufio.Reader`, `Pos` tracking, and the `Token` / `TokenType` types.
- `parser.go` — token stream in, `*File` AST out via the public
  `Parse(r io.Reader)` function. Owns the AST node types and the `expect`
  helper.
- `printer.go` — `*File` AST in, formatted text written to an `io.Writer`
  via `Print(w, f)`. Errors accumulate on the printer struct.

## Action-loop state machine

Every component is a state machine expressed as **recursive action
functions**. An action does some work, then returns the next action to
run (or `nil` to stop). A small driver loop calls actions until one
returns `nil`. The shape repeats three times so the package has one
mental model, not three.

The three action types — and what `nil` means for each:

| File           | Type                                                                            | `nil` means                                |
| -------------- | ------------------------------------------------------------------------------- | ------------------------------------------ |
| `tokenizer.go` | `tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction` | end iteration (no more tokens).            |
| `parser.go`    | `parserAction[T any] func(p *parser, t T) (parserAction[T], error)`             | done with this nesting level.              |
| `printer.go`   | `printerAction func(pr *printer, f *File) printerAction`                        | end of output. Errors flow through `pr.err`. |

## Helpers

- **`yieldThen(tok, next)` (tokenizer)** — yield a token, then continue
  with `next`. The most common ending of a tokenizer action.
- **`yieldErrorAndStop(err)` (tokenizer)** — yield an error and return
  `nil`. Every error path uses this so the convention is consistent.
- **`skipWhitespace` (tokenizer)** — pre-wired skip-whitespace action;
  almost every text format needs it.
- **`p.expect(types...)` (parser)** — pulls the next token, returns it if
  its type is in `types`, or `UnexpectedTokenError`. Use it everywhere the
  grammar requires a specific token; never inline the type check.
- **`writeThen(s, next)` (printer)** — write a string, then continue with
  `next`. Same shape as `yieldThen`, opposite direction.

## The inner action loop rule (this is the rule most likely to be broken)

For complex/nested types — anything with nested members, repetition, or
alternation (records, tables, arrays, expressions) — implementations
**must** use an inner action loop, not an inline `for` with a switch:

```go
func parseRecord(p *parser, f *File) (parserAction[*File], error) {
    rec := &Record{}
    var err error
    for action := parseRecordOpen; action != nil && err == nil; {
        action, err = action(p, rec)
    }
    if err != nil { return nil, err }
    f.Records = append(f.Records, rec)
    return parseFile, nil
}
```

Each state of the nested parse — open, member, separator, close — gets
its own `parserAction[*Record]`. Real implementations accrete states
(trailing commas, comments inside a record, nested records) and a flat
switch becomes unreadable and untestable. Action functions stay small,
name themselves, and can be exercised by the parser tests directly.

The same pattern applies to the printer when iterating slices: use a
closure that captures the current index and returns the next action.

## Testing style

- **Parser tests must drive the public `Parse()` with real source
  strings.** Constructing AST nodes by hand bypasses the parser, masks
  regressions, and rewrites the test every time the AST shape changes.
  This is the rule a fast implementer is most likely to break.
- **Printer tests must include round-trips.** Every printer test pairs a
  direct test (AST in, expected string out) with a round-trip
  (`Parse → Print → Parse → require.Equal`). The round-trip catches drift
  between the parser and printer cheaply; the direct test pins down
  formatting choices the round-trip can't see.
- `t.Parallel()` at **both** the test function and each subtest. Action
  functions are pure; parallel tests catch hidden global state.
- Table-driven with a `testCases` slice and `t.Run(tc.name, ...)`. Names
  are lowercase descriptive.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  parser test that keeps running after the first failure produces noise,
  not signal.
- Failure-path subtests use `require.ErrorAs` for typed errors
  (`*UnexpectedTokenError`, `*UnexpectedCharacterError`) and
  `require.ErrorIs` for sentinels.
- Run `go test -race ./...` after every change.

## What to implement next

1. Extract a `SPEC.md` for the TOML format you target (use the
   `extract-text-spec` skill).
2. Replace `File`'s placeholder `Nodes` field with the real top-level
   structure (tables, key/value pairs, comments).
3. Add concrete AST nodes that satisfy the `Type` marker interface.
4. Wire up `tokenizeDispatch` in `tokenizer.go` to route runes to
   per-token actions (string, number, identifier, comment, symbol).
5. Replace the stub `parseFile` with the real top-level parser action,
   using inner action loops for any nested/composite production.
6. Replace the stub `printFile` with real printer actions; add round-trip
   tests for every method as it lands.
7. Run the `implement-go-text-file-library` agent to drive the
   test-first implementation, or fill in tests + implementation by hand.
