package ini

import (
	"fmt"
	"io"
)

// printer wraps an io.Writer and accumulates the first write error in err.
// All write helpers short-circuit when err is non-nil so action bodies stay
// clean of error plumbing.
type printer struct {
	w   io.Writer
	err error
}

// write writes s to the underlying writer, recording the first error.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef writes a formatted string to the underlying writer, recording the
// first error.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step of the printer state machine. Returning nil ends
// iteration. Errors flow through pr.err, not the return value.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s and returns next as the action to run after the write.
// Used for the common "emit some text, then continue" pattern.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// Print writes the textual representation of f to w. It returns the first
// write error encountered, or nil on success.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}

// printFile is the top-level entry action. The implementer wires up
// dispatch here (one closure per node, advance via captured index) and uses
// the inner action loop pattern for any nested structure (see CLAUDE.md).
func printFile(pr *printer, f *File) printerAction {
	return nil
}
