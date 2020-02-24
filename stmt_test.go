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

	_ "github.com/mattn/go-sqlite3"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleExec() {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Printf("Open: %v", err)
		return
	}
	defer db.Close()

	// POI = Point of Interest
	_, err = db.ExecContext(ctx, `CREATE TABLE poi (lat DECIMAL, lon DECIMAL, name VARCHAR(255))`)
	if err != nil {
		log.Printf("Create table: %v", err)
		return
	}

	// newPOI is the function that will call the INSERT statement
	var newPOI func(ctx context.Context, lat float32, lon float32, name string) (sql.Result, error)
	closeStmt, err := sqlfunc.Exec(
		ctx, db,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&newPOI,
	)
	if err != nil {
		log.Printf("Prepare: %v", err)
		return
	}
	defer closeStmt()

	// To call the prepared statement we use the strongly typed function
	_, err = newPOI(ctx, 48.8016, 2.1204, "Château de Versailles")
	if err != nil {
		log.Printf("newPOI: %v", err)
		return
	}

	var name string
	err = db.QueryRow(`` +
		`SELECT name` +
		` FROM poi` +
		` WHERE lat BETWEEN 48.8015 AND 48.8017` +
		` AND lon BETWEEN 2.1203 AND 2.1205`,
	).Scan(&name)
	if err != nil {
		log.Printf("Query: %v", err)
		return
	}
	fmt.Println(name)

	var getPOICoord func(ctx context.Context, name string) (lat float64, lon float64, err error)
	closeStmt, err = sqlfunc.QueryRow(
		ctx, db, ``+
			`SELECT lat, lon`+
			` FROM poi`+
			` WHERE name = ?`,
		&getPOICoord,
	)
	if err != nil {
		log.Printf("Prepare: %v", err)
		return
	}
	defer closeStmt()

	_, _, err = getPOICoord(ctx, "Trifoully-les-Oies")
	if err != sql.ErrNoRows {
		log.Printf("getPOICoord should fail with sql.ErrNoRows")
		return
	}

	lat, lon, err := getPOICoord(ctx, "Château de Versailles")
	if err != nil {
		log.Printf("getPOICoord should succeed but %q", err)
		return
	}
	fmt.Printf("%.4f, %.4f\n", lat, lon)

	// Output:
	// Château de Versailles
	// 48.8016, 2.1204
}
