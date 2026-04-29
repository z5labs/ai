package graphql

import (
	"fmt"
	"io"
)

// printer is the printer state. It accumulates the first error it observes
// in err so action functions can short-circuit on subsequent calls.
type printer struct {
	w   io.Writer
	err error
}

// write emits s to the underlying writer. If a prior write recorded an error,
// it is a no-op.
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef formats a string with fmt.Fprintf and emits it. If a prior write
// recorded an error, it is a no-op.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	_, pr.err = fmt.Fprintf(pr.w, format, args...)
}

// printerAction is one step of the printer state machine. Returning nil
// signals that printing is complete.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s and returns next. It is a convenience for actions that
// need to emit a literal string and then transition.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// Print formats f to w using the printer state machine. The first I/O error
// terminates printing and is returned.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil; {
		action = action(pr, f)
		if pr.err != nil {
			return pr.err
		}
	}
	return nil
}

// printFile is the entry-point action of the printer state machine. The
// scaffold completes immediately; real grammar will iterate over the file's
// definitions and emit them.
func printFile(pr *printer, f *File) printerAction {
	return nil
}
