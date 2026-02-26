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
	// This is possible because the shapes of the return types never overlap.
	Stmt registryOf[FuncStmt]
)

type (
	FuncForEach = func(*sql.Rows, any) error
	FuncScan    = reflect.Value

	FuncStmt     = func(stmt *sql.Stmt) reflect.Value // Exec, QueryRow, Query
	FuncExec     = FuncStmt
	FuncQueryRow = FuncStmt
	FuncQuery    = FuncStmt
)
