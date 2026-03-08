//go:build sqlfunc_registry_on || !sqlfunc_registry_off

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
	"database/sql"
	"reflect"

	"github.com/dolmen-go/sqlfunc/internal/registry"
)

func registrySetForEach(typ reflect.Type, f any) {
	registry.ForEach.Register(typ, f)
}

func registryForEach(typ reflect.Type) any {
	return registry.ForEach.Get(typ)
}

func registrySetScan(typ reflect.Type, f reflect.Value) {
	registry.Scan.Register(typ, f)
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
