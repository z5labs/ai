package ini

import (
	"fmt"
	"io"
)

// printer is the formatting state shared across actions. It wraps an
// io.Writer and accumulates the first write error in err so action bodies
// don't need to thread error through every call.
type printer struct {
	w   io.Writer
	err error
}

// write writes s to the underlying writer, recording the first error in
// pr.err. Subsequent calls are no-ops once an error has been recorded.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef is the printf-style sibling of write. It formats with fmt.Sprintf
// and short-circuits on a previously recorded error.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step in the printer state machine. Errors flow
// through pr.err; the action signature has no error return so individual
// actions stay focused on what to write next.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s and returns next so "emit a literal then continue" is
// a one-liner at every action call site.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// Print formats f to w. It runs the printer action loop, surfacing the
// first write error encountered (if any).
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printStart; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}

// printStart is the top-level entry action. The implementer dispatches on
// the AST shape here. The scaffold returns nil so an empty file prints as
// the empty string.
func printStart(pr *printer, f *File) printerAction {
	_ = pr
	_ = f
	return nil
}
