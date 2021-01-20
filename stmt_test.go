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

	"github.com/dolmen-go/sqlfunc"
)

func ExampleExec() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}

	ctx := context.Background()
	db, err := sql.Open(sqliteDriver, ":memory:")
	check("Open", err)
	defer db.Close()

	// POI = Point of Interest
	_, err = db.ExecContext(ctx, `CREATE TABLE poi (lat DECIMAL, lon DECIMAL, name VARCHAR(255))`)
	check("Create table", err)

	// newPOI is the function that will call the INSERT statement
	var newPOI func(ctx context.Context, lat float32, lon float32, name string) (sql.Result, error)
	closeStmt, err := sqlfunc.Exec(
		ctx, db,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&newPOI,
	)
	check("Prepare newPOI", err)
	defer closeStmt()

	// To call the prepared statement we use the strongly typed function
	_, err = newPOI(ctx, 48.8016, 2.1204, "Château de Versailles")
	check("newPOI", err)

	var name string
	err = db.QueryRow(`` +
		`SELECT name` +
		` FROM poi` +
		` WHERE lat BETWEEN 48.8015 AND 48.8017` +
		` AND lon BETWEEN 2.1203 AND 2.1205`,
	).Scan(&name)
	check("Query", err)

	fmt.Println(name)

	var getPOICoord func(ctx context.Context, name string) (lat float64, lon float64, err error)
	closeStmt, err = sqlfunc.QueryRow(
		ctx, db, ``+
			`SELECT lat, lon`+
			` FROM poi`+
			` WHERE name = ?`,
		&getPOICoord,
	)
	check("Prepare getPOICoord", err)
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

// ExampleExec_withTx shows support for transactions.
func ExampleExec_withTx() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}

	ctx := context.Background()
	db, err := sql.Open(sqliteDriver, ":memory:")
	check("Open", err)
	defer db.Close()

	// POI = Point of Interest
	_, err = db.ExecContext(ctx, `CREATE TABLE poi (lat DECIMAL, lon DECIMAL, name VARCHAR(255))`)
	check("Create table", err)

	var countPOI func(ctx context.Context) (int64, error)
	closeCountPOI, err := sqlfunc.QueryRow(
		ctx, db,
		`SELECT COUNT(*) FROM poi`,
		&countPOI,
	)
	check("Prepare countPOI", err)
	defer closeCountPOI()

	nbPOI, err := countPOI(ctx)
	check("countPOI", err)

	fmt.Println("countPOI before insert:", nbPOI)

	var insertPOI func(ctx context.Context, tx *sql.Tx, lat, lon float64, name string) (sql.Result, error)
	closeInsertPOI, err := sqlfunc.Exec(
		ctx, db,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&insertPOI,
	)
	check("Prepare insertPOI", err)
	defer closeInsertPOI()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	check("BeginTx", err)
	defer tx.Rollback()

	res, err := insertPOI(ctx, tx, 48.8016, 2.1204, "Château de Versailles")
	check("newPOI", err)

	nbRows, err := res.RowsAffected()
	check("RowsAffected", err)

	fmt.Println("Rows inserted:", nbRows)

	/*
		// FIXME count here too
		nbPOI, err = countPOI(ctx)
		if err != nil {
			log.Println("countPOI after insert:", err)
			return
		}
		fmt.Println("countPOI after insert:", nbPOI)
	*/

	tx.Rollback()

	nbPOI, err = countPOI(ctx)
	check("countPOI after rollback", err)

	fmt.Println("countPOI after rollback:", nbPOI)

	// Output:
	// countPOI before insert: 0
	// Rows inserted: 1
	// countPOI after rollback: 0
}
