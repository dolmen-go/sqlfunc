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
)

var _ *sql.DB // Fake var just to have database/sql imported for go doc

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
	if fnPtr == nil {
		panic("fnPtr must be non-nil")
	}
	return anyExec(reflect.TypeFor[Func](), ctx, db, query, reflect.ValueOf(fnPtr))
}

func anyExec(fnType reflect.Type, ctx context.Context, db PrepareConn, query string, fnValue reflect.Value) (close func() error, err error) {
	makeFn := registryExec(fnType)
	withTx := false

	if makeFn == nil {
		if fnType.Kind() != reflect.Func {
			panic("fnPtr must be a pointer to a *func* variable")
		}
		numIn := fnType.NumIn()
		if numIn < 1 || fnType.In(0) != typeContext {
			panic("func first arg must be a context.Context")
		}
		// Optional *sql.Tx as In(1) (if db is not already a *sql.Tx)
		if numIn > 1 && fnType.In(1).Implements(typeTxStmt) {
			withTx = true
		}
		if fnType.NumOut() != 2 || fnType.Out(0) != typeResult || fnType.Out(1) != typeError {
			panic("func must return (sql.Result, error)")
		}
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	var fn reflect.Value
	if makeFn != nil {
		fn = makeFn(stmt)
	} else {
		firstArg := 1
		if withTx {
			firstArg = 2
		}
		fn = reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
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
	}

	fnValue.Elem().Set(fn)

	return stmt.Close, nil
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
	if fnPtr == nil {
		panic("fnPtr must be non-nil")
	}
	return anyQueryRow(reflect.TypeFor[Func](), ctx, db, query, reflect.ValueOf(fnPtr))
}

func anyQueryRow(fnType reflect.Type, ctx context.Context, db PrepareConn, query string, fnValue reflect.Value) (close func() error, err error) {
	makeFn := registryQueryRow(fnType)
	withTx := false

	if makeFn == nil {
		if fnType.Kind() != reflect.Func {
			panic("fnPtr must be a pointer to a *func* variable")
		}
		numIn := fnType.NumIn()
		if numIn < 1 || fnType.In(0) != typeContext {
			panic("func first arg must be a context.Context")
		}
		// Optional *sql.Tx as In(1) (if db is not already a *sql.Tx)
		if numIn > 1 && fnType.In(1).Implements(typeTxStmt) {
			withTx = true
		}
		numOut := fnType.NumOut()
		if numOut < 2 {
			panic("func must return at least one column")
		}
		if fnType.Out(numOut-1) != typeError {
			panic("func must return an error")
		}
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	var fn reflect.Value
	if makeFn != nil {
		fn = makeFn(stmt)
	} else {
		firstArg := 1
		if withTx {
			firstArg = 2
		}
		numOut := fnType.NumOut()
		fn = reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
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
	}

	fnValue.Elem().Set(fn)

	return stmt.Close, nil
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
	if fnPtr == nil {
		panic("fnPtr must be non-nil")
	}
	return anyQuery(reflect.TypeFor[Func](), ctx, db, query, reflect.ValueOf(fnPtr))
}

func anyQuery(fnType reflect.Type, ctx context.Context, db PrepareConn, query string, fnValue reflect.Value) (close func() error, err error) {
	// FIXME add support for *sql.Tx arg

	makeFn := registryQuery(fnType)

	if makeFn == nil {
		if fnType.Kind() != reflect.Func {
			panic("fnPtr must be a pointer to a *func* variable")
		}
		if fnType.NumIn() < 1 || fnType.In(0) != typeContext {
			panic("func first arg must be a context.Context")
		}
		if fnType.NumOut() != 2 || fnType.Out(0) != typeRows || fnType.Out(1) != typeError {
			panic("func must return (*sql.Rows, error)")
		}
	}

	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return func() error { return nil }, err
	}

	var fn reflect.Value
	if makeFn != nil {
		fn = makeFn(stmt)
	} else {
		fn = reflect.MakeFunc(fnType, func(in []reflect.Value) []reflect.Value {
			ctx := in[0].Interface().(context.Context)
			var args []any
			if len(in) > 1 {
				args = make([]any, len(in)-1)
				for i, a := range in[1:] {
					args[i] = a.Interface()
				}
			}
			rows, err := stmt.QueryContext(ctx, args...)
			return []reflect.Value{reflect.ValueOf(&rows).Elem(), reflect.ValueOf(&err).Elem()}
		})
	}

	fnValue.Elem().Set(fn)

	return stmt.Close, nil
}
