package toml

import (
	"fmt"
	"io"
)

// printer wraps an io.Writer plus an accumulated err. Every write goes
// through write or writef, which short-circuit when err != nil — actions
// don't need to thread error through every call, and an early write error
// terminates the loop without poisoning the rest of the output.
type printer struct {
	w   io.Writer
	err error
}

// write writes s, accumulating any error. Subsequent writes are no-ops.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef formats and writes, accumulating any error.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step of the printer state machine. Returning nil
// ends the loop; errors flow through pr.err, not through a return value.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s and returns next. The most common ending of any
// printer action.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// printFile is the top-level entry action. Returns nil so empty input
// prints nothing.
func printFile(pr *printer, f *File) printerAction {
	// Stub: the real implementation dispatches over f.Nodes.
	return nil
}

// Print writes the textual form of f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printerAction(printFile); action != nil; {
		action = action(pr, f)
		if pr.err != nil {
			return pr.err
		}
	}
	return pr.err
}
