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
	"fmt"
	"log"
	"time"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleExec() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	// As the DB is in-memory, we need to use the same connection for all operations
	conn, err := db.Conn(ctx)
	check("Conn", err)
	defer checkDeferred("conn.Close", conn.Close)

	// POI = Point of Interest
	_, err = conn.ExecContext(ctx, `CREATE TABLE poi (lat DECIMAL, lon DECIMAL, name VARCHAR(255))`)
	check("Create table", err)

	// newPOI is the function that will call the INSERT statement
	var newPOI func(ctx context.Context, lat float32, lon float32, name string) (sql.Result, error)
	closeStmt, err := sqlfunc.Exec(
		ctx, conn,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&newPOI,
	)
	check("Prepare newPOI", err)
	defer checkDeferred("closeStmt", closeStmt)

	// To call the prepared statement we use the strongly typed function
	_, err = newPOI(ctx, 48.8016, 2.1204, "Château de Versailles")
	check("newPOI", err)

	var name string
	err = conn.QueryRowContext(ctx, ``+
		`SELECT name`+
		` FROM poi`+
		` WHERE lat BETWEEN 48.8015 AND 48.8017`+
		` AND lon BETWEEN 2.1203 AND 2.1205`,
	).Scan(&name)
	check("Query", err)

	fmt.Println(name)

	var getPOICoord func(ctx context.Context, name string) (lat float64, lon float64, err error)
	closeStmt, err = sqlfunc.QueryRow(
		ctx, conn, ``+
			`SELECT lat, lon`+
			` FROM poi`+
			` WHERE name = ?`,
		&getPOICoord,
	)
	check("Prepare getPOICoord", err)
	defer checkDeferred("closeStmt", closeStmt)

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
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, ":memory:")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	// As the DB is in-memory, we need to use the same connection for all operations
	conn, err := db.Conn(ctx)
	check("Conn", err)
	defer checkDeferred("conn.Close", conn.Close)

	// POI = Point of Interest
	_, err = conn.ExecContext(ctx, `CREATE TABLE poi (lat DECIMAL, lon DECIMAL, name VARCHAR(255))`)
	check("Create table", err)

	var countPOI func(ctx context.Context) (int64, error)
	closeCountPOI, err := sqlfunc.QueryRow(
		ctx, conn,
		`SELECT COUNT(*) FROM poi`,
		&countPOI,
	)
	check("Prepare countPOI", err)
	defer checkDeferred("closeCountPOI", closeCountPOI)

	var queryNames func(ctx context.Context) (*sql.Rows, error)
	closeQueryNames, err := sqlfunc.Query(
		ctx, conn,
		`SELECT name FROM poi ORDER BY name`,
		&queryNames,
	)
	check("Prepare queryNames", err)
	defer checkDeferred("closeQueryNames", closeQueryNames)

	nbPOI, err := countPOI(ctx)
	check("countPOI", err)

	fmt.Println("countPOI before insert:", nbPOI)

	var insertPOI func(ctx context.Context, tx *sql.Tx, lat, lon float64, name string) (sql.Result, error)
	closeInsertPOI, err := sqlfunc.Exec(
		ctx, conn,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&insertPOI,
	)
	check("Prepare insertPOI", err)
	defer checkDeferred("closeInsertPOI", closeInsertPOI)

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	check("BeginTx", err)
	defer func() {
		if tx != nil {
			check("tx.Rollback", tx.Rollback())
		}
	}()

	res, err := insertPOI(ctx, tx, 48.8016, 2.1204, "Château de Versailles")
	check("newPOI", err)

	nbRows, err := res.RowsAffected()
	check("RowsAffected", err)
	fmt.Println("Rows inserted:", nbRows)

	res, err = insertPOI(ctx, tx, 47.2009, 0.6317, "Villeperdue")
	check("newPOI", err)

	nbRows, err = res.RowsAffected()
	check("RowsAffected", err)
	fmt.Println("Rows inserted:", nbRows)

	nbPOI, err = countPOI(ctx)
	check("countPOI", err)
	fmt.Println("countPOI after inserts:", nbPOI)

	rows, err := queryNames(ctx)
	check("queryNames", err)
	var names []string
	err = sqlfunc.ForEach(rows, func(name string) {
		names = append(names, name)
	})
	check("ForEach", err)
	fmt.Println("names:", names)

	check("tx.Rollback", tx.Rollback())
	tx = nil // avoid double rollback in defer

	nbPOI, err = countPOI(ctx)
	check("countPOI after rollback", err)

	fmt.Println("countPOI after rollback:", nbPOI)

	// Output:
	// countPOI before insert: 0
	// Rows inserted: 1
	// Rows inserted: 1
	// countPOI after inserts: 2
	// names: [Château de Versailles Villeperdue]
	// countPOI after rollback: 0
}

