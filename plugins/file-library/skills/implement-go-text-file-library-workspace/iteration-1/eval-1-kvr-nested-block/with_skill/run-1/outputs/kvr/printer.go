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

// printFile is the top-level action. It first prints all top-level records
// using a closure that walks the slice index by index, then chains to the
// blocks closure.
func printFile(pr *printer, f *File) printerAction {
	return printRecordsThenBlocks(f)
}

// printRecordsThenBlocks returns a closure that iterates f.Records, then
// chains to the block-printing closure.
func printRecordsThenBlocks(f *File) printerAction {
	i := 0
	var step printerAction
	step = func(pr *printer, f *File) printerAction {
		if i >= len(f.Records) {
			return printBlocks(f)
		}
		printRecord(pr, f.Records[i], "")
		i++
		return step
	}
	return step
}

// printBlocks returns a closure that iterates f.Blocks.
func printBlocks(f *File) printerAction {
	i := 0
	var step printerAction
	step = func(pr *printer, f *File) printerAction {
		if i >= len(f.Blocks) {
			return nil
		}
		printBlock(pr, f.Blocks[i])
		i++
		return step
	}
	return step
}

// printRecord writes one record line. indent is "" at the top level and a
// fixed indent inside a block. The trailing terminator (newline at top level,
// ";\n" inside a block) is supplied by the caller.
func printRecord(pr *printer, rec Record, indent string) {
	for _, c := range rec.LeadingComments {
		pr.writef("%s# %s\n", indent, c)
	}
	pr.writef("%srecord %s %s = %s\n", indent, rec.Type, rec.Key, formatValue(rec.Type, rec.Value))
}

// printBlockRecord is the inside-a-block variant — same shape as printRecord
// but with a `;` separator before the newline.
func printBlockRecord(pr *printer, rec Record, indent string) {
	for _, c := range rec.LeadingComments {
		pr.writef("%s# %s\n", indent, c)
	}
	pr.writef("%srecord %s %s = %s;\n", indent, rec.Type, rec.Key, formatValue(rec.Type, rec.Value))
}

// printBlock writes one block's full output: leading comments, header line,
// inner records (each indented and `;`-terminated), and the closing brace.
func printBlock(pr *printer, blk Block) {
	for _, c := range blk.LeadingComments {
		pr.writef("# %s\n", c)
	}
	pr.writef("block %s {\n", blk.Name)
	for _, rec := range blk.Records {
		printBlockRecord(pr, rec, "    ")
	}
	pr.write("}\n")
}

// formatValue produces the source-form value text for a record. Strings are
// quoted with the same escapes the tokenizer recognises; numbers are emitted
// verbatim.
func formatValue(typ, value string) string {
	switch typ {
	case "string":
		return quoteString(value)
	default:
		return value
	}
}

// quoteString wraps s in double quotes and escapes the runes the tokenizer
// would otherwise reject or interpret.
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
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
	b.WriteByte('"')
	return b.String()
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}
