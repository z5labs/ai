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
	idx int
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

// printFile is the top-level action. It walks f.Records, emitting each record
// on its own line.
func printFile(pr *printer, f *File) printerAction {
	if pr.idx >= len(f.Records) {
		return nil
	}
	rec := f.Records[pr.idx]
	pr.idx++
	printRecord(pr, rec)
	return printFile
}

// printRecord writes one record in the canonical form
//
//	record TYPE KEY = VALUE\n
func printRecord(pr *printer, r Record) {
	switch r.Type {
	case RecordTypeString:
		pr.writef("record string %s = %s\n", r.Key, quoteString(r.Value))
	default:
		if pr.err == nil {
			pr.err = fmt.Errorf("printer: unknown record type %s", r.Type)
		}
	}
}

// quoteString re-encodes s as a KVR string literal: surrounding double quotes,
// with backslash, double-quote, newline, and tab escaped.
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

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
