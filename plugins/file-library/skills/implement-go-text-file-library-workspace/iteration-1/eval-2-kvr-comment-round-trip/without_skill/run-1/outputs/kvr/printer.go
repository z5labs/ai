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

// printFile is the top-level action. It dispatches on the AST nodes the file
// contains in order: each statement becomes a single line followed by a
// newline, with leading comments printed first.
func printFile(pr *printer, f *File) printerAction {
	return printFileStatementAt(0)(pr, f)
}

// printFileStatementAt prints the i-th statement and recursively schedules the
// next. When i is past the end, returns nil.
func printFileStatementAt(i int) printerAction {
	return func(pr *printer, f *File) printerAction {
		if i >= len(f.Statements) {
			return nil
		}
		switch s := f.Statements[i].(type) {
		case Record:
			printLeadingComments(pr, s.LeadingComments, "")
			pr.writef("record %s %s = %s\n", s.Type, s.Key, formatValue(s.Type, s.Value))
		case Block:
			printLeadingComments(pr, s.LeadingComments, "")
			pr.writef("block %s {\n", s.Name)
			for _, r := range s.Records {
				printLeadingComments(pr, r.LeadingComments, "    ")
				pr.writef("    record %s %s = %s;\n", r.Type, r.Key, formatValue(r.Type, r.Value))
			}
			pr.write("}\n")
		}
		return printFileStatementAt(i + 1)
	}
}

// printLeadingComments emits each comment on its own line, indented by indent.
func printLeadingComments(pr *printer, comments []string, indent string) {
	for _, c := range comments {
		if c == "" {
			pr.writef("%s#\n", indent)
			continue
		}
		pr.writef("%s# %s\n", indent, c)
	}
}

// formatValue renders a record's value back to source form. Strings get
// re-quoted with the four recognised escapes; numbers pass through verbatim.
func formatValue(typ, value string) string {
	switch typ {
	case "string":
		return quoteString(value)
	case "number":
		return value
	}
	return value
}

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
