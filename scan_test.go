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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleForEach() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, ``+
		`SELECT 1`+
		` UNION ALL`+
		` SELECT 2`)
	if err != nil {
		log.Printf("Query: %v", err)
		return
	}

	err = sqlfunc.ForEach(rows, func(n int) {
		fmt.Println(n)
	})
	if err != nil {
		log.Printf("ScanRows: %v", err)
		return
	}

	fmt.Println("Done.")

	// Output:
	// 1
	// 2
	// Done.
}

func ExampleForEach_returnBool() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, ``+
		`SELECT 1`+
		` UNION ALL`+
		` SELECT 2`+
		` UNION ALL`+
		` SELECT 3`)
	if err != nil {
		log.Printf("Query: %v", err)
		return
	}

	err = sqlfunc.ForEach(rows, func(n int) bool {
		fmt.Println(n)
		return n < 2 // Stop iterating on n == 2
	})
	if err != nil {
		log.Printf("ScanRows: %v", err)
		return
	}

	fmt.Println("Done.")

	// Output:
	// 1
	// 2
	// Done.
}

func ExampleForEach_returnError() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, ``+
		`SELECT 1`+
		` UNION ALL`+
		` SELECT 2`+
		` UNION ALL`+
		` SELECT 3`)
	if err != nil {
		log.Printf("Query: %v", err)
		return
	}

	err = sqlfunc.ForEach(rows, func(n int) error {
		fmt.Println(n)
		if n == 2 {
			return io.EOF
		}
		return nil
	})
	if err != nil && !errors.Is(err, io.EOF) {
		log.Printf("ScanRows: %v", err)
		return
	}

	fmt.Println("Done.")

	// Output:
	// 1
	// 2
	// Done.
}

func TestForEachMulti(t *testing.T) {
	testForEachMulti := func(t *testing.T) {
		ctx := t.Context()
		// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
		db, err := sql.Open(sqliteDriver, ":memory:")
		if err != nil {
			t.Fatalf("Open: %v", err)
			return
		}
		defer db.Close()

		start := time.Now()
		for i := 0; i < 10; i++ {
			rows, err := db.QueryContext(ctx, ``+
				`SELECT 1`+
				` UNION ALL`+
				` SELECT 2`)
			if err != nil {
				t.Errorf("Query: %v", err)
				return
			}

			err = sqlfunc.ForEach(rows, func(n int) error {
				t.Log(n)
				return nil
			})
			if err != nil {
				t.Errorf("ScanRows: %v", err)
				return
			}
			t.Log(time.Since(start))
		}
	}

	sqlfunc.InternalRegistry.ForEach.Disable(true)
	t.Run("registryDISABLED", testForEachMulti)
	sqlfunc.InternalRegistry.ForEach.Disable(false)
	t.Run("registryENABLED", testForEachMulti)
}

func ExampleScan() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	var scan1 func(*sql.Rows, *int) error
	rows, err := db.QueryContext(ctx, ``+
		`SELECT 1`+
		` UNION ALL`+
		` SELECT 2`)
	if err != nil {
		log.Printf("Query1: %v", err)
		return
	}
	defer rows.Close()

	sqlfunc.Scan(&scan1)

	var values1 []int
	for rows.Next() {
		var n int
		if err = scan1(rows, &n); err != nil {
			log.Printf("Scan1: %v", err)
			return
		}
		values1 = append(values1, n)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Next1: %v", err)
	}
	fmt.Println(values1)

	var scan2 func(*sql.Rows) (string, error)
	rows, err = db.QueryContext(ctx, ``+
		`SELECT 'a'`+
		` UNION ALL`+
		` SELECT 'b'`)
	if err != nil {
		log.Printf("Query2: %v", err)
		return
	}
	defer rows.Close()

	sqlfunc.Scan(&scan2)

	var values2 []string
	for rows.Next() {
		s, err := scan2(rows)
		if err != nil {
			log.Printf("Scan2: %v", err)
			return
		}
		values2 = append(values2, s)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Next2: %v", err)
	}
	fmt.Println(values2)

	// Output:
	// [1 2]
	// [a b]
}

