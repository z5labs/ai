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

// printFile is the top-level action. It hands off to a record-iterating
// closure that walks f.Records in order.
func printFile(pr *printer, f *File) printerAction {
	return printRecords(f.Records)
}

// printRecords returns a closure that prints each record in order, then
// terminates. Index state is captured in the closure (no mutable state on the
// printer struct).
func printRecords(records []Record) printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(records) {
			return nil
		}
		printRecord(pr, records[i])
		i++
		return step
	}
	return step
}

// printRecord writes one record line: `record TYPE KEY = "VALUE"\n`. The
// value is encoded with the inverse of the tokenizer's escape rules.
func printRecord(pr *printer, rec Record) {
	pr.writef("record %s %s = ", rec.Type, rec.Key)
	switch rec.Type {
	case "string":
		pr.write(`"`)
		pr.write(encodeString(rec.Value))
		pr.write(`"`)
	default:
		pr.write(rec.Value)
	}
	pr.write("\n")
}

// encodeString returns the source-form encoding of s: the inverse of the
// tokenizer's string-decoding rules. Backslash, double-quote, newline, and
// tab become their two-character escape sequences; every other rune passes
// through.
func encodeString(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '"':
			sb.WriteString(`\"`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
