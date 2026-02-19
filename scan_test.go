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
	"testing"
	"time"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleForEach() {
	ctx := context.Background()
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
	ctx := context.Background()
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
	ctx := context.Background()
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
		ctx := context.Background()
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
	ctx := context.Background()
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
	ctx := context.Background()
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	var scan1 func(*sql.Rows, *interface{}) error
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
		var v interface{}
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

func BenchmarkForEach(b *testing.B) {
	ctx := context.Background()
	// As the DB is in-memory, we need to use the same connection for all operations that change the DB state
	db, err := sql.Open(sqliteDriver, ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	const nbRows = 500

	var query = "SELECT 1"
	for i := 2; i <= nbRows; i++ {
		query += fmt.Sprint(" UNION ALL SELECT ", i)
	}
	stmt, err := db.PrepareContext(ctx, query)
	defer stmt.Close()

	runQuery := func(b *testing.B) *sql.Rows {
		rows, err := stmt.Query()
		if err != nil {
			b.Fatalf("Query: %v", err)
			return nil
		}
		return rows
	}

	values := make([]int, 0, nbRows)

	b.Run("manual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			for rows.Next() {
				var n int
				if err = rows.Scan(&n); err != nil {
					b.Error(err)
					break
				}
				values = append(values, n)
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

	b.Run("sqlfunc.ForEach", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(n int) {
				values = append(values, n)
			})
			if err != nil {
				b.Error(err)
			}
			if len(values) != nbRows {
				b.Fatal("unexpected result")
			}
		}
	})

	b.Run("sqlfunc.ForEach-err", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			values = values[:0]
			rows := runQuery(b)
			err = sqlfunc.ForEach(rows, func(n int) error {
				values = append(values, n)
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
}

func BenchmarkScan(b *testing.B) {
	ctx := context.Background()
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
	stmt, err := db.PrepareContext(ctx, query)
	defer stmt.Close()

	runQuery := func(b *testing.B) *sql.Rows {
		rows, err := stmt.Query()
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