func ExampleScan_any() {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	var scan1 func(*sql.Rows, *any) error
	rows, err := db.QueryContext(ctx, ``+
		`SELECT 1`+
		` UNION ALL`+
		` SELECT NULL`+
		` UNION ALL`+
		` SELECT 'a'`)
	if err != nil {
		log.Printf("Query1: %v", err)
		return
	}
	defer rows.Close()

	sqlfunc.Scan(&scan1)

	for rows.Next() {
		var v any
		if err = scan1(rows, &v); err != nil {
			log.Printf("Scan1: %v", err)
			return
		}
		fmt.Printf("%T %#[1]v\n", v)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Next1: %v", err)
	}

	// Output:
	// int64 1
	// <nil> <nil>
	// string "a"
}

func testForEach_oneColumn[T any](
	t *testing.T,
	db interface {
		PrepareContext(context.Context, string) (*sql.Stmt, error)
	},
	query string,
	nbRows int,
) {
	stmt, err := db.PrepareContext(t.Context(), query)
	defer stmt.Close()

	runQuery := func(t *testing.T) *sql.Rows {
		rows, err := stmt.QueryContext(t.Context())
		if err != nil {
			t.Fatalf("Query: %v", err)
			return nil
		}
		return rows
	}

	values := make([]T, 0, nbRows)

	t.Run("manual", func(t *testing.T) {
		values = values[:0]
		rows := runQuery(t)
		for rows.Next() {
			var v T
			if err = rows.Scan(&v); err != nil {
				t.Error(err)
				break
			}
			values = append(values, v)
		}
		if err = rows.Err(); err != nil {
			t.Error(err)
		}
		if err = rows.Close(); err != nil {
			t.Error(err)
		}
		if len(values) != nbRows {
			t.Fatal("unexpected result")
		}
		t.Logf("values: %#v", values)
	})

	t.Run("sqlfunc.ForEach-void", func(t *testing.T) {
		t.Logf("%T", func(T) {})
		values = values[:0]
		rows := runQuery(t)
		err = sqlfunc.ForEach(rows, func(v T) {
			values = append(values, v)
		})
		if err != nil {
			t.Error(err)
		}
		if len(values) != nbRows {
			t.Fatal("unexpected result")
		}
		t.Logf("values: %#v", values)
	})

	t.Run("sqlfunc.ForEach-error", func(t *testing.T) {
		t.Logf("%T", func(T) error { return nil })
		values = values[:0]
		rows := runQuery(t)
		err = sqlfunc.ForEach(rows, func(v T) error {
			values = append(values, v)
			return nil
		})
		if err != nil {
			t.Error(err)
		}
		if len(values) != nbRows {
			t.Fatal("unexpected result")
		}
		t.Logf("values: %#v", values)
	})

	t.Run("sqlfunc.ForEach-bool", func(t *testing.T) {
		t.Logf("%T", func(T) bool { return true })
		values = values[:0]
		rows := runQuery(t)
		err = sqlfunc.ForEach(rows, func(v T) bool {
			values = append(values, v)
			return true // Continue iterating
		})
		if err != nil {
			t.Error(err)
		}
		if len(values) != nbRows {
			t.Fatal("unexpected result")
		}
		t.Logf("values: %#v", values)
	})

}

