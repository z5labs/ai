package graphql

import (
	"fmt"
	"io"
)

// printer is the internal state for printing. It wraps an io.Writer and
// accumulates the first write error so action bodies stay clean.
type printer struct {
	w   io.Writer
	err error
}

// write writes s to the underlying writer, short-circuiting if a previous
// write already failed. Once pr.err is set, subsequent writes are no-ops.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef formats according to format and writes the result via pr.write.
// It short-circuits on a prior error just like write.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step of the printer state machine. Errors flow
// through pr.err rather than the action signature, so action bodies stay
// readable.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s and returns next. Use it for the common
// "emit a fixed string then continue" pattern; it composes cleanly with
// closures that capture iteration state.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// printFile is the top-level entry action. The scaffold returns nil so
// empty input produces empty output; the implementer dispatches on
// f.Definitions and chains element-printing actions.
func printFile(pr *printer, f *File) printerAction {
	return nil
}

// Print formats f to w as a GraphQL document. The driver loop checks
// pr.err each iteration so a write error stops the loop and surfaces
// without poisoning the rest of the output.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printerAction(printFile); action != nil; {
		action = action(pr, f)
		if pr.err != nil {
			return pr.err
		}
	}
	return nil
}
