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
	return stripNamesRecursive(typ, stripNamesCache{})
}

type stripNamesCache map[types.Type]types.Type

/*
func (seen stripNamesCache) stripNamesSeqTypes(seq iter.Seq[types.Type]) iter.Seq[types.Type] {
	return func(yield func(types.Type) bool) {
		for typ := range seq {
			yield(stripNamesRecursive(typ, seen))
		}
	}
}

func (seen stripNamesCache) stripNamesSliceTypes(n int, seqFunc func() iter.Seq[types.Type]) []types.Type {
	if n == 0 {
		return nil
	}
	return slices.AppendSeq(make([]types.Type, 0, n), seen.stripNamesSeqTypes(seqFunc()))
}
*/

// stripNamesAny saves a reference to a future copy of typ in the cache, then does the copy,
// allowing to resolve self-references.
func stripNamesAny[TPtr interface {
	*T
	types.Type
}, T any](seen stripNamesCache, typ TPtr, stripNames func(typ TPtr) TPtr) types.Type {
	newT := TPtr(new(T))
	seen[typ] = newT
	stripped := stripNames(typ)
	if stripped == typ { // If strip returned identity, newT hasn't been used so we can just drop it.
		seen[typ] = typ
		return typ
	}
	*newT = *stripped
	return newT
}

// stripNamesElem strips names in the element type of Pointer, Slice, Array, Chan.
// Identity is returned if the element type is already clean.
func stripNamesElem[TPtr interface {
	*T
	types.Type
	Elem() types.Type
}, T any](seen stripNamesCache, typ TPtr, build func(elemType types.Type) TPtr) types.Type {
	return stripNamesAny(seen, typ, func(typ TPtr) TPtr {
		elem := typ.Elem()
		elemNew := stripNamesRecursive(elem, seen)
		if elemNew == elem {
			return typ
		}
		return build(elemNew)
	})
}

func stripNamesRecursive(typ types.Type, seen stripNamesCache) types.Type {
	if typ == nil {
		return nil
	}

	// Memoization to handle recursive structures
	//
	// Note that Go types can only be self-referencing via Named or Alias,
	// but we don't need to explore them for our purpose.
	if cached, ok := seen[typ]; ok {
		return cached
	}

	switch t := typ.(type) {

	// FIXME for *type.Alias, and the alias scope is not at package scope, we should unalias it

	case *types.Signature:
		return stripNamesAny(seen, t, func(t *types.Signature) *types.Signature {
			params := stripNamesTuple(t.Params(), seen)
			results := stripNamesTuple(t.Results(), seen)

			if params == t.Params() && results == t.Results() {
				return t
			}

			return types.NewSignatureType(
				t.Recv(),
				slices.Collect(t.RecvTypeParams().TypeParams()),
				slices.Collect(t.TypeParams().TypeParams()),
				params,
				results,
				t.Variadic(),
			)
		})

	case *types.Interface:
		return stripNamesAny(seen, t, func(t *types.Interface) *types.Interface {
			var methods []*types.Func
			for i := range t.NumExplicitMethods() {
				m := t.ExplicitMethod(i)
				sig := m.Type()
				newSig := stripNamesRecursive(sig, seen).(*types.Signature)
				if newSig != sig {
					if methods == nil {
						methods = slices.Collect(t.ExplicitMethods())
					}
					methods[i] = types.NewFunc(m.Pos(), m.Pkg(), m.Name(), newSig)
				}
			}

			// embeddeds := seen.stripNamesSliceTypes(t.NumEmbeddeds(), t.EmbeddedTypes)
			var embeddeds []types.Type
			for i := range t.NumEmbeddeds() {
				em := t.EmbeddedType(i)
				stripped := stripNamesRecursive(em, seen)
				if stripped != em {
					if embeddeds == nil {
						embeddeds = slices.Collect(t.EmbeddedTypes())
					}
					embeddeds[i] = stripped
				}
			}

			if methods == nil && embeddeds == nil {
				return t
			}

			if methods == nil {
				methods = slices.Collect(t.ExplicitMethods())
			}

			return types.NewInterfaceType(methods, embeddeds).Complete()
		})

	case *types.Pointer:
		return stripNamesElem(seen, t, types.NewPointer)
	case *types.Slice:
		return stripNamesElem(seen, t, types.NewSlice)
	case *types.Array:
		return stripNamesElem(seen, t, func(elem types.Type) *types.Array {
			return types.NewArray(elem, t.Len())
		})
	case *types.Chan:
		return stripNamesElem(seen, t, func(elem types.Type) *types.Chan {
			return types.NewChan(t.Dir(), elem)
		})
	case *types.Map:
		return stripNamesAny(seen, t, func(t *types.Map) *types.Map {
			key := stripNamesRecursive(t.Key(), seen)
			elem := stripNamesRecursive(t.Elem(), seen)
			if key == t.Key() && elem == t.Elem() {
				return t
			}
			return types.NewMap(key, elem)
		})

	case *types.Struct:
		return stripNamesAny(seen, t, func(t *types.Struct) *types.Struct {
			numFields := t.NumFields()
			if numFields == 0 {
				return t
			}

			var fields []*types.Var

			for i := range numFields {
				f := t.Field(i)
				newT := stripNamesRecursive(f.Type(), seen)
				if newT != f.Type() {
					if fields == nil {
						fields = slices.Collect(t.Fields())
					}
					fields[i] = types.NewField(f.Pos(), f.Pkg(), f.Name(), newT, f.Embedded())
				}
			}
			if fields == nil {
				return t
			}

			tags := make([]string, t.NumFields())
			for i := range numFields {
				tags[i] = t.Tag(i)
			}
			return types.NewStruct(fields, tags)
		})

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
			vars[i] = types.NewParam(v.Pos(), v.Pkg(), "", typStripped)
			// vars[i].SetKind(v.Kind()) // FIXME go 1.25
		}
	}

	// If vars is still nil, it means the loop finished without finding a single change.
	if vars == nil {
		return tup
	}

	return types.NewTuple(vars...)
}