func TestForEach_oneColumn(t *testing.T) {
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:?cache=shared")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	db.SetMaxOpenConns(1)

	const nbRows = 5

	t.Run("oneColumn_int", func(t *testing.T) {
		var query = `SELECT 1`
		for i := 2; i <= nbRows; i++ {
			query += fmt.Sprint(` UNION ALL SELECT `, i)
		}

		testForEach_oneColumn[int](t, db, query, nbRows)
	})

	t.Run("oneColumn_string", func(t *testing.T) {
		var query = `SELECT 'a'`
		for i := 2; i <= nbRows; i++ {
			query += fmt.Sprintf(` UNION ALL SELECT '%c'`, rune('a'+i-1))
		}

		testForEach_oneColumn[string](t, db, query, nbRows)
	})

	t.Log("Get connection...")
	conn, err := db.Conn(t.Context())
	if err != nil {
		t.Errorf("Conn: %v", err)
		return
	}

	t.Log("Create table...")
	_, err = conn.ExecContext(t.Context(), `CREATE TABLE t1 (dt DATETIME)`)
	if err != nil {
		t.Errorf("Create table: %v", err)
		return
	}
	_, err = conn.ExecContext(t.Context(), `INSERT INTO t1 (dt) VALUES (datetime('2026-02-20 15:19:56'))`)
	if err != nil {
		t.Errorf("Insert: %v", err)
		return
	}

	t.Run("oneColumn_Time", func(t *testing.T) {
		const query = `SELECT dt FROM t1`

		testForEach_oneColumn[time.Time](t, conn, query, 1)
	})
	t.Run("oneColumn_PtrTime", func(t *testing.T) {
		const query = `SELECT dt FROM t1`

		testForEach_oneColumn[*time.Time](t, conn, query, 1)
	})
	t.Run("oneColumn_PtrTimeNULL", func(t *testing.T) {
		const query = `SELECT NULL FROM t1`

		testForEach_oneColumn[*time.Time](t, conn, query, 1)
	})

	t.Run("oneColumn_NullTime", func(t *testing.T) {
		const query = `SELECT dt FROM t1`

		testForEach_oneColumn[sql.NullTime](t, conn, query, 1)
	})

	t.Run("oneColumn_NullTimeNULL", func(t *testing.T) {
		const query = `SELECT NULL FROM t1`

		testForEach_oneColumn[sql.NullTime](t, conn, query, 1)
	})
}

