# graphql package — contributor notes

This package reads, parses, and prints GraphQL schema documents. It follows
the **tokenizer / parser / printer** pipeline the repo uses for every text
file library:

```
text  ── Tokenize ──►  iter.Seq2[Token, error]  ── Parse ──►  *File  ── Print ──►  text
              │                                       │                  │
        tokenizer.go                              parser.go          printer.go
```

Three production files, three responsibilities. Tokens live in
`tokenizer.go`; AST nodes live in `parser.go`; formatting lives in
`printer.go`. Don't introduce a fourth file unless the package outgrows the
shape — and it almost certainly won't.

## The action-loop state machine

Each component is a state machine driven by **recursive action functions**.
A driver loop calls actions until one returns `nil` (or, for the parser, a
non-nil error).

| Component  | Action type                                                                    | What `nil` means              |
| ---------- | ------------------------------------------------------------------------------ | ----------------------------- |
| Tokenizer  | `tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction` | end iteration                 |
| Parser     | `parserAction[T any] func(p *parser, t T) (parserAction[T], error)`            | `(nil, nil)` succeeds; `(nil, err)` terminates with error |
| Printer    | `printerAction func(pr *printer, f *File) printerAction`                       | end printing (errors via `pr.err`) |

Closures are the workhorse. When an action needs to carry state across calls
(start position of a string literal, current index when iterating
definitions), return a closure that captures it. No mutable iterator state
on the struct, no resumable goroutines.

## Helper signatures

- `yieldThen(tok Token, next tokenizerAction) tokenizerAction` — emit tok
  then continue with next. The most common ending of any tokenizer action.
- `yieldErrorAndStop(err error) tokenizerAction` — emit an error and end.
  Every error path uses this so the convention is consistent.
- `skipWhitespace` — consumes whitespace runes and chains back to dispatch.
- `(*parser).expect(types ...TokenType) (Token, error)` — pulls one token
  and verifies type. **Never inline a token-type comparison; always call
  `expect`.**
- `writeThen(s string, next printerAction) printerAction` — write a fixed
  string then continue. Compose with closures that capture iteration state.

## The rule that matters: inner action loop for complex types

For complex types — anything with nested members, repetition, or alternation
(GraphQL types, fields, directives, arguments, descriptions) — implementations
**must** use an inner action loop, not an inline `for` with a `switch`:

```go
func parseObjectType(p *parser, f *File) (parserAction[*File], error) {
    obj := &ObjectType{}
    var err error
    for action := parseObjectTypeOpen; action != nil && err == nil; {
        action, err = action(p, obj)
    }
    if err != nil { return nil, err }
    f.Definitions = append(f.Definitions, obj)
    return parseFile, nil
}
```

This is the single rule a fast implementer is most likely to break. A flat
`for { switch tok.Type {…} }` looks shorter at first but becomes unreadable
the moment trailing commas, embedded comments, or nested types arrive — and
GraphQL has all three. Each state of the parse gets its own
`parserAction[*ConcreteType]`; named functions stay small and testable.

## Testing style

- `t.Parallel()` at both the test function and each subtest. Action
  functions are pure; parallel tests catch hidden global state.
- Table-driven with a `testCases` slice and `t.Run(tc.name, …)`. Names are
  lowercase descriptive sentences.
- Assertions via `github.com/stretchr/testify/require` (not `assert`).
- Run `go test -race ./...` after every change.

### Parser tests

**Parser tests must drive the public `Parse()` with real source strings.**
Constructing AST nodes by hand in tests bypasses the parser, masks
regressions, and rewrites the test every time the AST shape changes. If you
need a built AST for a printer test, get it from `Parse(strings.NewReader(...))`.

### Printer tests

Every printer method needs **two** tests:

1. **Direct test** — AST in (built via `Parse`), expected string out. This
   pins down formatting choices the round-trip can't see.
2. **Round-trip test** — `Parse → Print → Parse → require.Equal`. This is
   the cheapest end-to-end correctness check available; it catches drift
   between the parser and printer almost for free.

`TestPrinterRoundTrip` already has the skeleton — fill in cases as the
grammar grows.

## When tests fail

- **Round-trip mismatch** — the parser dropped a token or the printer
  omitted punctuation the parser made optional. Read the AST diff first.
- **Position off-by-one in tokenizer tests** — `next()` is updating `pos`
  before yielding instead of after, or `backup()` isn't rewinding position.
  Audit those two methods together.
