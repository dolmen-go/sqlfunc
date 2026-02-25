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

// Package sqlfuncregistry provides the private API for sqlfunc-gen generated code.
package sqlfuncregistry

import (
	"database/sql"
	"reflect"

	"github.com/dolmen-go/sqlfunc/internal/registry"
)

func ForEach[Func any](f func(*sql.Rows, any) error) {
	registry.ForEach.Register(reflect.TypeFor[Func](), f)
}

func Scan[Func any](f reflect.Value) {
	registry.Scan.Register(reflect.TypeFor[Func](), f)
}

func Exec[Func any](f func(*sql.Stmt) reflect.Value) {
	registry.Stmt.Register(reflect.TypeFor[Func](), f)
}

func QueryRow[Func any](f func(*sql.Stmt) reflect.Value) {
	registry.Stmt.Register(reflect.TypeFor[Func](), f)
}

func Query[Func any](f func(*sql.Stmt) reflect.Value) {
	registry.Stmt.Register(reflect.TypeFor[Func](), f)
}