func benchmarkForEach_oneColumn[T any](
	b *testing.B,
	db interface {
		PrepareContext(context.Context, string) (*sql.Stmt, error)
	},
	query string,
	nbRows int,
) {
	stmt, err := db.PrepareContext(b.Context(), query)
	defer stmt.Close()

	runQuery := func(b *testing.B) *sql.Rows {
		rows, err := stmt.QueryContext(b.Context())
		if err != nil {
			b.Fatalf("Query: %v", err)
			return nil
		}
		return rows
	}

	values := make([]T, 0, nbRows)

	b.Run("manual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				var v T
				if err = rows.Scan(&v); err != nil {
					b.Error(err)
					break
				}
				values = append(values, v)
			}
			if err = rows.Err(); err != nil {
				b.Error(err)
			}
			if err = rows.Close(); err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-void", func(b *testing.B) {
		b.Logf("%T", func(T) {})
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(v T) {
				values = append(values, v)
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-error", func(b *testing.B) {
		b.Logf("%T", func(T) error { return nil })
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(v T) error {
				values = append(values, v)
				return nil
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-bool", func(b *testing.B) {
		b.Logf("%T", func(T) bool { return true })
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(n T) bool {
				values = append(values, n)
				return true // Continue iterating
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})
}

func benchmarkForEach_fiveColumns[T, U, V, W, X any](
	b *testing.B,
	db interface {
		PrepareContext(context.Context, string) (*sql.Stmt, error)
	},
	query string,
	nbRows int,
) {
	stmt, err := db.PrepareContext(b.Context(), query)
	defer stmt.Close()

	runQuery := func(b *testing.B) *sql.Rows {
		rows, err := stmt.QueryContext(b.Context())
		if err != nil {
			b.Fatalf("Query: %v", err)
			return nil
		}
		return rows
	}

	type Row struct {
		A T
		B U
		C V
		D W
		E X
	}

	values := make([]Row, 0, nbRows)

	b.Run("manual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				var r Row
				if err = rows.Scan(&r.A, &r.B, &r.C, &r.D, &r.E); err != nil {
					b.Error(err)
					break
				}
				values = append(values, r)
			}
			if err = rows.Err(); err != nil {
				b.Error(err)
			}
			if err = rows.Close(); err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-void", func(b *testing.B) {
		b.Logf("%T", func(T, U, V, W, X) {})
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(a T, b U, c V, d W, e X) {
				values = append(values, Row{a, b, c, d, e})
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-error", func(b *testing.B) {
		b.Logf("%T", func(T, U, V, W, X) error { return nil })
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(a T, b U, c V, d W, e X) error {
				values = append(values, Row{a, b, c, d, e})
				return nil
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-bool", func(b *testing.B) {
		b.Logf("%T", func(T, U, V, W, X) bool { return true })
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(a T, b U, c V, d W, e X) bool {
				values = append(values, Row{a, b, c, d, e})
				return true // Continue iterating
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})
}

func BenchmarkForEach(b *testing.B) {
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	const nbRows = 500

	b.Run("oneColumn_int", func(b *testing.B) {
		var query = `SELECT 1`
		for i := 2; i <= nbRows; i++ {
			query += fmt.Sprint(` UNION ALL SELECT `, i)
		}

		benchmarkForEach_oneColumn[int](b, db, query, nbRows)
	})

	b.Run("oneColumn_string", func(b *testing.B) {
		const oneRow = `SELECT 'abcdefghijklmnopqrstuvwxyz'`
		var query = oneRow + strings.Repeat(` UNION ALL `+oneRow, nbRows-1)

		benchmarkForEach_oneColumn[string](b, db, query, nbRows)
	})

	/*
		b.Run("oneColumn_Time", func(b *testing.B) {
			const oneRow = `SELECT CAST(datetime('2026-02-19 11:34:56') AS DATETIME)`
			var query = oneRow + strings.Repeat(` UNION ALL `+oneRow, nbRows-1)

			benchmarkForEach_oneColumn[time.Time](b, db, query, nbRows)
		})
	*/

	b.Run("fiveColumns", func(b *testing.B) {
		const oneRow = `SELECT 'abcdefghijklmnopqrstuvwxyz' "str", 1, 2, 'abc', 42.42`
		var query = oneRow + strings.Repeat(` UNION ALL `+oneRow, nbRows-1)

		benchmarkForEach_fiveColumns[string, int64, int32, string, float64](b, db, query, nbRows)
	})
}

func BenchmarkScan(b *testing.B) {
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		b.Errorf("Open: %v", err)
		return
	}
	defer db.Close()

	const nbRows = 500

	var query = "SELECT 1"
	for i := 2; i <= nbRows; i++ {
		query += fmt.Sprint(" UNION ALL SELECT ", i)
	}
	stmt, err := db.PrepareContext(b.Context(), query)
	defer stmt.Close()

	runQuery := func(b *testing.B) *sql.Rows {
		rows, err := stmt.QueryContext(b.Context())
		if err != nil {
			b.Fatalf("Query: %v", err)
			return nil
		}
		return rows
	}

	values := make([]int, 0, nbRows)

	b.Run("direct-Scan", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				var n int
				if err := rows.Scan(&n); err != nil {
					b.Fatal(err)
				}
				_ = n
			}
			rows.Close()
		}
	})

	b.Run("sqlfunc.Scan_ptr", func(b *testing.B) {
		b.ReportAllocs()
		var scan func(rows *sql.Rows, n *int) error
		sqlfunc.Scan(&scan)
		b.ResetTimer()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				var n int
				if err := scan(rows, &n); err != nil {
					b.Fatal(err)
				}
				_ = n
			}
			rows.Close()
		}
	})

	// values = values[:0]
	b.Run("sqlfunc.Scan_return", func(b *testing.B) {
		b.ReportAllocs()
		var scan func(rows *sql.Rows) (int, error)
		sqlfunc.Scan(&scan)
		b.ResetTimer()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				n, err := scan(rows)
				if err != nil {
					b.Fatal(err)
				}
				_ = n
			}
			rows.Close()
		}
	})
}