func ExampleQuery() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	var queryNames func(ctx context.Context) (*sql.Rows, error)
	closeQueryNames, err := sqlfunc.Query(
		ctx, db,
		`SELECT name FROM poi ORDER BY name`,
		&queryNames,
	)
	check("Prepare queryNames", err)
	defer checkDeferred("closeQueryNames", closeQueryNames)

	rows, err := queryNames(ctx)
	check("queryNames", err)
	err = sqlfunc.ForEach(rows, func(name string) {
		fmt.Println("-", name)
	})
	check("read rows", err)

	// Output:
	// - Château de Versailles
	// - Villeperdue
}

func ExampleQuery_withArgs() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	var queryByName func(ctx context.Context, name string) (*sql.Rows, error)
	closeQueryByName, err := sqlfunc.Query(
		ctx, db,
		`SELECT lat, lon FROM poi WHERE name = ?`,
		&queryByName,
	)
	check("Prepare queryByName", err)
	defer checkDeferred("closeQueryByName", closeQueryByName)

	rows, err := queryByName(ctx, "Château de Versailles")
	check("queryByName", err)
	err = sqlfunc.ForEach(rows, func(lat, lon float64) {
		fmt.Printf("(%.4f %.4f)\n", lat, lon)
	})
	check("read rows", err)

	// Output:
	// (48.8016 2.1204)
}

func ExampleQuery_withTx() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("Close", db.Close)

	var queryByNameTx func(ctx context.Context, tx *sql.Tx, name string) (*sql.Rows, error)
	closeQueryByNameTx, err := sqlfunc.Query(
		ctx, db,
		`SELECT lat, lon FROM poi WHERE name = ?`,
		&queryByNameTx,
	)
	check("Prepare queryByName", err)
	defer checkDeferred("closeQueryByNameTx", closeQueryByNameTx)

	tx, err := db.BeginTx(ctx, nil)
	check("BeginTx", err)
	defer checkDeferred("Rollback", tx.Rollback)

	rows, err := queryByNameTx(ctx, tx, "Château de Versailles")
	check("queryByNameTx", err)
	err = sqlfunc.ForEach(rows, func(lat, lon float64) {
		fmt.Printf("(%.4f %.4f)\n", lat, lon)
	})
	check("ForEach", err)

	// Output:
	// (48.8016 2.1204)
}

func ExampleQueryRow_withArgs() {
	check := func(msg string, err error) {
		if err != nil {
			panic(fmt.Errorf("%s: %v", msg, err))
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	var queryByName func(ctx context.Context, name string) (lat, lon float64, err error)
	closeQueryByName, err := sqlfunc.QueryRow(
		ctx, db,
		`SELECT lat, lon FROM poi WHERE name = ?`,
		&queryByName,
	)
	check("Prepare queryByName", err)
	defer checkDeferred("closeQueryByName", closeQueryByName)

	lat, lon, err := queryByName(ctx, "Château de Versailles")
	check("queryByName", err)
	fmt.Printf("(%.4f %.4f)\n", lat, lon)

	// Output:
	// (48.8016 2.1204)
}
