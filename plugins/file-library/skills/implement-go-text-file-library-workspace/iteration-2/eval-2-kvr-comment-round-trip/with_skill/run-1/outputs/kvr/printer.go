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

// quoteString re-encodes a decoded string Value back to its quoted, escaped
// source form. This is the inverse of the tokenizer's escape-decoding: the
// tokenizer maps `\\`, `\"`, `\n`, `\t` into their runes, so we map them
// back here.
func quoteString(s string) string {
	var b strings.Builder
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

// printLeadingComments emits each comment on its own line, prefixed with
// "# ". A comment with empty Value still emits "#\n" (the parser preserves
// empty-content comments by yielding TokenComment with Value="").
func printLeadingComments(pr *printer, comments []string) {
	for _, c := range comments {
		if c == "" {
			pr.write("#\n")
			continue
		}
		pr.writef("# %s\n", c)
	}
}

// printRecord emits a single record. Caller is responsible for any trailing
// punctuation (top-level uses '\n'; block-internal uses ";\n").
func printRecord(pr *printer, rec Record) {
	printLeadingComments(pr, rec.LeadingComments)
	pr.writef("record %s %s = ", rec.Type, rec.Key)
	switch rec.Type {
	case "string":
		pr.write(quoteString(rec.Value))
	default:
		// number (and any future numeric kinds) — Value is already the
		// digit text from the tokenizer.
		pr.write(rec.Value)
	}
}

// printBlock emits a single block, including its inner records each
// terminated by ';'.
func printBlock(pr *printer, blk Block) {
	printLeadingComments(pr, blk.LeadingComments)
	pr.writef("block %s {\n", blk.Name)
	for _, rec := range blk.Records {
		printRecord(pr, rec)
		pr.write(";\n")
	}
	pr.write("}")
}

// printRecords iterates File.Records, emitting each followed by '\n'.
// Uses the closure pattern to carry the index across action invocations.
func printRecords() printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(f.Records) {
			return printBlocks()
		}
		printRecord(pr, f.Records[i])
		pr.write("\n")
		i++
		return step
	}
	return step
}

// printBlocks iterates File.Blocks, emitting each followed by '\n'.
func printBlocks() printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(f.Blocks) {
			return nil
		}
		printBlock(pr, f.Blocks[i])
		pr.write("\n")
		i++
		return step
	}
	return step
}

// printFile is the top-level action. Records first, then Blocks.
func printFile(pr *printer, f *File) printerAction {
	return printRecords()
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
