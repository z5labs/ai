# Text File Library Architecture

A Go text file library reads, parses, and formats one text-based format. Three files, three responsibilities:

```
text  ─── Tokenize ─►  iter.Seq2[Token, error]  ─── Parse ─►  *File (AST)  ─── Print ─►  text
                │                                       │                       │
            tokenizer.go                             parser.go               printer.go
```

The whole pipeline is a state machine expressed as **recursive action functions**. Each component has a slightly different action signature, but they all behave the same way: an action does some work, then returns the next action to run (or `nil` to stop). A small driver loop calls actions until one returns `nil`.

The orchestrator hands each phase subagent only the spec slices and source files it needs. This document is the architectural spine those subagents work from.

## 1. Tokenizer (`tokenizer.go`)

The tokenizer turns bytes into a stream of `Token` values, lazily, via `iter.Seq2[Token, error]`.

### Streaming via `iter.Seq2`

```go
func Tokenize(r io.Reader) iter.Seq2[Token, error]
```

`iter.Seq2` lets the parser consume one token at a time without buffering the whole file, and lets the tokenizer surface errors at the position they occur instead of returning a partial slice plus a final error.

### State machine via action functions

```go
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction
```

An action reads some runes, optionally calls `yield(tok, nil)` to emit a token, and returns the next action. Returning `nil` ends iteration.

The closure pattern is the workhorse: when an action needs to capture state across rune reads (the start position of a string literal, the accumulated digits of a number), it returns a closure that holds that state. The tokenizer struct still owns the reader and position cursor — that's its job — but no per-token accumulation fields creep onto the struct, no resumable goroutines, no parser callbacks. Just a function that takes the tokenizer and yields tokens.

### Dispatch from the main action

The top-level tokenizer action peeks one rune and dispatches via switch case. Each branch returns a more specialized action (or a closure) that handles that token type, then chains back to the main action when done. This pattern keeps every tokenizer flat:

```go
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
    r, err := t.next()
    if err == io.EOF { return nil }
    if err != nil    { return yieldErr(err) }
    switch {
    case unicode.IsSpace(r):  return skipWhitespace
    case r == '#':            return tokenizeComment(t.prevPos)  // closure captures rune's start pos (next() already advanced t.pos)
    case r == '"':            return tokenizeString(t.prevPos)
    case unicode.IsDigit(r):  t.backup(); return tokenizeNumber
    // ...
    }
    return yieldErr(&UnexpectedCharacterError{Pos: t.prevPos, Char: r})
}
```

After yielding, every specialized action returns `tokenize` to resume dispatch — never another specialized action and never `nil` (unless EOF was reached during the read).

### The tokenizer struct

Wraps a `*bufio.Reader` for one-rune lookahead, and tracks two positions so every token knows where it came from:
- `pos Pos{Line, Column}` — the position of the *next* rune to be read.
- `prevPos Pos{Line, Column}` — the position of the rune most recently returned by `next()`. `next()` snapshots `pos → prevPos` *before* advancing, so right after `r, err := t.next()` the rune `r` started at `t.prevPos` and `t.pos` already points one past it. Capture `t.prevPos` whenever a closure or error needs the start position of the rune just read; reaching for `t.pos` there shifts every reported column by one.

Two methods:
- `next() (rune, error)` snapshots `prevPos`, then advances `pos`.
- `backup()` rewinds the last rune (used when an action peeks one rune past the end of its token); restores `pos` from `prevPos` so newline boundaries don't underflow.

A position-off-by-one is almost always a `pos`/`prevPos` mix-up or a `next`/`backup` ordering bug — audit those together when tokenizer tests miss by one column.

### Helpers worth pre-wiring

- **Yield-then-continue**: a one-liner that calls `yield(tok, nil)` and returns the dispatch action — the most common ending of any specialized action.
- **Yield-error-and-stop**: calls `yield(Token{}, err)` and returns `nil`. Used by every error path so the convention is consistent.
- **Skip-whitespace**: a tiny action that consumes whitespace runes and chains back to dispatch.

### Errors

