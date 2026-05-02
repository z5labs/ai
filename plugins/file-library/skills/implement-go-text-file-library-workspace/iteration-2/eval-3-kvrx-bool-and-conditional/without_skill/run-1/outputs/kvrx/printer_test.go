package kvrx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		file *File
		want string
	}{
		{
			name: "empty_file_prints_empty_string",
			file: &File{},
			want: "",
		},
		{
			name: "record_bool_true",
			file: &File{
				Statements: []Statement{
					Record{
						Type:  NamedType{Name: "bool"},
						Key:   "ENABLED",
						Value: BoolLiteral{Value: true},
					},
				},
			},
			want: "record bool ENABLED = true\n",
		},
		{
			name: "record_bool_false",
			file: &File{
				Statements: []Statement{
					Record{
						Type:  NamedType{Name: "dark"},
						Key:   "DARK",
						Value: BoolLiteral{Value: false},
					},
				},
			},
			want: "record dark DARK = false\n",
		},
		{
			name: "if_else_with_bool_reference_comparison",
			file: &File{
				Statements: []Statement{
					Record{
						Type:  NamedType{Name: "bool"},
						Key:   "MODE",
						Value: BoolLiteral{Value: true},
					},
					Conditional{
						Branches: []ConditionalBranch{
							{
								Kind: "if",
								Cond: BinaryExpr{
									Op:    "==",
									Left:  Reference{Name: "MODE"},
									Right: BoolLiteral{Value: true},
								},
								Body: []Statement{
									Record{
										Type:  NamedType{Name: "bool"},
										Key:   "PORT_OPEN",
										Value: BoolLiteral{Value: true},
									},
								},
							},
							{
								Kind: "else",
								Body: []Statement{
									Record{
										Type:  NamedType{Name: "bool"},
										Key:   "PORT_OPEN",
										Value: BoolLiteral{Value: false},
									},
								},
							},
						},
					},
				},
			},
			want: "record bool MODE = true\n" +
				"if (&MODE == true) {\n" +
				"\trecord bool PORT_OPEN = true;\n" +
				"} else {\n" +
				"\trecord bool PORT_OPEN = false;\n" +
				"}\n",
		},
		{
			name: "if_elif_else_chain_with_bool_literals",
			file: &File{
				Statements: []Statement{
					Conditional{
						Branches: []ConditionalBranch{
							{
								Kind: "if",
								Cond: BoolLiteral{Value: false},
								Body: []Statement{
									Record{
										Type:  NamedType{Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Value: true},
									},
								},
							},
							{
								Kind: "elif",
								Cond: BoolLiteral{Value: true},
								Body: []Statement{
									Record{
										Type:  NamedType{Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Value: false},
									},
								},
							},
							{
								Kind: "else",
								Body: []Statement{
									Record{
										Type:  NamedType{Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Value: true},
									},
								},
							},
						},
					},
				},
			},
			want: "if (false) {\n" +
				"\trecord bool A = true;\n" +
				"} elif (true) {\n" +
				"\trecord bool A = false;\n" +
				"} else {\n" +
				"\trecord bool A = true;\n" +
				"}\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Print(&buf, tc.file)
			require.NoError(t, err)
			require.Equal(t, tc.want, buf.String())
		})
	}
}

// stripPositions returns a copy of f with every AST node's Pos zeroed out.
// Round-trip tests cannot assert positional equality because the printer
// doesn't preserve column-exact spacing.
func stripPositions(f *File) *File {
	out := &File{}
	for _, stmt := range f.Statements {
		out.Statements = append(out.Statements, stripStatement(stmt))
	}
	return out
}

func stripStatement(s Statement) Statement {
	switch v := s.(type) {
	case Record:
		v.Pos = Pos{}
		v.Type = stripType(v.Type)
		v.Value = stripExpression(v.Value)
		return v
	case Conditional:
		v.Pos = Pos{}
		var brs []ConditionalBranch
		for _, br := range v.Branches {
			br.Pos = Pos{}
			if br.Cond != nil {
				br.Cond = stripExpression(br.Cond)
			}
			var body []Statement
			for _, inner := range br.Body {
				body = append(body, stripStatement(inner))
			}
			br.Body = body
			brs = append(brs, br)
		}
		v.Branches = brs
		return v
	}
	return s
}

func stripType(t Type) Type {
	switch v := t.(type) {
	case NamedType:
		v.Pos = Pos{}
		return v
	}
	return t
}

func stripExpression(e Expression) Expression {
	switch v := e.(type) {
	case BoolLiteral:
		v.Pos = Pos{}
		return v
	case Reference:
		v.Pos = Pos{}
		return v
	case BinaryExpr:
		v.Pos = Pos{}
		v.Left = stripExpression(v.Left)
		v.Right = stripExpression(v.Right)
		return v
	}
	return e
}

func TestPrinterRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "empty_source_round_trips",
			source: "",
		},
		{
			name:   "record_bool_true_round_trips",
			source: "record bool ENABLED = true",
		},
		{
			name:   "record_bool_false_round_trips",
			source: "record bool DARK = false",
		},
		{
			name: "if_else_with_bool_reference_comparison_round_trips",
			source: "record bool MODE = true\n" +
				"if (&MODE == true) {\n" +
				"\trecord bool PORT_OPEN = true;\n" +
				"} else {\n" +
				"\trecord bool PORT_OPEN = false;\n" +
				"}",
		},
		{
			name: "if_elif_else_chain_round_trips",
			source: "if (false) {\n" +
				"\trecord bool A = true;\n" +
				"} elif (true) {\n" +
				"\trecord bool A = false;\n" +
				"} else {\n" +
				"\trecord bool A = true;\n" +
				"}",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, err := Parse(strings.NewReader(tc.source))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, Print(&buf, first))

			second, err := Parse(&buf)
			require.NoError(t, err)
			require.Equal(t, stripPositions(first), stripPositions(second))
		})
	}
}
