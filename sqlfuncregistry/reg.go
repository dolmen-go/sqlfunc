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
