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

package registry

import (
	"database/sql"
	"reflect"
)

var (
	ForEach registryOf[FuncForEach]
	Scan    registryOf[FuncScan]
	// Stmt is the shared registry for Exec, QueryRow, Query.
	// This is possible because the shapes of the return types never overlap:
	// - Exec: returns (sql.Result, error) or (error)
	// - Query: returns (*sql.Rows, error)
	// - QueryRow: returns (*sql.Row) or (values..., error)
	Stmt registryOf[FuncStmt]
)

type (
	FuncForEach = any
	FuncScan    = reflect.Value

	FuncStmt     = func(*sql.Stmt, any)
	FuncExec     = FuncStmt
	FuncQueryRow = FuncStmt
	FuncQuery    = FuncStmt
)
