/*
Copyright 2020 Olivier Mengué

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
	"fmt"
	"log"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleForEach() {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
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

	err = sqlfunc.ForEach(rows, func(n int) (_ error) {
		fmt.Println(n)
		return
	})
	if err != nil {
		log.Printf("ScanRows: %v", err)
		return
	}

	// Output:
	// 1
	// 2
}

func ExampleScan() {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
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
		n, err := scan2(rows)
		if err != nil {
			log.Printf("Scan2: %v", err)
			return
		}
		values2 = append(values2, n)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Next2: %v", err)
	}
	fmt.Println(values2)

	// Output:
	// [1 2]
	// [a b]
}

func BenchmarkScan(b *testing.B) {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
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

	values := make([]int, 0, nbRows)

	b.Run("direct-Scan", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			values = values[:0]
			rows, err := stmt.Query()
			if err != nil {
				log.Println(err)
				break
			}
			for rows.Next() {
				var n int
				if err := rows.Scan(&n); err != nil {
					log.Println(err)
					break
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
		for i := 0; i < b.N; i++ {
			values = values[:0]
			rows, err := stmt.Query()
			if err != nil {
				log.Println(err)
				break
			}
			for rows.Next() {
				var n int
				if err := scan(rows, &n); err != nil {
					log.Println(err)
					break
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
		for i := 0; i < b.N; i++ {
			values = values[:0]
			rows, err := stmt.Query()
			if err != nil {
				log.Println(err)
				break
			}
			for rows.Next() {
				n, err := scan(rows)
				if err != nil {
					log.Println(err)
					break
				}
				_ = n
			}
			rows.Close()
		}
	})
}
