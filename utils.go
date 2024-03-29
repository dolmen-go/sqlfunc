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
	"context"
	"database/sql"
	"reflect"
)

// PrepareConn is a subset of [*database/sql.DB], [*database/sql.Conn] or [*database/sql.Tx].
type PrepareConn interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// txStmt is a subset of [*database/sql.Tx].
type txStmt = interface {
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

var (
	// Concrete types
	typeBool = reflect.TypeOf(true)
	typeRows = reflect.TypeOf((*sql.Rows)(nil))

	// Interfaces
	typeContext = reflect.TypeOf([]context.Context(nil)).Elem()
	typeResult  = reflect.TypeOf([]sql.Result(nil)).Elem()
	typeError   = reflect.TypeOf([]error(nil)).Elem()
	typeScanner = reflect.TypeOf([]sql.Scanner(nil)).Elem()
	typeTxStmt  = reflect.TypeOf([]txStmt(nil)).Elem()
)
