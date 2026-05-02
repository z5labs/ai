package kvrx

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

// printFile is the top-level action. It walks f.Statements via a closure
// over the current index — the closure pattern keeps slice iteration off
// the printer struct.
func printFile(pr *printer, f *File) printerAction {
	return printStatements(f.Statements, 0)
}

// printStatements emits each top-level statement. Each statement is followed
// by `\n` (no extra blank line between top-level statements in this scope).
func printStatements(stmts []Statement, indent int) printerAction {
	var step printerAction
	i := 0
	step = func(pr *printer, f *File) printerAction {
		if i >= len(stmts) {
			return nil
		}
		printStatement(pr, stmts[i], indent)
		i++
		return step
	}
	return step
}

// printStatement dispatches on the concrete statement kind and emits its
// rendering followed by a newline. indent is the current indentation level
// (0 at file scope, deeper inside a conditional body).
func printStatement(pr *printer, s Statement, indent int) {
	pr.write(strings.Repeat("  ", indent))
	switch n := s.(type) {
	case *Record:
		printRecord(pr, n)
	case *Conditional:
		printConditional(pr, n, indent)
	default:
		// unknown statement kind — surface as a print error rather than
		// silently dropping it. The hot path uses typed errors elsewhere;
		// for printer-internal "shouldn't happen" cases a fmt.Errorf is
		// acceptable per references/architecture.md (these are not in
		// the tokenizer/parser hot paths assertions reach).
		pr.err = fmt.Errorf("kvrx: printer: unknown statement type %T", s)
	}
}

// printRecord emits `record TYPE KEY = EXPR\n`.
func printRecord(pr *printer, r *Record) {
	pr.writef("record %s %s = ", r.Type, r.Key)
	printExpression(pr, r.Value)
	pr.write("\n")
}

// printConditional emits the if/elif*/else? chain. Body statements are
// indented by indent+1 and terminated with `;`. Closing braces sit on
// their own line at the conditional's indent; elif/else continue on the
// same line as the prior `}`.
func printConditional(pr *printer, c *Conditional, indent int) {
	innerIndent := strings.Repeat("  ", indent+1)
	for i, br := range c.Branches {
		if i == 0 {
			pr.writef("%s ", br.Keyword) // "if "
		} else {
			pr.writef(" %s ", br.Keyword) // " elif " or " else "
		}
		if br.Condition != nil {
			pr.write("(")
			printExpression(pr, br.Condition)
			pr.write(") ")
		}
		pr.write("{\n")
		for _, stmt := range br.Body {
			pr.write(innerIndent)
			printBranchBodyStatement(pr, stmt, indent+1)
		}
		pr.write(strings.Repeat("  ", indent))
		pr.write("}")
	}
	pr.write("\n")
}

// printBranchBodyStatement is like printStatement, but inner statements end
// with `;` (per Block grammar) rather than `\n`. The caller already wrote
// the leading indent.
func printBranchBodyStatement(pr *printer, s Statement, indent int) {
	switch n := s.(type) {
	case *Record:
		pr.writef("record %s %s = ", n.Type, n.Key)
		printExpression(pr, n.Value)
		pr.write(";\n")
	case *Conditional:
		// nested conditional inside a branch: indent its `{ ... }`, but
		// terminate the whole construct with `;` per Block grammar.
		printConditionalInline(pr, n, indent)
		pr.write(";\n")
	default:
		pr.err = fmt.Errorf("kvrx: printer: unknown branch-body statement type %T", s)
	}
}

// printConditionalInline emits a conditional without a trailing newline so
// the caller (printBranchBodyStatement) can add the `;`.
func printConditionalInline(pr *printer, c *Conditional, indent int) {
	innerIndent := strings.Repeat("  ", indent+1)
	for i, br := range c.Branches {
		if i == 0 {
			pr.writef("%s ", br.Keyword)
		} else {
			pr.writef(" %s ", br.Keyword)
		}
		if br.Condition != nil {
			pr.write("(")
			printExpression(pr, br.Condition)
			pr.write(") ")
		}
		pr.write("{\n")
		for _, stmt := range br.Body {
			pr.write(innerIndent)
			printBranchBodyStatement(pr, stmt, indent+1)
		}
		pr.write(strings.Repeat("  ", indent))
		pr.write("}")
	}
}

// printExpression emits the source form of an expression. Supported forms
// match the parser's stub Expression grammar: BoolLiteral, StringLiteral,
// Reference, EqualExpr.
func printExpression(pr *printer, e Expression) {
	switch n := e.(type) {
	case *BoolLiteral:
		if n.Value {
			pr.write("true")
		} else {
			pr.write("false")
		}
	case *StringLiteral:
		pr.write(quoteString(n.Value))
	case *Reference:
		pr.writef("&%s", n.Name)
	case *EqualExpr:
		printExpression(pr, n.Left)
		pr.write(" == ")
		printExpression(pr, n.Right)
	default:
		pr.err = fmt.Errorf("kvrx: printer: unknown expression type %T", e)
	}
}

// quoteString re-encodes s as a single-line double-quoted string, escaping
// the reverse of the tokenizer's escape recognition.
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
		case '\r':
			sb.WriteString(`\r`)
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
