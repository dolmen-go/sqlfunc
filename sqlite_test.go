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

package sqlfunc_test

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var sqliteDriver = "sqlite3"

// go test -v -run TestSQLiteVersion
// go test -v -run TestSQLiteVersion -tags nomodernc
func TestSQLiteVersion(t *testing.T) {
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		t.Logf("Open: %v", err)
		return
	}
	defer db.Close()

	var version string
	if err = db.QueryRowContext(t.Context(), `SELECT sqlite_version()`).Scan(&version); err != nil {
		t.Logf("sqlite_version(): %v", err)
		return
	}

	t.Logf(
		"SQLite version %s (driver %q, package %q)",
		version, sqliteDriver, reflect.TypeOf(db.Driver()).Elem().PkgPath(),
	)
}
