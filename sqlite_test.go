package sqlfunc_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var sqliteDriver = "sqlite3"

// go test -v -run TestSQLiteVersion
// go test -v -run TestSQLiteVersion -tags nomodernc
func TestSQLiteVersion(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		t.Logf("Open: %v", err)
		return
	}
	defer db.Close()

	var version string
	if err = db.QueryRowContext(ctx, `SELECT sqlite_version()`).Scan(&version); err != nil {
		t.Logf("sqlite_version(): %v", err)
		return
	}

	t.Logf(
		"SQLite version %s (driver %q, package %q)",
		version, sqliteDriver, reflect.TypeOf(db.Driver()).Elem().PkgPath(),
	)
}
