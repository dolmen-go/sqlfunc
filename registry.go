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
