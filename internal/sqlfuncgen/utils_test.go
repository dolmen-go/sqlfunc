/*
Copyright 2026 Olivier Mengué

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlfuncgen

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestSuffixFromPkgID(t *testing.T) {
	for _, test := range []*[2]string{
		{"github.com/dolmen-go/sqlfunc", ".go"},
		{"github.com/dolmen-go/sqlfunc [github.com/dolmen-go/sqlfunc.test]", "_test.go"},
		{"github.com/dolmen-go/sqlfunc_test [github.com/dolmen-go/sqlfunc.test]", "_x_test.go"},
	} {
		in, expected := test[0], test[1]
		got := suffixFromPkgID(in)
		if got != expected {
			t.Errorf("%q: got %q, expected %q", in, got, expected)
		}
	}
}

func TestStripNames(t *testing.T) {
	tests := []struct {
		name     string
		input    string // Go type literal or definition
		expected string // Expected String() representation
	}{
		{
			name:     "simple signature",
			input:    "func(x int, y string)",
			expected: "func(int, string)",
		},
		{
			name:     "nested signature",
			input:    "func(f func(x int)) error",
			expected: "func(func(int)) error",
		},
		{
			name:     "interface with methods",
			input:    "interface { F(x int) error }",
			expected: "interface{F(int) error}",
		},
		{
			name:     "recursive interface",
			input:    "interface { F(i interface{ F(x int) }) }",
			expected: "interface{F(interface{F(int)})}",
		},
		{
			name:     "struct with tags and func",
			input:    "struct { Callback func(x int) `json:\"cb\"` }",
			expected: "struct{Callback func(int) \"json:\\\"cb\\\"\"}",
		},
		{
			name:     "pointer to signature",
			input:    "*func(x int)",
			expected: "*func(int)",
		},
		{
			name:     "named type remains intact",
			input:    "error",
			expected: "error",
		},
		{
			name:     "aliased type remains intact",
			input:    "any",
			expected: "any",
		},
		{
			name:     "basic type remains intact",
			input:    "int",
			expected: "int",
		},
		{
			name:     "pointer to basic type remains intact",
			input:    "*int",
			expected: "*int",
		},
		{
			name:     "pointer to Alias type remains intact",
			input:    "*any",
			expected: "*any",
		},
		{
			name:     "pointer to Named type remains intact",
			input:    "*error",
			expected: "*error",
		},
		{
			name:     "underscore",
			input:    "func(_ int, _ int)",
			expected: "func(int, int)",
		},
		{
			name:     "multiple parameters",
			input:    "func(a int, b int)",
			expected: "func(int, int)",
		},
		{
			name:     "return values with names",
			input:    "func() (a int, b int)",
			expected: "func() (int, int)",
		},
		{
			name:     "deeply nested signatures",
			input:    "func() (int, func(x int))",
			expected: "func() (int, func(int))",
		},
		{
			name:     "(error, error) (error, error)",
			input:    "func(_ error, _ error) (_ error, _ error)",
			expected: "func(error, error) (error, error)",
		},
		{
			name:     "channels",
			input:    "func(_ chan func(x int), _ chan<- func(y int), _ <-chan func(z int))",
			expected: "func(chan func(int), chan<- func(int), <-chan func(int))",
		},
		{
			name:     "array, slice",
			input:    "func(_ [0]func(x int), _ [42]func(y int), _ []func(z string))",
			expected: "func([0]func(int), [42]func(int), []func(string))",
		},
		{
			name:     "map",
			input:    "func(_ map[int]int, _ map[int]func(x int), _ map[*func(y string)]int)",
			expected: "func(map[int]int, map[int]func(int), map[*func(string)]int)",
		},
		{
			name:     "struct",
			input:    "func(_ struct{}, _ func(x struct{}), _ func(x struct{X int}), _ func(x struct{int}), _ func(struct{F func(y int); G struct{G func(z string)}}))",
			expected: "func(struct{}, func(struct{}), func(struct{X int}), func(struct{int}), func(struct{F func(int); G struct{G func(string)}}))",
		},
		{
			name:     "interface",
			input:    "struct{a interface{ F(x int) (_ int, err error)}; b interface{ G(y bool); interface {F(x int)}}}",
			expected: "struct{a interface{F(int) (int, error)}; b interface{G(bool); interface{F(int)}}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := parseType(t, tt.input)
			t.Logf("%T", typ)
			stripped := stripNames(typ)

			// 1. Verify the string representation
			got := types.TypeString(stripped, nil)
			if got != tt.expected {
				t.Errorf("\ngot:      %q\nexpected: %q", got, tt.expected)
				if stripped == typ && tt.expected != tt.input {
					t.Errorf("expected new pointer, but got original (logic failed to create new)")
				}
			} else {
				if tt.expected == tt.input && stripped != typ {
					t.Errorf("expected original pointer (CoW), but got a new one")
				}
			}

			// 2. Verify Assignability
			// A value of the original type must be assignable to the new type
			if !types.AssignableTo(typ, stripped) {
				t.Errorf("Assignability check failed: %s is not assignable to %s", typ, stripped)
			}

			// Applying stripNames twice should do nothing: the stringified result should be the same.
			stripped2 := stripNames(stripped)
			got2 := types.TypeString(stripped2, nil)
			if got2 != tt.expected {
				t.Error("stripNames(clean) should return clean.")
				t.Errorf("\ngot:      %q\nexpected: %q", got2, tt.expected)
			} else if stripped2 != stripped {
				t.Log("TODO expected original pointer (CoW), but got a new one")
			}

		})
	}
}

// TestRecursiveCycle specifically tests that we don't stack overflow on cycles
func TestRecursiveCycle(t *testing.T) {
	// type T interface { F(T) }
	// This is hard to parse as a literal, so we construct it via types API
	objT := types.NewTypeName(token.NoPos, nil, "T", nil)
	namedT := types.NewNamed(objT, nil, nil)

	sig := types.NewSignatureType(nil, nil, nil,
		types.NewTuple(types.NewVar(token.NoPos, nil, "x", namedT)),
		nil, false)
	method := types.NewFunc(token.NoPos, nil, "F", sig)

	iface := types.NewInterfaceType([]*types.Func{method}, nil).Complete()
	namedT.SetUnderlying(iface)

	// This should not panic or hang
	stripped := stripNames(namedT)

	if !types.Identical(namedT, stripped) {
		t.Error("Named types should be returned exactly as-is")
	}
}

// Helper to turn a string into a types.Type
func parseType(t *testing.T, exprStr string) types.Type {
	fset := token.NewFileSet()
	// Wrap in a package context to allow parsing interfaces/structs
	src := "package p; var _ " + exprStr
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	_, err = conf.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("failed to typecheck: %v", err)
	}

	// Extract the type from the variable declaration
	var specType ast.Expr
	ast.Inspect(f, func(n ast.Node) bool {
		if vs, ok := n.(*ast.ValueSpec); ok {
			specType = vs.Type
			return false
		}
		return true
	})

	return info.TypeOf(specType)
}

func TestStripNamesWithAlias(t *testing.T) {
	fset := token.NewFileSet()
	// Define an alias to a signature
	src := `
package p
type RealFunc func(x int)
type AliasFunc = RealFunc
var a AliasFunc
`
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("typecheck error: %v", err)
	}

	// Find the AliasFunc type
	var aliasType types.Type
	for id, obj := range info.Defs {
		if id.Name == "a" {
			aliasType = obj.Type()
			break
		}
	}

	if _, ok := aliasType.(*types.Alias); !ok {
		t.Skip("Current Go version does not support *types.Alias or it is disabled")
	}

	stripped := stripNames(aliasType)

	// Verify identity: StripNames should return the Alias itself
	if stripped != aliasType {
		t.Errorf("Expected Alias type to be returned as-is to preserve identity")
	}

	// Verify that the underlying signature, if accessed independently, can be stripped
	underlyingStripped := stripNames(aliasType.Underlying())
	expectedStr := "func(int)"
	if got := types.TypeString(underlyingStripped, nil); got != expectedStr {
		t.Errorf("Underlying stripped string mismatch: got %s, want %s", got, expectedStr)
	}
}

func BenchmarkStripNames(b *testing.B) {
	// Setup a complex "dirty" type: func(x int, y string, f func(z bool))
	dirty := types.NewSignatureType(nil, nil, nil,
		types.NewTuple(
			types.NewVar(0, nil, "x", types.Typ[types.Int]),
			types.NewVar(0, nil, "y", types.Typ[types.String]),
			types.NewVar(0, nil, "f", types.NewSignatureType(nil, nil, nil,
				types.NewTuple(types.NewVar(0, nil, "z", types.Typ[types.Bool])),
				nil, false),
			),
		),
		nil, false)

	// Setup an "already clean" version of the same type
	clean := stripNames(dirty)

	b.Run("Dirty", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = stripNames(dirty)
		}
	})

	b.Run("Clean_CoW", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = stripNames(clean)
		}
	})
}
