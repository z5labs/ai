# kvr package guide

`kvr` reads, parses, and formats the KVR text file format. The package follows the **tokenizer / parser / printer** pipeline: each component owns one concern and ships in one file.

```
text  ─── Tokenize ─►  iter.Seq2[Token, error]  ─── Parse ─►  *File (AST)  ─── Print ─►  text
                │                                       │                       │
            tokenizer.go                             parser.go               printer.go
```

The implementer's job is to extend each phase against `SPEC.md`. Read the spec, identify the section line ranges for the feature, and pass slices to phase work — never load the whole spec into the working context.

## Action loop state machine

All three components share a recursive action-function pattern. An action does some work and returns the next action to run (or `nil` to stop). A small driver loop calls actions until one returns `nil`.

| Component  | Action type                                                        | `nil` means         |
|------------|--------------------------------------------------------------------|---------------------|
| tokenizer  | `func(t *tokenizer, yield func(Token, error) bool) tokenizerAction` | end of stream        |
| parser     | `func(p *parser, t T) (parserAction[T], error)` (generic over `T`) | success (or err set) |
| printer    | `func(pr *printer, f *File) printerAction`                          | done printing        |

The closure pattern is the workhorse: when an action needs to carry state across calls (start position of a string literal, accumulated digits, current index in a slice), it returns a closure that holds that state.

## Helpers

- **Tokenizer**: `t.next()` advances and updates position; `t.backup()` rewinds. Yield-then-continue and yield-error-and-stop are the two ending shapes for every specialised action.
- **Parser**: `p.expect(types...)` pulls the next token and verifies the type. Use it everywhere the grammar requires a specific token — never inline the type check, since the error must flow through `UnexpectedTokenError`.
- **Printer**: `pr.write(s)` and `pr.writef(format, args...)` short-circuit when `pr.err != nil`. Action bodies don't thread errors; they just call `write` and let `pr.err` accumulate.

## The inner action loop rule

For complex types — anything with nested members, repetition, or alternation (records, blocks, lists, expressions) — implementations **must** use an inner action loop, not an inline `for` with a switch. Each state of the parse (open delimiter, member, separator, close) gets its own `parserAction[*T]`. A flat for-with-switch becomes unmaintainable as states accrete; small named action functions stay readable.

```go
func parseBlock(p *parser, f *File) (parserAction[*File], error) {
    blk := &Block{}
    var err error
    for action := parseBlockOpen; action != nil && err == nil; {
        action, err = action(p, blk)
    }
    if err != nil { return nil, err }
    f.Blocks = append(f.Blocks, *blk)
    return parseFile, nil
}
```

This is the rule most likely to be violated by a fast implementer. Reviewers should send back a flat `for { switch tok.Type { ... } }` for any complex type.

## Testing

- `t.Parallel()` at both the test function and each subtest. Action functions are pure; parallel runs catch hidden global state.
- Table-driven tests (`testCases` slice + `t.Run(tc.name, ...)`) with descriptive lowercase names.
- Assertions via `github.com/stretchr/testify/require` — never `assert`.
- **Parser tests must drive the public `Parse()`.** Constructing AST nodes by hand in tests bypasses the parser, masks regressions, and rewrites the test every time the AST shape changes. The empty-input scaffold case (`Parse("") == &File{}`) is the only allowed exception.
- **Every printer rule gets both a direct test and a round-trip test** (`Parse → Print → Parse → require.Equal`). Round-trip catches drift between parser and printer cheaply; the direct test pins formatting choices the round-trip can't see.
- Run `go test -race ./...` after every change.

## Errors

Every typed error lives in the file of the component that returns it. Add new typed errors as needed — `tokenizer.go` for lexical errors, `parser.go` for grammar errors. Sentinels (`var ErrXxx = errors.New(...)`) are appropriate for cross-cutting failure modes; struct errors carrying context (`UnterminatedStringError{Pos}`) are appropriate for lexical/grammatical errors callers want to inspect.

The hard rule: never use a bare `fmt.Errorf` in the hot path. Tests assert via `errors.As` and `errors.Is`; a stringly-typed error breaks both.
