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

// printFile is the top-level action. It dispatches to a closure-driven
// records iterator when the file has any records, or returns nil for an
// empty file.
func printFile(pr *printer, f *File) printerAction {
	if len(f.Records) == 0 {
		return nil
	}
	return printRecords(f.Records)
}

// printRecords returns a self-recursive printerAction that walks the records
// slice one element per call (closure pattern: index captured in the
// closure, no mutable iterator state on the printer struct).
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

// printRecord writes a single record statement plus its trailing newline.
// Every write goes through pr.write/pr.writef so pr.err short-circuits.
func printRecord(pr *printer, rec Record) {
	pr.writef("record %s %s = ", rec.Type, rec.Key)
	switch rec.Type {
	case "string":
		pr.write(quoteString(rec.Value))
	default:
		pr.write(rec.Value)
	}
	pr.write("\n")
}

// quoteString wraps s in double quotes and escapes the four recognised
// escape characters (\\, \", \n, \t) — matches the tokenizer's decode rules
// so a Parse → Print → Parse round-trip preserves the value.
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
