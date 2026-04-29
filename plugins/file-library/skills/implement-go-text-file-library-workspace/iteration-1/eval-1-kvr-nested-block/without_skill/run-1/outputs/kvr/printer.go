package kvr

import (
	"fmt"
	"io"
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

// printFile is the top-level dispatch action. It prints all top-level records
// followed by all blocks, separated by newlines.
func printFile(pr *printer, f *File) printerAction {
	for i := range f.Records {
		printRecord(pr, &f.Records[i], "")
		pr.write("\n")
	}
	for i := range f.Blocks {
		printBlock(pr, &f.Blocks[i])
		pr.write("\n")
	}
	return nil
}

// printRecord writes a record statement, including any leading comments. The
// indent string is prepended to every line written, so callers can nest a
// record inside a block.
func printRecord(pr *printer, r *Record, indent string) {
	for _, c := range r.LeadingComments {
		pr.writef("%s# %s\n", indent, c)
	}
	pr.writef("%srecord %s %s = ", indent, r.Type, r.Key)
	switch v := r.Value.(type) {
	case StringValue:
		pr.write(encodeStringLiteral(v.V))
	case NumberValue:
		pr.write(v.V)
	default:
		// Unknown value types should not happen for parser-produced ASTs,
		// but stay defensive: emit a placeholder rather than panicking.
		pr.writef("%v", v)
	}
}

// encodeStringLiteral wraps s in double quotes and escapes the four runes
// the KVR tokenizer recognises (\\, \", \n, \t). All other runes are
// emitted verbatim — KVR strings only support these four escapes, so any
// extra escaping would round-trip into an InvalidEscapeError.
func encodeStringLiteral(s string) string {
	var b []byte
	b = append(b, '"')
	for _, r := range s {
		switch r {
		case '\\':
			b = append(b, '\\', '\\')
		case '"':
			b = append(b, '\\', '"')
		case '\n':
			b = append(b, '\\', 'n')
		case '\t':
			b = append(b, '\\', 't')
		default:
			b = append(b, []byte(string(r))...)
		}
	}
	b = append(b, '"')
	return string(b)
}

// printBlock writes a block statement and its inner records. Inner records
// are indented one tab.
func printBlock(pr *printer, b *Block) {
	for _, c := range b.LeadingComments {
		pr.writef("# %s\n", c)
	}
	pr.writef("block %s {\n", b.Name)
	for i := range b.Records {
		printRecord(pr, &b.Records[i], "\t")
		pr.write(";\n")
	}
	pr.write("}")
}

// Print formats f to w.
func Print(w io.Writer, f *File) error {
	pr := &printer{w: w}
	for action := printFile; action != nil && pr.err == nil; {
		action = action(pr, f)
	}
	return pr.err
}

