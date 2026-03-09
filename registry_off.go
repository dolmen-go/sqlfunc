//go:build sqlfunc_registry_off && !sqlfunc_registry_on && !sqlfunc_registry_sync

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
)

const registryEnabled = false

func registrySetForEach(typ reflect.Type, f any) {
}

func registryForEach(typ reflect.Type) any {
	return nil
}

func registrySetScan(typ reflect.Type, f reflect.Value) {
}

func registryScan(typ reflect.Type) reflect.Value {
	return reflect.Value{}
}

func registrySetStmt(typ reflect.Type, f func(*sql.Stmt, any)) {
}

func registryStmt(typ reflect.Type) func(*sql.Stmt, any) {
	return nil
}
