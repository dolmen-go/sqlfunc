/*
Copyright 2022 Olivier Mengué

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

package sqlfunc

import (
	"database/sql"
	"reflect"
)

// Scan allows to define a function that will scan one row from an *sql.Rows.
//
// The signature of the function defines how the column values are retrieved into variables.
// Two styles are available:
//   - as pointer variables (like sql.Rows.Scan()): func (rows *sql.Rows, pval1 *int, pval2 *string) error
//   - as returned values (implies copies): func (rows *sql.Rows) (val1 int, val2 string, err error)
func Scan(fnPtr interface{}) {
	vPtr := reflect.ValueOf(fnPtr)
	if vPtr.Type().Kind() != reflect.Ptr {
		panic("fnPtr must be a *pointer* to a func variable")
	}
	if vPtr.IsNil() {
		panic("fnPtr must be non-nil")
	}
	fnType := reflect.TypeOf(fnPtr).Elem()
	if fnType.Kind() != reflect.Func {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	numIn := fnType.NumIn()
	if numIn < 1 || fnType.In(0) != typeRows {
		panic("func first arg must be an *sql.Rows")
	}
	numOut := fnType.NumOut()
	if numOut < 1 || fnType.Out(numOut-1) != typeError {
		panic("func must return error as last value")
	}
	if numIn == 1 {
		if numOut == 1 {
			panic("func must either take scanners as arguments or return values")
		}
		// TODO check that for each Out type:
		// - either pointer to element type either implements sql.Scanner
		// - or element type is a concrete type (kind not Func, Interface) that can be copied
	} else {
		if numOut != 1 {
			panic("func must either take scanners as arguments or return values")
		}
		// TODO check that each In:
		// - either is an sql.Out
		// - or implements sql.Scanner
		// - or is an anonymous pointer to a concrete type
	}

	var fn func(in []reflect.Value) []reflect.Value
	if numIn > 1 {
		scanners := make([]interface{}, numIn-1)
		out := make([]reflect.Value, 1)
		fn = func(in []reflect.Value) []reflect.Value {
			// in[0] is *sql.Rows, scanners follow...
			for i := range in[1:] {
				scanners[i] = in[i+1].Interface()
			}
			err := in[0].Interface().(*sql.Rows).Scan(scanners...)
			out[0] = reflect.ValueOf(&err).Elem()
			return out
		}
	} else { // numOut > 1
		scanners := make([]interface{}, numOut-1)
		out := make([]reflect.Value, numOut)
		fn = func(in []reflect.Value) []reflect.Value {
			for i := range scanners {
				ptr := reflect.New(fnType.Out(i))
				scanners[i] = ptr.Interface()
				out[i] = ptr.Elem()
			}
			err := in[0].Interface().(*sql.Rows).Scan(scanners...)
			out[numOut-1] = reflect.ValueOf(&err).Elem()
			return out
		}
	}
	vPtr.Elem().Set(reflect.MakeFunc(fnType, fn))
}

// ForEach iterates an *sql.Rows, scans the values of the row and calls the given callback function with the values.
//
// The callback receives the scanned columns values as arguments and may return an error to stop iterating.
//
// rows are closed before returning.
func ForEach(rows *sql.Rows, callback interface{}) error {
	fnType := reflect.TypeOf(callback)
	f := registry.ForEach.Get(fnType)
	if f == nil {

		if fnType.Kind() != reflect.Func {
			panic("callback must be a func")
		}
		numIn := fnType.NumIn()
		if numIn == 0 {
			panic("callback must accept at least one argument")
		}
		withError := fnType.NumOut() > 0
		if withError {
			if fnType.NumOut() != 1 || fnType.Out(0) != typeError {
				panic("callback may only return an error")
			}
		}
		inTypes := make([]reflect.Type, numIn, numIn)
		for i := 0; i < numIn; i++ {
			inTypes[i] = fnType.In(i)
		}

		f = (&runForEach{
			inTypes:   inTypes,
			withError: withError,
		}).run
		// Register in the background
		go registry.ForEach.Register(callback, f)
	}
	return f(rows, callback)
}

type runForEach struct {
	inTypes   []reflect.Type
	withError bool
}

func (r *runForEach) run(rows *sql.Rows, callback interface{}) (err error) {
	defer func() {
		e := rows.Close()
		if err == nil {
			err = e // TODO wrap
		}
	}()

	fn := reflect.ValueOf(callback)
	if fn.IsNil() {
		panic("callback must be non-nil")
	}

	numIn := len(r.inTypes)
	scanners := make([]interface{}, numIn)
	fnArgs := make([]reflect.Value, numIn)

	for rows.Next() {
		for i := 0; i < numIn; i++ {
			ptr := reflect.New(r.inTypes[i])
			scanners[i] = ptr.Interface()
			fnArgs[i] = ptr.Elem()
		}

		err = rows.Scan(scanners...)
		if err != nil {
			// TODO wrap err
			return
		}
		if r.withError {
			var isError bool
			if err, isError = fn.Call(fnArgs)[0].Interface().(error); isError {
				return // user error: don't wrap
			}
		} else {
			fn.Call(fnArgs)
		}
	}

	err = rows.Err() // TODO wrap
	return
}
