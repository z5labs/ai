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

// printFile drives top-level printing by walking the records slice via a
// closure that carries the current index.
func printFile(pr *printer, f *File) printerAction {
	return printRecordAt(0)
}

// printRecordAt returns an action that prints the record at index i (if any)
// and advances to i+1.
func printRecordAt(i int) printerAction {
	return func(pr *printer, f *File) printerAction {
		if i >= len(f.Records) {
			return nil
		}
		printRecord(pr, f.Records[i])
		return printRecordAt(i + 1)
	}
}

// printRecord writes one record. The output form is:
//
//	record TYPE KEY = "VALUE"\n
//
// where VALUE is encoded with `\\`, `\"`, `\n`, `\t` escapes so the result
// round-trips through the tokenizer.
func printRecord(pr *printer, r Record) {
	pr.writef("record %s %s = \"%s\"\n", r.Type, r.Key, encodeStringLiteral(r.Value))
}

// encodeStringLiteral re-encodes the recognised escape sequences in the
// printer-side encoding. The four pairs mirror the tokenizer: \\ \" \n \t.
func encodeStringLiteral(s string) string {
	var b strings.Builder
	b.Grow(len(s))
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
