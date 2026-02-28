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
	"runtime"
	"slices"
	"strings"

	"go/types"
)

// alignLineNum tweaks a [text/template] with a long multiline comment that
// allows to align the source line number with the line number from where this
// function is called.
//
//go:noinline
func alignLineNum(template string) string {
	_, _, line, _ := runtime.Caller(1)
	return `{{/*` + strings.Repeat("\n", line-1) + ` */}}` + template
}

// stripNames exhaustively removes parameter names from UNNAMED signatures.
// It preserves *types.Named to maintain assignability and type identity.
func stripNames(typ types.Type) types.Type {
	return stripNamesRecursive(typ, make(map[types.Type]types.Type))
}

func stripNamesRecursive(typ types.Type, seen map[types.Type]types.Type) types.Type {
	if typ == nil {
		return nil
	}

	// 1. Memoization to handle recursive structures
	if cached, ok := seen[typ]; ok {
		return cached
	}

	switch t := typ.(type) {
	case *types.Signature:
		sig := &types.Signature{}
		seen[typ] = sig

		// Create the new signature
		*sig = *types.NewSignatureType(
			t.Recv(),
			slices.Collect(t.RecvTypeParams().TypeParams()),
			slices.Collect(t.TypeParams().TypeParams()),
			stripNamesTuple(t.Params(), seen),
			stripNamesTuple(t.Results(), seen),
			t.Variadic(),
		)
		return sig

	case *types.Interface:
		iface := &types.Interface{}
		seen[typ] = iface

		methods := make([]*types.Func, t.NumExplicitMethods())
		for i := range t.NumExplicitMethods() {
			m := t.ExplicitMethod(i)
			// Methods are Funcs; their Type() is always a Signature
			newSig := stripNamesRecursive(m.Type(), seen).(*types.Signature)
			methods[i] = types.NewFunc(m.Pos(), m.Pkg(), m.Name(), newSig)
		}

		embeddeds := make([]types.Type, t.NumEmbeddeds())
		for i := range t.NumEmbeddeds() {
			embeddeds[i] = stripNamesRecursive(t.EmbeddedType(i), seen)
		}

		*iface = *types.NewInterfaceType(methods, embeddeds).Complete()
		return iface

	case *types.Pointer:
		return types.NewPointer(stripNamesRecursive(t.Elem(), seen))

	case *types.Slice:
		return types.NewSlice(stripNamesRecursive(t.Elem(), seen))

	case *types.Array:
		return types.NewArray(stripNamesRecursive(t.Elem(), seen), t.Len())

	case *types.Map:
		return types.NewMap(stripNamesRecursive(t.Key(), seen), stripNamesRecursive(t.Elem(), seen))

	case *types.Chan:
		return types.NewChan(t.Dir(), stripNamesRecursive(t.Elem(), seen))

	case *types.Struct:
		fields := make([]*types.Var, t.NumFields())
		tags := make([]string, t.NumFields())

		for i := range t.NumFields() {
			f := t.Field(i)
			// We keep the field name (it's part of the struct identity),
			// but we strip names from the type of the field.
			fields[i] = types.NewVar(f.Pos(), f.Pkg(), f.Name(), stripNamesRecursive(f.Type(), seen))
			// Preserve the tag for identity/assignability
			tags[i] = t.Tag(i)
		}
		return types.NewStruct(fields, tags)

	default: // *types.Named, *types.Alias, *types.Basic
		return t
	}
}

func stripNamesTuple(tup *types.Tuple, seen map[types.Type]types.Type) *types.Tuple {
	if tup == nil {
		return nil
	}

	var vars []*types.Var // Lazy slice, remains nil if no changes

	// From https://go.dev/ref/spec#Function_types
	// "Within a list of parameters or results, the names (IdentifierList) must either all be present or all be absent."
	hasNames := tup.At(0).Name() != ""

	for i := range tup.Len() {
		v := tup.At(i)
		typ := v.Type()
		typStripped := stripNamesRecursive(typ, seen)

		// Check if we need to transform this Var
		if hasNames || typStripped != typ {
			// If this is the FIRST change we've found,
			// we must finally allocate and catch up.
			if vars == nil {
				vars = slices.Collect(tup.Variables())
			}
			vars[i] = types.NewVar(v.Pos(), v.Pkg(), "", typStripped)
		}
	}

	// If vars is still nil, it means the loop finished without finding a single change.
	if vars == nil {
		return tup
	}

	return types.NewTuple(vars...)
}
