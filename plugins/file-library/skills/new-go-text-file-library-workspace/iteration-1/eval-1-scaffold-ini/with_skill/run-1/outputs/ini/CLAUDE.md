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

- `tokenizer.go` — turns bytes into `Token` values lazily via
  `iter.Seq2[Token, error]`. Owns `Pos`, `TokenType`, `Token`, the
  unexported `tokenizer` struct, the `tokenizerAction` state-machine type,
  and the public `Tokenize(r io.Reader) iter.Seq2[Token, error]`.
- `parser.go` — pulls tokens via `iter.Pull2` and builds the AST rooted at
  `*File`. Owns the `parser` struct, the generic `parserAction[T]` type,
  the `Type` marker interface, and the public
  `Parse(r io.Reader) (*File, error)`.
- `printer.go` — formats `*File` back to text. Owns the `printer` struct
  with its accumulated `err`, the `printerAction` type, and the public
  `Print(w io.Writer, f *File) error`.

## The action-loop state machine

Every component is driven by a small loop that calls action functions until
one returns `nil`. Each action does some work and returns the next action.
The signatures and `nil` semantics differ slightly per component:

| Component  | Action type                                          | Returning `nil` means                       |
| ---------- | ---------------------------------------------------- | ------------------------------------------- |
| Tokenizer  | `func(t *tokenizer, yield) tokenizerAction`          | end iteration                               |
| Parser     | `func(p *parser, t T) (parserAction[T], error)`      | done with this loop (success or `err != nil`) |
| Printer    | `func(pr *printer, f *File) printerAction`           | end iteration; errors live on `pr.err`      |

Why this shape:

- Actions are pure functions of `(state, input) → next action`. They are
  trivially testable in isolation.
- Closures capture per-state data (start position of a string, accumulated
  digits of a number, current index of a slice) without any mutable fields
  on the state struct.
- Adding a new state means adding a function, not editing a sprawling
  switch.

## Helper signatures and when to use them

### Tokenizer (`tokenizer.go`)

- `yieldThen(tok Token, next tokenizerAction) tokenizerAction` — emit a
  token and continue with `next`. The most common ending of any
  tokenizer action.
- `yieldErrorAndStop(err error) tokenizerAction` — emit a token-error and
  end iteration. Use this on every error path; it keeps shutdown uniform.
- `skipWhitespace(next tokenizerAction) tokenizerAction` — consume runs of
  whitespace and chain back to `next`. Almost every text format calls
  this between tokens.

### Parser (`parser.go`)

- `(*parser).expect(types ...TokenType) (Token, error)` — pull the next
  token and require its type is in the allowed set. Returns
  `*UnexpectedTokenError` on mismatch and `*UnexpectedEndOfTokensError`
  on premature exhaustion. **Never inline the type check** — always go
  through `expect`, so error reporting is consistent and tests can
  assert via `errors.As`.

### Printer (`printer.go`)

- `(*printer).write(s string)` / `(*printer).writef(fmt, args...)` — the
  only way to emit output. Both short-circuit when `pr.err != nil`, so
  action bodies never need to thread `error` through every call.
- `writeThen(s string, next printerAction) printerAction` — emit some
  text and continue with `next`. The printer counterpart of
  `yieldThen`.

## The inner action loop pattern (the rule that matters)

For complex types — anything with nested members, repetition, or
alternation (sections, key/value lists, multi-line values) — implementations
**must** use an inner action loop, **not** an inline `for` with a switch.
This is the single most important rule in the parser and printer:

```go
func parseSection(p *parser, f *File) (parserAction[*File], error) {
    sec := &Section{}
    var err error
    for action := parseSectionHeader; action != nil && err == nil; {
        action, err = action(p, sec)
    }
    if err != nil {
        return nil, err
    }
    f.Nodes = append(f.Nodes, sec)
    return parseFile, nil
}
```

Each state of the parse — header, body member, blank line, close — is its
own `parserAction[*Section]`. The reason is that complex parsers grow: a
real INI implementation accretes states (comments inside sections,
trailing whitespace, line continuations), and a flat switch becomes
unreadable and untestable. Action functions stay small, name themselves,
and can be exercised by tests if needed.

The same rule applies to the printer when emitting a slice of children —
hand iteration off to a closure that captures the index and returns either
"print element i, then advance" or `nil` when `i` is past the end.

## Testing style

- Table-driven tests: a `testCases` slice, then `t.Run(tc.name, ...)`.
  Names are lowercase descriptive.
- `t.Parallel()` at **both** the test function and each subtest. Action
  functions are pure; parallel tests catch hidden global state if any
  sneaks in.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  parser test that keeps running after the first failure produces noise,
  not signal.
- **Parser tests must drive the public `Parse()` with real source
  strings.** Constructing AST nodes by hand in tests bypasses the parser,
  masks regressions, and rewrites the test every time the AST shape
  changes. This is the rule a fast implementer is most likely to break.
- **Printer tests must include a round-trip for every printer method**:
  `Parse → Print → Parse → require.Equal`. The round-trip catches drift
  between the parser and printer cheaply; pair it with a direct test
  (AST in, expected string out) that pins down formatting choices the
  round-trip can't see.
- Failure-path subtests use `require.ErrorAs` for typed errors
  (`*UnexpectedCharacterError`, `*UnexpectedTokenError`,
  `*UnexpectedEndOfTokensError`) and `require.ErrorIs` for sentinels.
- After every change run `go test -race ./...`.

## What to implement next

1. Extract a `SPEC.md` from the INI grammar you're targeting using the
   `extract-text-spec` skill.
2. Replace `File`'s placeholder `Nodes []Type` with the real AST shape
   (e.g. `Sections []*Section`, `Globals []*KeyValue`).
3. Add concrete AST node types alongside `parser.go` — each must
   implement `isType()` to satisfy the `Type` marker interface.
4. Wire up the tokenizer dispatch (`tokenizeMain`) for comments,
   identifiers, symbols (`=`, `[`, `]`), strings, and numbers.
5. Wire up the parser dispatch (`parseFile`), using one inner action
   loop per nested structure.
6. Wire up the printer (`printFile`) symmetrically, with a closure-based
   iterator for any slice in the AST.
7. Run the `implement-go-text-file-library` agent to drive the
   test-first implementation, or fill in tests + implementation by
   hand.
