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

package sqlfunc

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/dolmen-go/sqlfunc/internal/registry"
)

var _ *sql.DB // Fake var just to have database/sql imported for go doc

// makeStmtFuncTyped is the common logic of wrappers of prepared statements ([Exec], [QueryRow] and [Query])
// for the generic versions.
//
// The checkSignature and makeFn callback implement the parts specific to Exec/QueryRow/Query
func makeStmtFuncTyped[Func any](
	ctx context.Context,
	db PrepareConn,
	query string,
	fnPtr *Func,
	checkSignature func(reflect.Type),
	makeFn func(reflect.Type) func(*sql.Stmt, any),
) (
	func() error,
	error,
) {
	if fnPtr == nil {
		panic("fnPtr must be non-nil")
	}
	fnType := reflect.TypeFor[Func]()
	if fnType.Kind() != reflect.Func {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
		panic("func first arg must be a context.Context")
	}

	checkSignature(fnType)

	// setFn := registry.Stmt.Get(reflect.TypeFor[Func]())
	setFn := registryStmt(reflect.TypeFor[Func]())

	switch setFn := setFn.(type) {
	case func(*sql.Stmt, *Func):
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return func() (_ error) { return }, err
		}

		setFn(stmt, fnPtr)
		return stmt.Close, nil

	case func(*sql.Stmt, any):
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return func() (_ error) { return }, err
		}

		setFn(stmt, fnPtr)
		return stmt.Close, nil

	case nil:
		return setStmtAny(ctx, db, query, fnPtr, fnType, makeFn)

	default:
		panic(fnType.String() + " not handled")
	}
}

// makeStmtFuncAny is the common logic of wrappers of prepared statements ([Any.Exec], [Any.QueryRow] and [Any.Query]).
//
// The checkSignature and makeFn callback implement the parts specific to Exec/QueryRow/Query
func makeStmtFuncAny(
	ctx context.Context,
	db PrepareConn,
	query string,
	fnPtr any,
	checkSignature func(reflect.Type),
	makeFn func(reflect.Type) func(*sql.Stmt, any),
) (
	func() error,
	error,
) {
	if fnPtr == nil {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	fnPtrValue := reflect.ValueOf(fnPtr)
	if fnPtrValue.Kind() != reflect.Pointer {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	if fnPtrValue.IsNil() {
		panic("fnPtr must be non-nil")
	}

	fnType := fnPtrValue.Type().Elem()
	if fnType.Kind() != reflect.Func {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
		panic("func first arg must be a context.Context")
	}

	// As the registry is shared between Exec/QueryRow/Query, we must check
	// that the user didn't ask for a Exec-style func in a call to Query
	// as that might bring us here.
	// As the signatures do not overlap .
	// TODO benchmark the cost of separate registry vs the presence of this
	// check in all paths: with separate registries we could move this check
	// out of the paths where we found an entry in the registry.
	checkSignature(fnType)

	setFn := registry.Stmt.Get(fnType)

	switch setFn := setFn.(type) {

	case func(*sql.Stmt, any):
		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return func() (_ error) { return }, err
		}

		setFn(stmt, fnPtr)
		return stmt.Close, nil

	case nil:
		return setStmtAny(ctx, db, query, fnPtr, fnType, makeFn)

	default: // func(*sql.Stmt, *Func)

		stmt, err := db.PrepareContext(ctx, query)
		if err != nil {
			return func() (_ error) { return }, err
		}

		// like setFn(stmt, fnPtr)
		_ = reflect.ValueOf(setFn).Call([]reflect.Value{reflect.ValueOf(stmt), reflect.ValueOf(fnPtr)})
		return stmt.Close, nil
	}
}

func setStmtAny(
	ctx context.Context,
	db PrepareConn,
	query string,
	fnPtr any,
	fnType reflect.Type,
	makeFn func(reflect.Type) func(*sql.Stmt, any),
) (
	func() error,
	error,
) {
	setFn := makeFn(fnType)
	// go registry.Stmt.Register(fnType, setFn)

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() (_ error) { return }, err
	}
	setFn(stmt, fnPtr)
	return stmt.Close, nil
}

