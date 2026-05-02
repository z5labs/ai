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

// printFile is the top-level action. It iterates over file.Statements and
// dispatches to the appropriate node printer.
func printFile(pr *printer, f *File) printerAction {
	for _, stmt := range f.Statements {
		switch s := stmt.(type) {
		case Record:
			printLeadingComments(pr, s.LeadingComments)
			printRecord(pr, s)
			pr.write("\n")
		case Block:
			printLeadingComments(pr, s.LeadingComments)
			printBlock(pr, s)
			pr.write("\n")
		default:
			pr.err = fmt.Errorf("printer: unknown statement type %T", stmt)
			return nil
		}
		if pr.err != nil {
			return nil
		}
	}
	return nil
}

func printLeadingComments(pr *printer, comments []string) {
	for _, c := range comments {
		if c == "" {
			pr.write("#\n")
		} else {
			pr.writef("# %s\n", c)
		}
	}
}

func printRecord(pr *printer, rec Record) {
	pr.writef("record %s %s = ", rec.Type, rec.Key)
	switch rec.Type {
	case RecordTypeString:
		pr.write(quoteString(rec.Value))
	case RecordTypeNumber:
		pr.write(rec.Value)
	default:
		pr.err = fmt.Errorf("printer: unknown record type %v", rec.Type)
	}
}

// quoteString wraps s in double quotes, escaping the four sequences the
// tokenizer recognises: \\, \", \n, \t. Other characters pass through
// unchanged.
func quoteString(s string) string {
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

func printBlock(pr *printer, blk Block) {
	pr.writef("block %s {\n", blk.Name)
	for _, rec := range blk.Records {
		for _, c := range rec.LeadingComments {
			if c == "" {
				pr.write("    #\n")
			} else {
				pr.writef("    # %s\n", c)
			}
		}
		pr.write("    ")
		printRecord(pr, rec)
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
