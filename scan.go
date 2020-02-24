/*
Copyright 2020 Olivier Mengué

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

// ForEach iterates an *sql.Rows, scan the values of the row and calls `callback` with the values.
//
// `callback` receives the scanned column values as arguments and may return an error to stop iterating.
func ForEach(rows *sql.Rows, callback interface{}) (err error) {
	defer rows.Close()

	fnType := reflect.TypeOf(callback)
	if fnType.Kind() != reflect.Func {
		panic("callback must be a func")
	}
	numIn := fnType.NumIn()
	if numIn == 0 {
		panic("callback must accept at least one argument")
	}
	if fnType.NumOut() != 1 || fnType.Out(0) != typeError {
		panic("callback must return error")
	}
	fn := reflect.ValueOf(callback)
	if fn.IsNil() {
		panic("callback must be non-nil")
	}

	scanners := make([]interface{}, numIn)
	fnArgs := make([]reflect.Value, numIn)

	for rows.Next() {
		for i := 0; i < numIn; i++ {
			ptr := reflect.New(fnType.In(i))
			scanners[i] = ptr.Interface()
			fnArgs[i] = ptr.Elem()
		}

		err = rows.Scan(scanners...)
		if err != nil {
			return err // TODO wrap
		}
		if err, isError := fn.Call(fnArgs)[0].Interface().(error); isError {
			return err // user error: don't wrap
		}
	}

	return rows.Err() // TODO wrap
}