func benchScan_oneColumn[T any](
	tb TestingB,
	conn interface {
		PrepareContext(context.Context, string) (*sql.Stmt, error)
	},
	query string,
	nbRows int,
) {
	var x T
	tb.Run(fmt.Sprintf("%T-%d", x, nbRows), func(tb TestingB) {
		stmt, err := conn.PrepareContext(tb.Context(), query)
		if err != nil {
			tb.Fatalf("Prepare: %v", err)
			return
		}
		defer stmt.Close()

		runQuery := func(tb TestingB) *sql.Rows {
			tb.StopTimer()
			defer tb.StartTimer()

			rows, err := stmt.QueryContext(tb.Context())
			if err != nil {
				tb.Fatalf("Query: %v", err)
				return nil
			}
			return rows
		}

		tb.Run("std", func(tb TestingB) {
			tb.ReportAllocs()
			for tb.Loop() {
				rows := runQuery(tb)
				defer func() {
					if err := rows.Close(); err != nil {
						tb.Errorf("Close: %v", err)
					}
				}()

				var values []T
				for rows.Next() {
					var v T
					if err = rows.Scan(&v); err != nil {
						tb.Fatalf("Scan: %v", err)
						return
					}
					values = append(values, v)
				}

				tb.StopTimer()

				if err = rows.Err(); err != nil {
					tb.Errorf("Next: %v", err)
				}
				if len(values) != nbRows {
					tb.Fatal("unexpected result")
				}

				tb.StartTimer()
			}
		})

		tb.Run("out", func(tb TestingB) {
			tb.ReportAllocs()
			for tb.Loop() {
				rows := runQuery(tb)
				defer func() {
					if err := rows.Close(); err != nil {
						tb.Errorf("Close: %v", err)
					}
				}()

				var scan func(*sql.Rows) (T, error)
				sqlfunc.Scan(&scan)

				var values []T
				for rows.Next() {
					v, err := scan(rows)
					if err != nil {
						tb.Fatalf("Scan: %v", err)
						return
					}
					values = append(values, v)
				}

				tb.StopTimer()

				if err = rows.Err(); err != nil {
					tb.Errorf("Next: %v", err)
				}
				if len(values) != nbRows {
					tb.Fatal("unexpected result")
				}

				tb.StartTimer()
			}
		})

		tb.Run("in", func(tb TestingB) {
			tb.ReportAllocs()
			for tb.Loop() {
				rows := runQuery(tb)
				defer func() {
					if err := rows.Close(); err != nil {
						tb.Errorf("Close: %v", err)
					}
				}()

				var scan func(*sql.Rows, *T) error
				sqlfunc.Scan(&scan)

				var values []T
				for rows.Next() {
					var v T
					if err = scan(rows, &v); err != nil {
						tb.Fatalf("Scan: %v", err)
						return
					}
					values = append(values, v)
				}

				tb.StopTimer()

				if err = rows.Err(); err != nil {
					tb.Errorf("Next: %v", err)
				}
				if len(values) != nbRows {
					tb.Fatal("unexpected result")
				}

				tb.StartTimer()
			}
		})
	})
}

func suiteScan(tb TestingB) {
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:?cache=shared")
	if err != nil {
		tb.Errorf("Open: %v", err)
		return
	}
	defer db.Close()

	benchScan_oneColumn[int](tb, db, `SELECT 1 UNION ALL SELECT 2`, 2)
	benchScan_oneColumn[string](tb, db, `SELECT 'a' UNION ALL SELECT 'b'`, 2)
}

func BenchmarkScanSuite(b *testing.B) {
	suiteScan(TestingBAsB(b))
}

func TestScanSuite(t *testing.T) {
	suiteScan(TestingTAsB(t))
}
