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
	"testing"
)

func scanInt(rows *sql.Rows) (v0 int, err error) {
	err = rows.Scan(&v0)
	return
}

func TestScan(t *testing.T) {
	// This test exists just to get coverage of registry_off.go and reach 100%.
	// The real tests are in sqlfunc package.

	typ := reflect.TypeOf(scanInt)
	if Scan.Get(typ) != nil {
		t.Errorf("registry is expected to be empty")
	}
	v := reflect.ValueOf(scanInt)
	Scan.Register(typ, v)

	v2 := Scan.Get(typ)
	if v2 == nil {
		t.Logf("registry is disabled (sqlfunc_registry_off)")
	} else if reflect.TypeOf(v2) != reflect.TypeOf(v) {
		t.Errorf("not same type")
	}
}
