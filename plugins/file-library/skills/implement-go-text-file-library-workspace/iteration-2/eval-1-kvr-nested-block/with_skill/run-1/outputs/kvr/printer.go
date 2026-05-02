package kvr

import (
	"fmt"
	"io"
	"strings"
)

// printer formats a *File back to its source form. Errors accumulate in pr.err
// and short-circuit any further writes — keeps action bodies clean.
type printer struct {
	w   io.Writer
	err error
}

// write writes s and stores any error in pr.err. Subsequent calls become
// no-ops once an error is set.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef formats and writes; same short-circuit semantics as write.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step in the printer state machine. Returning nil ends.
// Errors flow through pr.err, not the return value.
type printerAction func(pr *printer, f *File) printerAction

// printFile dispatches across the file's top-level Records and Blocks.
// Records come first (matching the canonical SPEC examples), then blocks.
func printFile(pr *printer, f *File) printerAction {
	return printRecords(f.Records, printBlocks(f.Blocks, nil))
}

// printRecords returns a closure that emits each top-level record on its own
// line, then continues with next.
func printRecords(records []Record, next printerAction) printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(records) {
			return next
		}
		printTopLevelRecord(pr, records[i])
		i++
		return step
	}
	return step
}

// printBlocks returns a closure that emits each block, then continues with
// next.
func printBlocks(blocks []Block, next printerAction) printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(blocks) {
			return next
		}
		printBlock(pr, blocks[i])
		i++
		return step
	}
	return step
}

// printTopLevelRecord emits a record line with no trailing `;` (top-level
// records are separated by newlines, not semicolons).
func printTopLevelRecord(pr *printer, rec Record) {
	pr.writef("record %s %s = %s\n", rec.Type, rec.Key, formatValue(rec))
}

// printBlock emits a block. The body is emitted with 4-space indentation; an
// empty block prints `block NAME {\n}\n` (open brace on the header line,
// close brace on its own line).
func printBlock(pr *printer, blk Block) {
	pr.writef("block %s {\n", blk.Name)
	for _, rec := range blk.Records {
		pr.writef("    record %s %s = %s;\n", rec.Type, rec.Key, formatValue(rec))
	}
	pr.write("}\n")
}

// formatValue renders a Record's Value as it should appear after `=`. For
// strings, the value is re-quoted with the format's recognised escapes; for
// numbers, the digits are emitted bare.
func formatValue(rec Record) string {
	if rec.Type == "string" {
		return quoteString(rec.Value)
	}
	return rec.Value
}

// quoteString returns s wrapped in `"`, with `\\`, `"`, `\n`, and `\t`
// rewritten as escape sequences so the result round-trips through the
// tokenizer.
func quoteString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
