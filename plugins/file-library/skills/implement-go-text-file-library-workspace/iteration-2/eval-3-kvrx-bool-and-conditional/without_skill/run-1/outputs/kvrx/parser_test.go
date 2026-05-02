package kvrx

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  *File
	}{
		{
			name:  "empty_input_yields_zero_file",
			input: "",
			want:  &File{},
		},
		{
			name:  "record_bool_true",
			input: "record bool ENABLED = true",
			want: &File{
				Statements: []Statement{
					Record{
						Pos:   Pos{Line: 1, Column: 1},
						Type:  NamedType{Pos: Pos{Line: 1, Column: 8}, Name: "bool"},
						Key:   "ENABLED",
						Value: BoolLiteral{Pos: Pos{Line: 1, Column: 23}, Value: true},
					},
				},
			},
		},
		{
			name:  "record_bool_false",
			input: "record bool DARK = false",
			want: &File{
				Statements: []Statement{
					Record{
						Pos:   Pos{Line: 1, Column: 1},
						Type:  NamedType{Pos: Pos{Line: 1, Column: 8}, Name: "bool"},
						Key:   "DARK",
						Value: BoolLiteral{Pos: Pos{Line: 1, Column: 20}, Value: false},
					},
				},
			},
		},
		{
			name: "if_else_with_bool_reference_condition",
			input: "record bool MODE = true\n" +
				"if (&MODE == true) {\n" +
				"\trecord bool PORT_OPEN = true;\n" +
				"} else {\n" +
				"\trecord bool PORT_OPEN = false;\n" +
				"}",
			want: &File{
				Statements: []Statement{
					Record{
						Pos:   Pos{Line: 1, Column: 1},
						Type:  NamedType{Pos: Pos{Line: 1, Column: 8}, Name: "bool"},
						Key:   "MODE",
						Value: BoolLiteral{Pos: Pos{Line: 1, Column: 20}, Value: true},
					},
					Conditional{
						Pos: Pos{Line: 2, Column: 1},
						Branches: []ConditionalBranch{
							{
								Pos:  Pos{Line: 2, Column: 1},
								Kind: "if",
								Cond: BinaryExpr{
									Pos:   Pos{Line: 2, Column: 5},
									Op:    "==",
									Left:  Reference{Pos: Pos{Line: 2, Column: 5}, Name: "MODE"},
									Right: BoolLiteral{Pos: Pos{Line: 2, Column: 14}, Value: true},
								},
								Body: []Statement{
									Record{
										Pos:   Pos{Line: 3, Column: 2},
										Type:  NamedType{Pos: Pos{Line: 3, Column: 9}, Name: "bool"},
										Key:   "PORT_OPEN",
										Value: BoolLiteral{Pos: Pos{Line: 3, Column: 26}, Value: true},
									},
								},
							},
							{
								Pos:  Pos{Line: 4, Column: 3},
								Kind: "else",
								Body: []Statement{
									Record{
										Pos:   Pos{Line: 5, Column: 2},
										Type:  NamedType{Pos: Pos{Line: 5, Column: 9}, Name: "bool"},
										Key:   "PORT_OPEN",
										Value: BoolLiteral{Pos: Pos{Line: 5, Column: 26}, Value: false},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "if_elif_else_chain_with_bool_literals",
			input: "if (false) {\n" +
				"\trecord bool A = true;\n" +
				"} elif (true) {\n" +
				"\trecord bool A = false;\n" +
				"} else {\n" +
				"\trecord bool A = true;\n" +
				"}",
			want: &File{
				Statements: []Statement{
					Conditional{
						Pos: Pos{Line: 1, Column: 1},
						Branches: []ConditionalBranch{
							{
								Pos:  Pos{Line: 1, Column: 1},
								Kind: "if",
								Cond: BoolLiteral{Pos: Pos{Line: 1, Column: 5}, Value: false},
								Body: []Statement{
									Record{
										Pos:   Pos{Line: 2, Column: 2},
										Type:  NamedType{Pos: Pos{Line: 2, Column: 9}, Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Pos: Pos{Line: 2, Column: 18}, Value: true},
									},
								},
							},
							{
								Pos:  Pos{Line: 3, Column: 3},
								Kind: "elif",
								Cond: BoolLiteral{Pos: Pos{Line: 3, Column: 9}, Value: true},
								Body: []Statement{
									Record{
										Pos:   Pos{Line: 4, Column: 2},
										Type:  NamedType{Pos: Pos{Line: 4, Column: 9}, Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Pos: Pos{Line: 4, Column: 18}, Value: false},
									},
								},
							},
							{
								Pos:  Pos{Line: 5, Column: 3},
								Kind: "else",
								Body: []Statement{
									Record{
										Pos:   Pos{Line: 6, Column: 2},
										Type:  NamedType{Pos: Pos{Line: 6, Column: 9}, Name: "bool"},
										Key:   "A",
										Value: BoolLiteral{Pos: Pos{Line: 6, Column: 18}, Value: true},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(strings.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParser_UndeclaredReferenceInCondition(t *testing.T) {
	t.Parallel()

	input := "if (&MISSING == true) { record bool A = true; }"
	_, err := Parse(strings.NewReader(input))
	require.Error(t, err)
	require.IsType(t, &UndeclaredReferenceError{}, err)
}
