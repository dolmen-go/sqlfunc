package sqlfunc_test

import (
	"context"
	"database/sql"

	"github.com/dolmen-go/sqlfunc"
)

// Triggers for attempt at code generation for calls to sqlfunc.Any methods
func _(ctx context.Context, db *sql.DB) {
	// Unique signature for Exec via Any
	var f1 func(context.Context, int, int, int) (sql.Result, error)
	_, _ = sqlfunc.Any.Exec(ctx, db, "INSERT ...", &f1)

	// Unique signature for QueryRow via Any
	var f2 func(context.Context, string) (int, string, error)
	_, _ = sqlfunc.Any.QueryRow(ctx, db, "SELECT ...", &f2)

	// Unique signature for Query via Any
	var f3 func(context.Context, float64) (*sql.Rows, error)
	_, _ = sqlfunc.Any.Query(ctx, db, "SELECT ...", &f3)

	// Unique signature for ForEach via Any
	_ = sqlfunc.Any.ForEach(nil, func(a bool, b string) {})

	// Unique signature for ForEach via Any, with indirection
	var f4 func(a bool, b string, c float64)
	_ = sqlfunc.Any.ForEach(nil, f4)

	// Unique signature for Scan via Any
	var f5 func(*sql.Rows, *bool, *string) error
	sqlfunc.Any.Scan(&f5)

	// Unique signature for Scan via Any
	var f6 func(*sql.Rows, *bool, *string, *float64) error
	f6indirect := &f6
	sqlfunc.Any.Scan(f6indirect)
}
