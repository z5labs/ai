# ini

A Go text file library for the INI configuration file format.

This package follows the **tokenizer / parser / printer** pipeline pattern
documented in the repo-level `references/architecture.md`. Each component
owns one concern and can be tested in isolation.

## Pipeline

```
text  ─── Tokenize ─►  iter.Seq2[Token, error]  ─── Parse ─►  *File (AST)  ─── Print ─►  text
                │                                       │                       │
            tokenizer.go                             parser.go               printer.go
```

- `tokenizer.go` — turns runes into a stream of `Token` values, lazily, via
  `iter.Seq2[Token, error]`. Wraps a `*bufio.Reader` for one-rune lookahead
  and tracks `Pos{Line, Column}` so every token records where it came from.
- `parser.go` — pulls tokens from `iter.Pull2(Tokenize(r))` and builds the
  `*File` AST. Generic `parserAction[T]` lets nested parsers share the same
  loop without an interface dance.
- `printer.go` — formats a `*File` back to text. Errors flow through
  `pr.err` so action bodies stay focused on what to write next.

## State machine via action functions

Every component is a state machine expressed as recursive action functions.
An action does some work, then returns the next action to run (or `nil` to
stop). A small driver loop calls actions until one returns `nil`.

```go
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)
type printerAction func(pr *printer, f *File) printerAction
```

The closure pattern is the workhorse: when an action needs to capture state
across calls (the start position of a string literal, the accumulated
digits of a number, the current index in a slice), it returns a closure
that holds that state. No mutable iterator fields, no resumable goroutines.

## Helpers worth knowing

### Tokenizer (`tokenizer.go`)

- `yieldToken(tok, yield) tokenizerAction` — emit a token and return the
  dispatch action; the most common ending of any action.
- `yieldError(err, yield) tokenizerAction` — emit an error and return
  `nil`. Use this on every error path so the convention is consistent.
- `skipWhitespace` — consumes whitespace runes and chains back to dispatch.
- `(*tokenizer).next()` — reads one rune, advances `pos`.
- `(*tokenizer).backup()` — rewinds the most recent rune; only valid once
  per `next()` call.

### Parser (`parser.go`)

- `(*parser).expect(types ...TokenType) (Token, error)` — pull the next
  token and require its type matches one of the given types. Use this
  everywhere the grammar requires a specific token; never inline the type
  check.
- The generic `parserAction[T]` lets nested parsers reuse the same
  driver-loop shape with a more specific node type.

### Printer (`printer.go`)

- `(*printer).write(s string)` — write a literal string; short-circuits if
  `pr.err` is already set.
- `(*printer).writef(format string, args ...any)` — `fmt.Fprintf`-style
  sibling of `write`.
- `writeThen(s, next) printerAction` — one-liner for "emit a literal then
  continue with the next action".

## The inner action loop pattern (the rule that matters)

For complex types — anything with nested members, repetition, or
alternation (sections, key/value lists, expressions) — implementations
**must** use an inner action loop, not an inline `for` with a switch:

```go
func parseSection(p *parser, f *File) (parserAction[*File], error) {
    sec := &Section{}
    var err error
    for action := parseSectionOpen; action != nil && err == nil; {
        action, err = action(p, sec)
    }
    if err != nil { return nil, err }
    f.Nodes = append(f.Nodes, sec)
    return parseStart, nil
}
```

Each state of the section parse — open bracket, name, close bracket, body
— gets its own `parserAction[*Section]`. Action functions stay small, name
themselves, and can be exercised by the parser tests independently.

## Testing style

- Source string in, AST or token slice out — never hand-construct AST nodes
  in parser tests. Constructing AST by hand bypasses the parser, masks
  regressions, and rewrites the test every time the AST shape changes.
- Table-driven with a `testCases` slice and `t.Run(tc.name, ...)`. Names
  are lowercase descriptive.
- `t.Parallel()` at **both** the test function and each subtest. Action
  functions are pure; parallel tests catch hidden global state.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  parser test that keeps running after the first failure produces noise,
  not signal.
- Run `go test -race ./...` after every change.

## Three test shapes

1. **Tokenizer tests** (`tokenizer_test.go`): source string in, `[]Token`
   out. The `collect` helper drains the `iter.Seq2` so subtests stay close
   to one-liners. Position values are exact, not approximate.
2. **Parser tests** (`parser_test.go`): source string in, `*File` out via
   the public `Parse()`. One subtest per scenario from the spec's examples.
   Failure-path subtests use `require.ErrorAs` for typed errors and
   `require.ErrorIs` for sentinels.
3. **Printer tests** (`printer_test.go`): direct (AST in, string out) plus
   round-trip (`Parse → Print → Parse → require.Equal`). Both are required
   for every printer method once the implementer adds real ones.

## Error types

- `UnexpectedCharacterError{Pos, Char}` — tokenizer sentinel for a rune no
  action wanted at the current position.
- `UnexpectedEndOfTokensError{Want}` — parser sentinel for a token stream
  exhausted in a position that requires more input.
- `UnexpectedTokenError{Got, Want}` — parser sentinel for a token of the
  wrong type. Returned by `parser.expect`.

The rule is "typed error per failure mode, never a bare `fmt.Errorf` in
the hot path", so the parser and tests can assert via `errors.As`.

## What to implement next

1. Fill in `SPEC.md` for the real INI dialect you're targeting (sections,
   comments, quoting rules, escape sequences).
2. Replace the placeholder `File.Nodes` field with the real top-level AST
   shape (e.g. `Sections []Section`).
3. Add concrete AST node types (`Section`, `KeyValue`, `Comment`, ...) and
   make each implement the `Type` marker.
4. Expand the `tokenizeStart` dispatch with the format's real token rules.
5. Replace the `parseStart` stub with the real top-level grammar; use the
   inner action loop for any node with nested structure.
6. Implement `printStart` (and per-node printers) so every parser node has
   a corresponding printer that round-trips.
7. Run the `implement-go-text-file-library` agent (or fill in tests +
   implementation by hand) to drive the test-first build-out.
