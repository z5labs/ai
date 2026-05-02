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

// printFile is the top-level action. It iterates records first, then blocks,
// each on its own line.
func printFile(pr *printer, f *File) printerAction {
	for i := range f.Records {
		printRecord(pr, &f.Records[i], "")
	}
	for i := range f.Blocks {
		printBlock(pr, &f.Blocks[i])
	}
	return nil
}

// printRecord writes a record (with leading comments) followed by a newline.
// indent is prepended to each emitted line (used for records nested in blocks).
func printRecord(pr *printer, r *Record, indent string) {
	for _, c := range r.LeadingComments {
		pr.writef("%s# %s\n", indent, c)
	}
	pr.writef("%srecord %s %s = %s\n", indent, r.Type, r.Key, formatValue(r))
}

// printBlock writes a block declaration with its records indented.
func printBlock(pr *printer, b *Block) {
	for _, c := range b.LeadingComments {
		pr.writef("# %s\n", c)
	}
	if len(b.Records) == 0 {
		pr.writef("block %s {\n}\n", b.Name)
		return
	}
	pr.writef("block %s {\n", b.Name)
	for i := range b.Records {
		r := &b.Records[i]
		// Print leading comments on their own lines (indented), then the
		// record line ending with `;` instead of newline-only.
		for _, c := range r.LeadingComments {
			pr.writef("    # %s\n", c)
		}
		pr.writef("    record %s %s = %s;\n", r.Type, r.Key, formatValue(r))
	}
	pr.write("}\n")
}

// formatValue renders a record's value back into source form, quoting and
// escaping as required.
func formatValue(r *Record) string {
	switch r.Type {
	case RecordTypeString:
		return quoteString(r.Value)
	case RecordTypeNumber:
		return r.Value
	default:
		return r.Value
	}
}

// quoteString applies the inverse of the tokenizer's string-escape rules.
func quoteString(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
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
	sb.WriteByte('"')
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
