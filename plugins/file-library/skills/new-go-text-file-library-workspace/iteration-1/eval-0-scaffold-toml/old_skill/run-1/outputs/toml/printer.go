// Copyright (c) 2026 z5labs
//
// Licensed under the MIT License (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT

package toml

import (
	"fmt"
	"io"
)

// printer wraps an output writer and tracks the first error encountered
// during printing. Once err is non-nil all subsequent writes are no-ops.
type printer struct {
	w   io.Writer
	err error
}

// write writes s to the underlying writer if no prior error has been
// recorded. The first error is captured and exposed via [printer.err].
func (pr *printer) write(s string) {
	if pr.err != nil {
		return
	}
	_, pr.err = io.WriteString(pr.w, s)
}

// writef is a Printf-style wrapper around write.
func (pr *printer) writef(format string, args ...any) {
	if pr.err != nil {
		return
	}
	pr.write(fmt.Sprintf(format, args...))
}

// printerAction is one step of the printer state machine. Returning nil
// terminates the loop.
type printerAction func(pr *printer, f *File) printerAction

// writeThen writes s to the printer and then returns next as the follow-up
// action. It is a convenience for chaining writes with the next state.
func writeThen(s string, next printerAction) printerAction {
	return func(pr *printer, f *File) printerAction {
		pr.write(s)
		return next
	}
}

// Print writes f to w in canonical TOML form. Print stops at the first
// underlying write error and returns it.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	action := printStart
	for action != nil {
		if pr.err != nil {
			return pr.err
		}
		action = action(pr, f)
	}
	return pr.err
}

// printStart is the entry point of the printer state machine. The scaffold
// returns immediately; real implementations will dispatch on the AST.
func printStart(pr *printer, f *File) printerAction {
	_ = pr
	_ = f
	return nil
}
