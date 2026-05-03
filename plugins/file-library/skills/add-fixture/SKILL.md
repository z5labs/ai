---
name: add-fixture
description: Add a real-world file as a golden test fixture under a Go file library's `testdata/` directory and wire up a round-trip test against it. Use whenever the user wants to "add a fixture", "add a golden file", "harden the parser with this sample", "drop this real-world file in as a test", or any phrasing that pairs an input file with an existing file library package — even when they don't say "fixture" or "golden" explicitly. Works for both text packages (`tokenizer.go`/`parser.go`/`printer.go`, round-trip is `parse → print → parse → AST equality`) and binary packages (`types.go`/`decoder.go`/`encoder.go`, round-trip is `decode → encode → byte-equal`). Skip when the user wants to scaffold a new package (use `new-go-text-file-library` or `new-go-binary-file-library`), implement spec features (use `implement-go-text-file-library` or `implement-go-binary-file-library`), or audit coverage (use `review-file-library`).
---

You take a real-world file in a target format, place it under the package's `testdata/` directory with a sensible name, and wire up a table-driven round-trip test that reads from `testdata/` and asserts the pipeline preserves the file's meaning. Round-trip is the cheapest end-to-end correctness check available — every fixture you add becomes a regression guard against the next implementation change.

The skill never modifies tokenizer/parser/printer source or types/decoder/encoder source, never edits `SPEC.md`, and never changes existing test cases. It only **adds** — a new file under `testdata/`, and either a new entry in an existing `TestRoundTripFromTestdata` table or a new test function if that table doesn't exist yet.

## Inputs

- **Source file** (required) — path to the real-world file to add as a fixture (e.g. `~/Downloads/sample.kvr`). Source: user prompt. Must exist and be a regular readable file. The file's bytes are copied verbatim — no normalization, no re-encoding, no trimming. A fixture that's been edited in transit is no longer a real-world input.
- **Package directory** (required) — path to the Go file library package (e.g. `pkg/kvr`). Source: user prompt. Must contain either the text triple (`tokenizer.go`/`parser.go`/`printer.go`) or the binary triple (`types.go`/`decoder.go`/`encoder.go`). If neither, refuse and point at the scaffold skills (`new-go-text-file-library`, `new-go-binary-file-library`) — there's nothing to round-trip against.
- **Fixture name** (optional) — the filename to use under `testdata/`. Source: user prompt; defaults to the source file's basename. Keep the original extension; users grep `testdata/*.kvr` and expect that to work.

## Outputs

- **Copied fixture** at `<package>/testdata/<fixture-name>`, byte-identical to the source. Use `Read` on the source and `Write` on the destination so byte contents are preserved exactly.
- **Test wiring** in the round-trip test file (`printer_test.go` for text packages, `encoder_test.go` for binary packages). Either a new entry appended to `TestRoundTripFromTestdata`'s `testCases` table, or — if that test does not yet exist — a fresh `TestRoundTripFromTestdata` function added beside the existing tests, containing the new fixture as its only case.
- **Side effects** (run from `<package>/` after wiring; this repo has no root `go.mod`, so package tests run from inside the package):
  - `(cd <package> && go build ./...)` — verifies the test wiring compiles. A compile failure here is a skill bug; fix it before reporting success.
  - `(cd <package> && go test -race ./...)` — runs the suite. The new round-trip case **may legitimately fail** if the package's parser/printer or decoder/encoder is incomplete relative to the fixture — that's the entire point of a golden test. Report the failure clearly to the user (which test, what error) but treat the skill itself as successful: the fixture is on disk, the test is wired, and the failure is a real signal about implementation gaps, not a skill bug.

## Before adding the fixture

