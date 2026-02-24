package sqlfunc

import (
	"database/sql"
	"reflect"

	"github.com/dolmen-go/sqlfunc/internal/registry"
)

func registryForEach(typ reflect.Type) func(*sql.Rows, any) error {
	return registry.ForEach.Get(typ)
}

func registryScan(typ reflect.Type) reflect.Value {
	return registry.Scan.Get(typ)
}

func registryExec(typ reflect.Type) func(*sql.Stmt) reflect.Value {
	return registry.Stmt.Get(typ)
}

func registryQueryRow(typ reflect.Type) func(*sql.Stmt) reflect.Value {
	return registry.Stmt.Get(typ)
}

func registryQuery(typ reflect.Type) func(*sql.Stmt) reflect.Value {
	return registry.Stmt.Get(typ)
}