Every typed error sits in `tokenizer.go` next to the code that returns it. The starter set is `UnexpectedCharacterError{Pos, Char}`; add format-specific types as the implementer goes (`UnterminatedStringError`, `InvalidEscapeError`). The rule is **typed error per failure mode, never a bare `fmt.Errorf` in the hot path** — the parser and tests assert via `errors.As`, and a stringly-typed error breaks both.

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

The type parameter `T` is the AST node currently being built. The top-level loop runs `parserAction[*File]` actions; an action that's parsing a record runs `parserAction[*Record]` actions over its sub-state. Generic actions let nested parsers use the same loop without an interface dance.

Returning `(nil, nil)` completes successfully. Returning `(nil, err)` terminates with error — every error path returns `nil` for the next action so the loop stays monotone.

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

Each state of the record parse — open delimiter, member, separator, close — gets its own `parserAction[*Record]`. Why this matters: complex parsers grow. A real implementation accretes states (trailing commas, comments inside the record, nested records), and a flat switch becomes unreadable and untestable. Action functions stay small, name themselves, and can be exercised directly by tests if needed.

A reviewer who sees `for { switch tok.Type { ... } }` inside a parser action for a complex type should send it back. The compiler doesn't care — the next person to add a state does.

### `expect`

The parser struct exposes one helper: `expect(types ...TokenType) (Token, error)`. It pulls the next token, checks the type matches one of the given types, returns the token or `UnexpectedTokenError{Got, Want}`. Use it everywhere the grammar requires a specific token; never inline the type check, because doing so duplicates the error-construction logic and tests assert against `UnexpectedTokenError` exclusively.

### Tests drive `Parse()`, never the AST constructors

Parser tests must call the public `Parse()` with real source strings. Constructing AST nodes by hand in tests bypasses the parser, masks regressions, and rewrites the test every time the AST shape changes. The only allowed exception is the empty-input scaffold case (`Parse("") == &File{}`), which exists to prove the loop runs.

### Errors

`UnexpectedEndOfTokensError` and `UnexpectedTokenError{Got, Want}` cover the common cases. Add format-specific types as needed (`DuplicateKeyError`, `UnterminatedBlockError`). All errors live in `parser.go`.

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

When printing a slice (records, statements, members), use a closure that captures the current index and returns either "print the current element then advance" or `nil` when the index is past the end. Same shape as the tokenizer's closure pattern — no mutable iterator state on the printer struct.

```go
func printRecords(records []Record) printerAction {
    var step printerAction
    i := 0
    step = func(pr *printer, f *File) printerAction {
        if i >= len(records) { return printFooter }
        printRecord(pr, records[i])
        i++
        return step
    }
    return step
}
```

### Round-trip is the contract

Every printer test pairs a direct test (AST in, expected string out) with a round-trip test (`Parse → Print → Parse → require.Equal`). The round-trip catches drift between the parser and printer cheaply; the direct test pins formatting choices (whitespace, quoting, punctuation) the round-trip can't see. Both belong in `printer_test.go`.

A round-trip mismatch is almost always one of:
- the parser dropped a token (the printer reproduces what's in the AST, so the AST is missing it),
- the printer omitted punctuation the parser made optional, or
- a struct field changed shape and one side wasn't updated.

Read the AST diff first; the surface text diff is downstream of it.

## When tests fail

- **Position off-by-one** in tokenizer tests → `next()` updates `pos` before yielding instead of after, or `backup()` doesn't rewind position.
- **Round-trip mismatch** → see the printer section above; start with the AST diff.
- **`UnexpectedTokenError` with the wrong `Want`** → an `expect` call has the wrong type list, or a parser action constructed the error directly instead of going through `expect`.
- **Unimplemented stub test passes when it shouldn't** → some scaffolds ship with a stub returning `errUnimplemented`; if the new implementation short-circuits to that sentinel, the stub test stays green even though real input fails. Delete or rewrite the stub test as part of the phase that wires up the real public API.

## Why this shape

A single text format gets one package, and that package gets exactly three files of production code. Real implementations of JSON, INI, or SQL accrete dozens of token types, dozens of AST nodes, and a long tail of formatting rules — a sprawling layout makes the round-trip property impossible to audit at a glance. Three files, three responsibilities, one action-loop pattern repeated three times, round-trip tests on every printer method — that's the contract every phase subagent maintains.