1. **Validate inputs** before touching the filesystem so a half-applied change can't ship:
   - Source file exists and is a regular file (not a directory or symlink dangling at a missing target).
   - Package directory exists and contains either the text triple or the binary triple. If neither, stop and tell the user the directory is not a recognized file-library shape — point them at `new-go-text-file-library` / `new-go-binary-file-library`.
   - If both triples coexist somehow, stop and ask the user which pipeline to wire against; this skill targets one shape per run.
   - Resolve the fixture name (user-provided or source basename). The name must be a single path segment — reject any input containing `/`, `\`, or starting with `.` (no `..` traversal, no hidden files masquerading as fixtures).
   - Refuse if `<package>/testdata/<fixture-name>` already exists. Tell the user to either choose a different name or remove the existing fixture first. Overwriting silently would mask the case where two fixtures legitimately need similar names.
2. **Read the package's `CLAUDE.md`** if present — confirms test style (`t.Parallel()`, `require`, table-driven). If the file documents a deviation, mirror it rather than the defaults below.
3. **Check whether `TestRoundTripFromTestdata` already exists.** Grep the round-trip test file (`printer_test.go` for text, `encoder_test.go` for binary). The result drives the wiring step:
   - **Function exists** → append a new entry to its `testCases` table.
   - **Function does not exist** → add the function in full (template below) with the new fixture as its only case.

## Wiring the round-trip test

### Test case naming

The fixture's filename becomes the test case's `name` field, slugified to a Go-friendly identifier: lowercase, alphanumeric, `_` for any other character, collapse runs of `_` to one. Example: `Q4-sales FINAL_v2.kvr` → `q4_sales_final_v2_kvr`. The slug is only the test case name; the file on disk keeps its original filename. Mismatched names (slug ≠ filename) are fine and expected — Go test names follow Go identifier rules, fixture filenames don't have to.

### Text packages — `printer_test.go`

Round-trip shape: `Parse(file) → Print → Parse(buf) → require.Equal(*File, *File)`. AST equality, **not** byte equality — printers are allowed to re-format whitespace, reorder normalized fields, etc. The first `*File` is the ground truth; the second confirms the printer's output is itself parseable to the same AST. If you assert byte-equality here, every cosmetic printer change becomes a fixture churn.

If `TestRoundTripFromTestdata` does not exist, add this function to `printer_test.go`, mirroring the package's import style (testify is `github.com/stretchr/testify/require`):

```go
func TestRoundTripFromTestdata(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name    string
        fixture string
    }{
        {name: "<slug>", fixture: "<filename>"},
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            data, err := os.ReadFile(filepath.Join("testdata", tc.fixture))
            require.NoError(t, err)

            first, err := Parse(bytes.NewReader(data))
            require.NoError(t, err)

            var buf bytes.Buffer
            require.NoError(t, Print(&buf, first))

            second, err := Parse(&buf)
            require.NoError(t, err)
            require.Equal(t, first, second)
        })
    }
}
```

Add `bytes`, `os`, and `path/filepath` to the file's import block if any are missing.

If `TestRoundTripFromTestdata` already exists, find the `testCases` slice literal and append `{name: "<slug>", fixture: "<filename>"},` as a new entry before the slice's closing `}`. Do not touch any other case. Do not reorder.

### Binary packages — `encoder_test.go`

Round-trip shape: `Decode(file) → Encode → require.Equal(originalBytes, buf.Bytes())`. Byte-equality, **not** AST equality — binary formats are typically byte-stable, and a divergence is real evidence that the encoder is dropping or reordering bytes. If the format genuinely has multiple valid encodings of the same logical value (some checksums, some compressed streams), the user can edit the assertion after the fact and add a comment explaining the exception; default to byte-equality.

If `TestRoundTripFromTestdata` does not exist, add this function to `encoder_test.go`:

```go
func TestRoundTripFromTestdata(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        name    string
        fixture string
    }{
        {name: "<slug>", fixture: "<filename>"},
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            data, err := os.ReadFile(filepath.Join("testdata", tc.fixture))
            require.NoError(t, err)

            f, err := Decode(bytes.NewReader(data))
            require.NoError(t, err)

            var buf bytes.Buffer
            require.NoError(t, Encode(&buf, f))
            require.Equal(t, data, buf.Bytes())
        })
    }
}
```

Add `bytes`, `os`, and `path/filepath` to the file's import block if any are missing.

If `TestRoundTripFromTestdata` already exists, append to its `testCases` table the same way as the text case.

## After adding the fixture

1. `(cd <package> && go build ./...)` — must succeed. A compile failure here is the skill's responsibility — usually a missing import. Fix and rebuild before continuing.
2. `(cd <package> && go test -race ./...)` — capture the result.
3. **Report to the user** in three lines:
   - The fixture's destination path and size.
   - Where the test entry was wired (file + test name).
   - The test result — pass, or the specific failing case and its error. If the new round-trip case fails, frame it as expected information about implementation gaps (e.g., "the parser doesn't yet handle `<construct from the fixture>`"), not as a skill failure. Suggest `implement-go-text-file-library` / `implement-go-binary-file-library` as the next step if the failure looks implementation-driven.

## Why this shape

A fixture is a contract: "the package must always handle this real-world input." The contract is enforced by the round-trip test, which is why the file and the test are added together — a fixture without a test is just a file taking up space. Round-trip beats hand-written `expected` AST snippets because the truth is the *file itself*, not a brittle expected value that drifts every time the AST shape changes. Text uses AST equality (printers normalize) and binary uses byte equality (encoders shouldn't drift); mixing those up turns the test into either a cosmetic-change tripwire (text → bytes) or a permission slip for encoder bugs (binary → AST).

`testdata/` is the conventional Go location — `go test` ignores it for compilation but `os.ReadFile("testdata/...")` works because `go test` runs with `cwd = <package>/`. No special handling needed; the table just names the file.
