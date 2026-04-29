// Package ini reads and writes the INI text file format.
//
// The package follows the tokenizer / parser / printer pipeline: bytes flow
// through Tokenize to produce an iter.Seq2[Token, error]; the parser pulls
// tokens to build a *File AST; the printer formats a *File back to bytes.
// Each component is a state machine expressed as recursive action functions
// and can be tested in isolation.
package ini
