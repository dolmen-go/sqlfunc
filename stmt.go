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
	"context"
	"reflect"
)

func Exec(ctx context.Context, db PrepareConn, query string, fnPtr interface{}) (close func() error, err error) {
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
	if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
		panic("func first arg must be a context.Context")
	}
	if fnType.NumOut() != 2 || fnType.Out(0) != typeResult || fnType.Out(1) != typeError {
		panic("func must return (sql.Result, error)")
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	fn := func(in []reflect.Value) []reflect.Value {
		ctx := in[0].Interface().(context.Context)
		var args []interface{}
		if len(in) > 1 {
			args = make([]interface{}, len(in)-1)
			for i, a := range in[1:] {
				args[i] = a.Interface()
			}
		}
		r, err := stmt.ExecContext(ctx, args...)
		return []reflect.Value{reflect.ValueOf(&r).Elem(), reflect.ValueOf(&err).Elem()}
	}

	vPtr.Elem().Set(reflect.MakeFunc(fnType, fn))

	return stmt.Close, nil
}

func QueryRow(ctx context.Context, db PrepareConn, query string, fnPtr interface{}) (close func() error, err error) {
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
	if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
		panic("func first arg must be a context.Context")
	}
	numOut := fnType.NumOut()
	if numOut < 2 {
		panic("func must return at least one column")
	}
	if fnType.Out(numOut-1) != typeError {
		panic("func must return an error")
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	fn := func(in []reflect.Value) []reflect.Value {
		ctx := in[0].Interface().(context.Context)
		var args []interface{}
		if len(in) > 1 {
			args = make([]interface{}, len(in)-1)
			for i, a := range in[1:] {
				args[i] = a.Interface()
			}
		}
		out := make([]interface{}, numOut-1)
		outValues := make([]reflect.Value, numOut)
		for i := 0; i < numOut-1; i++ {
			ptr := reflect.New(fnType.Out(i))
			out[i] = ptr.Interface()
			outValues[i] = ptr.Elem()
		}

		err := stmt.QueryRowContext(ctx, args...).Scan(out...)
		outValues[numOut-1] = reflect.ValueOf(&err).Elem()
		return outValues
	}

	vPtr.Elem().Set(reflect.MakeFunc(fnType, fn))

	return stmt.Close, nil
}

func Query(ctx context.Context, db PrepareConn, query string, fnPtr interface{}) (close func() error, err error) {
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
	if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
		panic("func first arg must be a context.Context")
	}
	if fnType.NumOut() != 2 || fnType.Out(0) != typeRows || fnType.Out(1) != typeError {
		panic("func must return (*sql.Rows, error)")
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	fn := func(in []reflect.Value) []reflect.Value {
		ctx := in[0].Interface().(context.Context)
		var args []interface{}
		if len(in) > 1 {
			args = make([]interface{}, len(in)-1)
			for i, a := range in[1:] {
				args[i] = a.Interface()
			}
		}
		rows, err := stmt.QueryContext(ctx, args...)
		return []reflect.Value{reflect.ValueOf(&rows).Elem(), reflect.ValueOf(&err).Elem()}
	}

	vPtr.Elem().Set(reflect.MakeFunc(fnType, fn))

	return stmt.Close, nil
}
