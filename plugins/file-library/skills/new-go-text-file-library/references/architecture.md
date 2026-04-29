# Text File Library Architecture

A Go text file library is a package that reads, parses, and formats one text-based format. It follows a **tokenizer / parser / printer** pipeline, mirroring `go/scanner` + `go/parser` + `go/printer` in shape but specialized to a single language.

The pipeline is intentionally narrow. Each component owns one concern, and each can be tested in isolation:

```
text  ─── Tokenize ─►  iter.Seq2[Token, error]  ─── Parse ─►  *File (AST)  ─── Print ─►  text
                │                                       │                       │
            tokenizer.go                             parser.go               printer.go
```

The whole pipeline is a state machine expressed as **recursive action functions**. Each component has a slightly different action signature, but they all behave the same way: an action does some work, then returns the next action to run (or `nil` to stop). A small driver loop calls actions until one returns `nil`.

This shape is the load-bearing decision for the package. It scales — DNS, JSON, INI, and SQL all fit it without rearrangement — and it tests cleanly because actions are pure functions of `(state, input) → next action`. The scaffolding skill produces stubs for each; the implementer (or the `implement-go-text-file-library` agent) fills them in against a real `SPEC.md`.

## 1. Tokenizer (`tokenizer.go`)

The tokenizer turns bytes into a stream of `Token` values, lazily, via `iter.Seq2[Token, error]`.

### Streaming via `iter.Seq2`

The public surface is one function:

```go
func Tokenize(r io.Reader) iter.Seq2[Token, error]
```

`iter.Seq2` lets the parser consume one token at a time without buffering the whole file, and lets the tokenizer surface errors at the position they occur instead of returning a partial slice plus a final error. The parser pulls tokens with `iter.Pull2` (see Parser).

### State machine via action functions

```go
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction
```

An action reads some runes, optionally calls `yield(tok, nil)` to emit a token, and returns the next action. Returning `nil` ends the iteration.

The closure pattern is the workhorse: when an action needs to capture state across rune reads (the start position of a string literal, the accumulated digits of a number), it returns a closure that holds that state. The tokenizer still owns the reader and position cursor — that's its job — but no per-token accumulation fields creep onto the struct, no resumable goroutines, no parser callbacks. Just a function that takes the tokenizer and yields tokens.

### The tokenizer struct

Wraps a `*bufio.Reader` for one-rune lookahead, and tracks `Pos{Line, Column}` so every token knows where it came from. Two methods: `next() (rune, error)` advances and updates position, `backup()` rewinds the last rune (used when an action peeks one rune past the end of its token).

### Helpers worth pre-wiring

- **Yield-then-continue**: a one-liner that calls `yield(tok, nil)` and returns the main dispatch action — the most common ending of any action.
- **Yield-error-and-stop**: calls `yield(Token{}, err)` and returns `nil`. Used by every error path so the convention is consistent.
- **Skip-whitespace**: a tiny action that consumes whitespace runes and chains back to dispatch. Almost every text format needs this.

### Errors

A typed `UnexpectedCharacterError{Pos, Char}` covers the most common failure (a rune that no action wanted). Add format-specific error types as the implementer goes — the rule is "typed error per failure mode, never a bare `fmt.Errorf` in the hot path", so the parser and tests can assert via `errors.As`.

## 2. Parser (`parser.go`)

The parser turns the token stream into an AST rooted at `*File`.

### Public surface

```go
func Parse(r io.Reader) (*File, error)
```

Internally: `next, stop := iter.Pull2(Tokenize(r)); defer stop()`, then run the top-level action loop against a `*File`. Pull-based consumption pairs naturally with the action signature — each action calls `next()` zero or more times and decides what to do.

### Generic action type

```go
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)
```

The type parameter `T` is the AST node currently being built. The top-level loop runs `parserAction[*File]` actions; an action that's parsing a record runs `parserAction[*Record]` actions over its sub-state; and so on. Generic actions let nested parsers use the same loop without an interface dance.

Returning `(nil, nil)` completes successfully. Returning `(nil, err)` terminates with error — every error path returns `nil` for the next action so the loop is monotone.

### The inner action loop pattern (the rule that matters)

For complex types — anything with nested members, repetition, or alternation (records, objects, lists, expressions) — implementations **must** use an inner action loop, not an inline `for` with a switch. This is the single most important rule in the parser:

```go
func parseRecord(p *parser, f *File) (parserAction[*File], error) {
    rec := &Record{}
    var err error
    for action := parseRecordOpen; action != nil && err == nil; {
        action, err = action(p, rec)
    }
    if err != nil { return nil, err }
    f.Records = append(f.Records, *rec)
    return parseFile, nil
}
```

