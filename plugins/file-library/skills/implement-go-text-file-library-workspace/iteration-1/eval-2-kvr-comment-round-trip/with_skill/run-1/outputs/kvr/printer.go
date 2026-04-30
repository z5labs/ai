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

// printFile is the top-level action. It iterates File.Statements via a
// closure that captures the current index, prints each statement (with its
// leading comments), then returns nil when all statements are exhausted.
func printFile(pr *printer, f *File) printerAction {
	return printStatements(f.Statements)
}

// printStatements returns a closure-based action that prints each top-level
// statement in turn.
func printStatements(stmts []Type) printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(stmts) {
			return nil
		}
		stmt := stmts[i]
		i++
		switch v := stmt.(type) {
		case *Record:
			printLeadingComments(pr, v.LeadingComments, "")
			printRecord(pr, *v, "")
			pr.write("\n")
		case *Block:
			printLeadingComments(pr, v.LeadingComments, "")
			printBlock(pr, *v)
		default:
			// unknown AST node — ignore but record an error so callers see it
			if pr.err == nil {
				pr.err = fmt.Errorf("kvr: cannot print AST node of type %T", stmt)
			}
		}
		return step
	}
	return step
}

// printLeadingComments writes one `# comment\n` line per comment, prefixed
// with indent.
func printLeadingComments(pr *printer, comments []string, indent string) {
	for _, c := range comments {
		if c == "" {
			pr.writef("%s#\n", indent)
			continue
		}
		pr.writef("%s# %s\n", indent, c)
	}
}

// printRecord writes one record line WITHOUT a trailing newline. The caller
// adds whatever statement terminator (newline at top level, `;\n` inside a
// block) the surrounding context demands.
func printRecord(pr *printer, r Record, indent string) {
	pr.write(indent)
	pr.writef("record %s %s = ", r.Type, r.Key)
	switch r.ValueKind {
	case TokenString:
		pr.write(`"`)
		pr.write(encodeString(r.Value))
		pr.write(`"`)
	case TokenNumber:
		pr.write(r.Value)
	default:
		// Fall back to string-like quoting if ValueKind was never set —
		// keeps direct AST construction usable in tests.
		if r.Type == "number" {
			pr.write(r.Value)
		} else {
			pr.write(`"`)
			pr.write(encodeString(r.Value))
			pr.write(`"`)
		}
	}
}

// encodeString returns the value with `\` and `"` doubled, and `\n` and `\t`
// rewritten as the two-character escapes. Other runes pass through.
func encodeString(s string) string {
	// Use %q's escaping rules but keep them aligned with the tokenizer's
	// supported escapes: only \\, \", \n, \t. Avoid Go's broader set.
	var b strings.Builder
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

// printBlock writes one block, including its inner records (each with their
// own leading comments and a `;` terminator), and a trailing newline.
func printBlock(pr *printer, b Block) {
	pr.writef("block %s {\n", b.Name)
	for _, rec := range b.Records {
		printLeadingComments(pr, rec.LeadingComments, "    ")
		printRecord(pr, rec, "    ")
		pr.write(";\n")
	}
	pr.write("}\n")
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