// Exec prepares an SQL statement and creates a function wrapping [sql.Stmt.ExecContext].
//
// fnPtr is a pointer to a func variable. The function signature tells how it will be called.
//
// The first argument is a [context.Context].
// If a [*sql.Tx] is given as the second argument, the statement will be localized to the transaction (using [sql.Tx.StmtContext]).
// The following arguments will be given as arguments to [sql.Stmt.ExecContext].
//
// The function will return an [sql.Result] and an error.
//
// The returned func 'close' must be called once the statement is not needed anymore.
//
// Example:
//
//	var f func(ctx context.Context, arg1 int64, arg2 string, arg3 sql.NullInt, arg4 *sql.Time) (sql.Result, error)
//	close1, err = sqlfunc.Exec(ctx, db, "SELECT ?, ?, ?, ?", &f)
//	// if err != nil ...
//	defer close1()
//	res, err = f(ctx, 1, "a", sql.NullInt{Valid: false}, time.Now())
//
// Example with transaction:
//
//	var fTx func(ctx context, *sql.Tx, arg1 int64) (sql.Result, error)
//	close2, err = sqlfunc.Exec(ctx, db, "SELECT ?", &fTx)
//	// if err != nil ...
//	defer close2()
//
//	tx, err := db.BeginTxt()
//	// if err != nil ...
//	res, err := fTx(ctx, tx, 123)
//	// if err != nil ...
//	err = tx.Commit()
//	// if err != nil ...
func Exec[Func any](ctx context.Context, db PrepareConn, query string, fnPtr *Func) (close func() error, err error) {
	return makeStmtFuncTyped(ctx, db, query, fnPtr, checkExec, makeExec)
}

func checkExec(fnType reflect.Type) {
	if fnType.NumOut() != 2 || fnType.Out(0) != typeResult || fnType.Out(1) != typeError {
		panic("func must return (sql.Result, error)")
	}
}

func checkQueryRow(fnType reflect.Type) {
	numOut := fnType.NumOut()
	switch numOut {
	case 1:
		if fnType.Out(0) != typeRow {
			break
		}
		fallthrough
	case 0:
		panic("func must return either (*sql.Row) or (values..., error)")
	default:
		switch fnType.Out(0) {
		case typeRow:
			panic("func must return ONLY *sql.Row")
		// Ensure no overlap of signature with sqlfunc.Exec, sqlfunc.Query.
		// This is necessary because of the shared registry.
		case typeResult, typeRows:
			panic("func must return either (*sql.Row) or (values..., error)")
		}
		if fnType.Out(numOut-1) != typeError {
			panic("func must return an error")
		}
	}
}

func checkQuery(fnType reflect.Type) {
	if fnType.NumOut() != 2 || fnType.Out(0) != typeRows || fnType.Out(1) != typeError {
		panic("func must return (*sql.Rows, error)")
	}
}

func makeExec(fnType reflect.Type) func(*sql.Stmt, any) {
	if fnType.IsVariadic() {
		panic("func must not be variadic")
	}
	// Optional *sql.Tx as In(1) (if db is not already a *sql.Tx)
	withTx := fnType.NumIn() > 1 && fnType.In(1).Implements(typeTxStmt)

	// returning just an error (ignoring sql.Result) isn't implemented
	if fnType.NumOut() != 2 || fnType.Out(0) != typeResult || fnType.Out(1) != typeError {
		panic("func must return (sql.Result, error)")
	}

	firstArg := 1
	if withTx {
		firstArg = 2
	}

	return func(stmt *sql.Stmt, fnPtr any) {
		fn := reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
			ctx := in[0].Interface().(context.Context)
			stmtTx := stmt
			if withTx && !in[1].IsNil() {
				stmtTx = in[1].Interface().(txStmt).StmtContext(ctx, stmt)
				defer stmtTx.Close()
			}
			var args []any
			if len(in) > firstArg {
				args = make([]any, len(in)-firstArg)
				for i, a := range in[firstArg:] {
					args[i] = a.Interface()
				}
			}
			r, err := stmtTx.ExecContext(ctx, args...)
			return []reflect.Value{reflect.ValueOf(&r).Elem(), reflect.ValueOf(&err).Elem()}
		})

		reflect.ValueOf(fnPtr).Elem().Set(fn)
	}
}

