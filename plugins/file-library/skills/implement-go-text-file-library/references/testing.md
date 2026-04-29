# Testing Conventions for Text File Libraries

Tests are how you know the pipeline is honest. These conventions apply to every phase (tokenizer, parser, printer) — write the tests first, watch them fail for the right reason, then implement.

## Universal rules

- `t.Parallel()` at both the test function and each subtest. Action functions are pure; parallel tests catch hidden global state if any sneaks in.
- Table-driven with a `testCases` slice and `t.Run(tc.name, ...)`. Names are lowercase, descriptive: `"comment_at_start_of_line"`, `"empty_record"`, `"string_with_escaped_quote"`.
- Assertions via `github.com/stretchr/testify/require` — never `assert`. A pipeline test that keeps running after the first failure produces noise, not signal.
- Run `go test -race ./...` after each step. The race flag catches concurrent reuse of the tokenizer/parser/printer structs and is cheap.

## Phase 1 — tokenizer tests (`tokenizer_test.go`)

### The `collect` helper

Every tokenizer test file gets a small helper that drains the iterator into a flat slice:

```go
func collect(seq iter.Seq2[Token, error]) ([]Token, error) {
    var tokens []Token
    var err error
    for tok, e := range seq {
        if e != nil { err = e; break }
        tokens = append(tokens, tok)
    }
    return tokens, err
}
```

This keeps each subtest a one-liner over the assertion: `tokens, err := collect(Tokenize(strings.NewReader(tc.input)))`.

### Happy path: source string in, `[]Token` out

```go
testCases := []struct {
    name   string
    input  string
    want   []Token
}{
    {
        name:  "single_identifier",
        input: "hello",
        want: []Token{
            {Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "hello"},
        },
    },
}
```

### Position values are exact

`Pos{Line, Column}` values in tests are the exact values you expect — not approximate, not "any non-zero". Getting them right early saves hours of debugging when a higher-level format error reports the wrong location to the user. A position-off-by-one test failure is a real bug, not a flaky test.

### Failure path

```go
_, err := collect(Tokenize(strings.NewReader("\"unterminated")))
var ute *UnterminatedStringError
require.ErrorAs(t, err, &ute)
require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
```

Every typed error in `tokenizer.go` gets at least one test that asserts via `errors.As`. Sentinels (if any) get `errors.Is`.

## Phase 2 — parser tests (`parser_test.go`)

### Tests drive `Parse()` — never construct AST nodes by hand

This is the rule the parser phase is most likely to break. The expected value comes from running `Parse()` over a known-good source string, **not** from hand-constructing a `*File` with literal `Record{Type: ..., Value: ...}` entries:

```go
// WRONG — hand-constructs the AST, hides parser regressions
want := &File{Records: []Record{{Type: "string", Value: "hello"}}}

// RIGHT — uses Parse on a canonical source string as the expectation source
want, err := Parse(strings.NewReader(`record string = "hello"`))
require.NoError(t, err)
got, err := Parse(strings.NewReader(testInput))
require.NoError(t, err)
require.Equal(t, want, got)
```

The only allowed exception is the empty-input scaffold case: `Parse("") == &File{}`. Document the rule in the package's `CLAUDE.md` so a fast implementer doesn't reach for the convenient-but-broken approach.

For tests that actually want to assert the AST shape (not a parser equivalence), use `require.Equal` against fields of the parsed result, not against a literal struct: `require.Len(t, got.Records, 2); require.Equal(t, "string", got.Records[0].Type)`.

### Failure path

```go
_, err := Parse(strings.NewReader("record = }"))
var ute *UnexpectedTokenError
require.ErrorAs(t, err, &ute)
require.Equal(t, TokenSymbol, ute.Got.Type)
require.Contains(t, ute.Want, TokenString)
```

One subtest per spec example for the happy path; one per typed error for the failure path. Sentinels with `errors.Is`, struct errors with `errors.As`.

## Phase 3 — printer tests (`printer_test.go`)

Two test shapes are required for every printer rule.

### Direct: AST in, string out

```go
var buf bytes.Buffer
err := Print(&buf, &File{Records: []Record{{Type: "string", Value: "hello"}}})
require.NoError(t, err)
require.Equal(t, `record string = "hello"`+"\n", buf.String())
```

This pins formatting choices — whitespace, quoting, indentation, trailing newline — that the round-trip can't see because both sides agree on them.

### Round-trip: every new printer rule gets one

```go
testCases := []struct{
    name   string
    source string
}{
    {"single_string_record",   `record string = "hello"`},
    {"empty_file",             ``},
    {"two_records",            "record string = \"a\"\nrecord int = 1"},
}

for _, tc := range testCases {
    tc := tc
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()

        first, err := Parse(strings.NewReader(tc.source))
        require.NoError(t, err)

        var buf bytes.Buffer
        require.NoError(t, Print(&buf, first))

        second, err := Parse(&buf)
        require.NoError(t, err)
        require.Equal(t, first, second)
    })
}
```

Round-trip is the cheapest end-to-end correctness check available, and it's what keeps the parser/printer agreement honest. A round-trip mismatch is almost always one of:
- the parser dropping a token (the AST is missing what the source carried),
- the printer omitting punctuation the parser made optional, or
- a struct field that wasn't actually printed (or printed as zero-value when it shouldn't be).

If round-trip passes, the parser and printer agree about the format — even if both happen to be wrong about the spec. Pair round-trip with at least one direct test per printer rule, so the spec itself anchors the surface text.

## Comments and trivia

Many text formats carry comments that aren't grammatically meaningful but must round-trip cleanly when the printer replays them. The standard approach is to attach trivia (leading/trailing comments, blank lines) to the AST node it precedes, and the printer emits them before the node's own output. Tests for trivia are pure round-trip:

```go
{"comment_above_record", "# leading comment\nrecord string = \"hi\""},
```

If the round-trip drops the comment, either the parser didn't capture it or the printer didn't emit it — start at the AST diff.

## Fixture style

Inline source strings beat external fixture files for text tests — the source lives next to the assertion that depends on it, and string literals (with backticks for multi-line) are readable. Reach for a `testdata/` file only when the input is too large to read inline (multi-KB) or when you want to exercise a real-world specimen captured from another implementation.

## When tests fail

- **Wrong `Pos`**: `next()`/`backup()` ordering bug, or a closure captured the position too late. Audit `tokenizer.next()` and the action that yields the token together.
- **`Parse` returns a different `*File` than the round-trip expectation**: the printer is dropping or adding something the parser doesn't reproduce. Diff the two `*File` values, not the surface text.
- **`UnexpectedTokenError` with the wrong `Got` token**: a parser action read past where it should have, or an `expect` call was missing.
- **Race detector fires**: a closure captured a value by reference that's mutated across iterations. Capture by value (`tc := tc` in the loop) or move the state into the action's parameters.
