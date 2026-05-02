package kvrx

import (
	"fmt"
	"io"
)

// printer formats a *File back to its source form.
type printer struct {
	w   io.Writer
	err error
}

func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step in the printer state machine.
type printerAction func(pr *printer, f *File) printerAction

// printFile walks f.Statements and dispatches per statement type. Each
// statement emits its own trailing newline; no extra blank line is added
// between top-level statements.
func printFile(pr *printer, f *File) printerAction {
	for _, stmt := range f.Statements {
		switch s := stmt.(type) {
		case Record:
			printRecord(pr, s, "")
			pr.write("\n")
		case Conditional:
			printConditional(pr, s, "")
		}
	}
	return nil
}

// printRecord emits `record TYPE KEY = EXPR` (no trailing newline).
func printRecord(pr *printer, r Record, indent string) {
	pr.write(indent)
	pr.write("record ")
	printType(pr, r.Type)
	pr.write(" ")
	pr.write(r.Key)
	pr.write(" = ")
	printExpression(pr, r.Value)
}

func printType(pr *printer, t Type) {
	switch v := t.(type) {
	case NamedType:
		pr.write(v.Name)
	default:
		pr.writef("<unknown-type %T>", t)
	}
}

func printExpression(pr *printer, e Expression) {
	switch v := e.(type) {
	case BoolLiteral:
		if v.Value {
			pr.write("true")
		} else {
			pr.write("false")
		}
	case Reference:
		pr.write("&")
		pr.write(v.Name)
	case BinaryExpr:
		printExpression(pr, v.Left)
		pr.write(" ")
		pr.write(v.Op)
		pr.write(" ")
		printExpression(pr, v.Right)
	default:
		pr.writef("<unknown-expr %T>", e)
	}
}

// printConditional emits the full `if (...) {...} elif (...) {...} else {...}`
// chain. Each branch's body is printed with one tab of indent and a trailing
// `;` per statement.
func printConditional(pr *printer, c Conditional, indent string) {
	for i, br := range c.Branches {
		if i == 0 {
			pr.write(indent)
			pr.write("if (")
			printExpression(pr, br.Cond)
			pr.write(") {\n")
		} else if br.Kind == "elif" {
			pr.write(" elif (")
			printExpression(pr, br.Cond)
			pr.write(") {\n")
		} else {
			pr.write(" else {\n")
		}
		for _, stmt := range br.Body {
			switch s := stmt.(type) {
			case Record:
				printRecord(pr, s, indent+"\t")
				pr.write(";\n")
			case Conditional:
				printConditional(pr, s, indent+"\t")
				pr.write("\n")
			}
		}
		pr.write(indent)
		pr.write("}")
	}
	pr.write("\n")
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
