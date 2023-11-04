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
type txStmt interface {
	StmtContext(ctx context.Context, stmt *sql.Stmt) *sql.Stmt
}

var (
	typeBool    = reflect.TypeOf(true)
	typeContext = reflect.TypeOf(new(context.Context)).Elem()
	typeResult  = reflect.TypeOf(new(sql.Result)).Elem()
	typeError   = reflect.TypeOf(new(error)).Elem()
	typeScanner = reflect.TypeOf(new(sql.Scanner)).Elem()
	typeRows    = reflect.TypeOf(new(*sql.Rows)).Elem()
	typeTxStmt  = reflect.TypeOf(new(txStmt)).Elem()
)