// QueryRow prepares an SQL statement and creates a function wrapping [sql.Stmt.QueryRowContext] and [sql.Row.Scan].
//
// fnPtr is a pointer to a func variable. The function signature tells how it will be called.
//
// The first argument is a [context.Context].
// If a [*sql.Tx] is given as the second argument, the statement will be localized to the transaction (using [sql.Tx.StmtContext]).
// The following arguments will be given as arguments to [sql.Stmt.QueryRowContext].
//
// The function will return values scanned from the [sql.Row] and an error.
//
// The returned func 'close' must be called once the statement is not needed anymore.
func QueryRow[Func any](ctx context.Context, db PrepareConn, query string, fnPtr *Func) (close func() error, err error) {
	return makeStmtFuncTyped(ctx, db, query, fnPtr, checkQueryRow, makeQueryRow)
}

func makeQueryRow(fnType reflect.Type) func(*sql.Stmt, any) {
	if fnType.IsVariadic() {
		panic("func must not be variadic")
	}

	// Return (*sql.Row) is not yet implemented in the reflect-based version below
	if fnType.NumOut() == 1 {
		panic("func must return at least one column")
	}

	// Optional *sql.Tx as In(1) (if db is not already a *sql.Tx)
	withTx := fnType.NumIn() > 1 && fnType.In(1).Implements(typeTxStmt)

	firstArg := 1
	if withTx {
		firstArg = 2
	}
	numOut := fnType.NumOut()

	return func(stmt *sql.Stmt, fnPtr any) {
		fn := reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
			ctx := in[0].Interface().(context.Context)
			stmtTx := stmt
			if withTx && !in[1].IsNil() {
				stmtTx = in[1].Interface().(txStmt).StmtContext(ctx, stmt)
				defer stmtTx.Close()
			}
			var args []any
			if len(in) > firstArg {
				args = make([]any, len(in)-firstArg)
				for i, a := range in[firstArg:] {
					args[i] = a.Interface()
				}
			}
			out := make([]any, numOut-1)
			outValues := make([]reflect.Value, numOut)
			for i := 0; i < numOut-1; i++ {
				ptr := reflect.New(fnType.Out(i))
				out[i] = ptr.Interface()
				outValues[i] = ptr.Elem()
			}

			err := stmtTx.QueryRowContext(ctx, args...).Scan(out...)
			outValues[numOut-1] = reflect.ValueOf(&err).Elem()
			return outValues
		})

		reflect.ValueOf(fnPtr).Elem().Set(fn)
	}
}

// Query prepares an SQL statement and creates a function wrapping [sql.Stmt.QueryContext].
//
// fnPtr is a pointer to a func variable. The function signature tells how it will be called.
//
// The first argument is a [context.Context].
// If an [*sql.Tx] is given as the second argument, the statement will be localized to the transaction (using [sql.Tx.StmtContext]).
// The following arguments will be given as arguments to [sql.Stmt.QueryRowContext].
//
// The function will return an [*sql.Rows] and an error.
//
// The returned func 'close' must be called once the statement is not needed anymore.
func Query[Func any](ctx context.Context, db PrepareConn, query string, fnPtr *Func) (close func() error, err error) {
	return makeStmtFuncTyped(ctx, db, query, fnPtr, checkQuery, makeQuery)
}

func makeQuery(fnType reflect.Type) func(*sql.Stmt, any) {
	if fnType.IsVariadic() {
		panic("func must not be variadic")
	}

	// Optional *sql.Tx as In(1) (if db is not already a *sql.Tx)
	withTx := fnType.NumIn() > 1 && fnType.In(1).Implements(typeTxStmt)

	firstArg := 1
	if withTx {
		firstArg = 2
	}

	return func(stmt *sql.Stmt, fnPtr any) {
		fn := reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
			ctx := in[0].Interface().(context.Context)
			stmtTx := stmt
			if withTx && !in[1].IsNil() {
				stmtTx = in[1].Interface().(txStmt).StmtContext(ctx, stmt)
				defer stmtTx.Close()
			}
			var args []any
			if len(in) > firstArg {
				args = make([]any, len(in)-firstArg)
				for i, a := range in[firstArg:] {
					args[i] = a.Interface()
				}
			}
			rows, err := stmtTx.QueryContext(ctx, args...)
			return []reflect.Value{reflect.ValueOf(&rows).Elem(), reflect.ValueOf(&err).Elem()}
		})

		reflect.ValueOf(fnPtr).Elem().Set(fn)
	}
}