Each state of the record parse — open brace, member, separator, close — gets its own `parserAction[*Record]`. The reason is that complex parsers grow: a real implementation accretes states (trailing commas, comments inside the record, nested records), and a flat switch becomes unreadable and untestable. Action functions stay small, name themselves, and can be exercised by the parser tests if needed.

### `expect`

The parser struct exposes one helper: `expect(types ...TokenType) (Token, error)`. It pulls the next token, checks the type matches one of the given types, returns the token or `UnexpectedTokenError{Got, Want}`. Use it everywhere the grammar requires a specific token; never inline the type check.

### Tests drive `Parse()`, never the AST constructors

Parser tests must call the public `Parse()` with real source strings. Constructing AST nodes by hand in tests bypasses the parser, masks regressions, and rewrites the test every time the AST shape changes. The package `CLAUDE.md` should call this out — it's the rule a fast implementer is most likely to break.

## 3. Printer (`printer.go`)

The printer formats the AST back to text. Same action-loop shape, opposite direction.

### Public surface

```go
func Print(w io.Writer, f *File) error
```

### Action type

```go
type printerAction func(pr *printer, f *File) printerAction
```

Note: no error return. Errors flow through `pr.err`.

### Error accumulation pattern

The printer wraps `io.Writer` and stores `err error` as a field. Every write goes through `pr.write(s)` or `pr.writef(...)`, which short-circuits when `pr.err != nil`:

```go
func (pr *printer) write(s string) {
    if pr.err != nil { return }
    _, pr.err = io.WriteString(pr.w, s)
}
```

The driver loop checks `pr.err` each iteration and returns it once any action sets it. This keeps action bodies clean — actions don't need to thread `error` through every call, and an early write error terminates the loop without poisoning the rest of the output.

### Closure pattern for iteration

When printing a slice (records, statements, tokens), use a closure that captures the current index and returns either "print the current element then advance" or `nil` when the index is past the end. Same shape as the tokenizer's closure pattern — no mutable iterator state on the printer struct.

### Round-trip is the contract

Every printer test pairs a direct test (AST in, expected string out) with a round-trip test (`Parse → Print → Parse → require.Equal`). The round-trip catches drift between the parser and printer cheaply; the direct test pins down formatting choices (whitespace, quoting, punctuation) the round-trip can't see. Both belong in `printer_test.go`.

## 4. Tests

Tests are how you know the pipeline is honest.

### Conventions

- `t.Parallel()` at both the test function and each subtest. Action functions are pure; parallel tests catch hidden global state if any sneaks in.
- Table-driven with `testCases` slice and `t.Run(tc.name, ...)`. Names are lowercase descriptive.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a parser test that keeps running after the first failure produces noise, not signal.
- Run `go test -race ./...` after every change.

### Three test shapes

1. **Tokenizer tests** (`tokenizer_test.go`): source string in, `[]Token` out. A `collect` helper drains the `iter.Seq2` so subtests stay one-liner-ish. Position values are exact, not approximate — getting them right early saves debugging time later.
2. **Parser tests** (`parser_test.go`): source string in, `*File` out via the public `Parse()`. One subtest per scenario from the spec's examples. Failure-path subtests use `require.ErrorAs` for typed errors and `require.ErrorIs` for sentinels.
3. **Printer tests** (`printer_test.go`): direct (AST in, string out) plus round-trip (`Parse → Print → Parse → Equal`). Both are required for every printer method once the implementer adds real ones.

### When tests fail

- A round-trip mismatch is almost always either a parser dropping a token (the printer reproduces what's in the AST, so the AST is missing it) or a printer omitting punctuation the parser made optional. Read the AST diff first.
- A position-off-by-one in tokenizer tests usually means `next()` updates `pos` before yielding instead of after, or `backup()` doesn't rewind position. Audit those two methods together.

## 5. Why this shape

A single text format gets one package, and that package gets exactly three files of production code (plus types, but tokens and AST nodes live with the component that produces them). The constraint matters because text formats accrete: a real implementation of JSON or SQL ends up with dozens of token types, dozens of AST nodes, and a long tail of formatting rules, and a sprawling file layout makes the round-trip property impossible to audit at a glance.

Three files, three responsibilities, one action-loop pattern repeated three times, round-trip tests on every printer method — that's the contract. Everything in the scaffold exists to make that contract obvious from the moment the package is created.
