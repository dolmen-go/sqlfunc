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
	"reflect"
	"testing"
	"time"

	"github.com/dolmen-go/sqlfunc"
)

func ExampleExec() {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
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

	// newPOI is the function that will call the INSERT statement.
	// We won't use the sql.Result, so it's left out of the returned values.
	var newPOI func(ctx context.Context, lat float32, lon float32, name string) error
	closeStmt, err := sqlfunc.Exec(
		ctx, conn,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&newPOI,
	)
	check("Prepare newPOI", err)
	defer checkDeferred("closeStmt", closeStmt)

	// To call the prepared statement we use the strongly typed function
	err = newPOI(ctx, 48.8016, 2.1204, "Château de Versailles")
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
		log.Fatalf("getPOICoord should fail with sql.ErrNoRows")
	}

	lat, lon, err := getPOICoord(ctx, "Château de Versailles")
	if err != nil {
		log.Fatalf("getPOICoord should succeed but %q", err)
	}
	fmt.Printf("%.4f, %.4f\n", lat, lon)

	// Output:
	// Château de Versailles
	// 48.8016, 2.1204
}

func ExampleAnyAPI_Exec() {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
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

	// newPOI is the function that will call the INSERT statement.
	// We won't use the sql.Result, so it's left out of the returned values.
	var newPOI func(ctx context.Context, lat float32, lon float32, name string) error
	closeStmt, err := sqlfunc.Any.Exec(
		ctx, conn,
		`INSERT INTO poi (lat, lon, name) VALUES (?, ?, ?)`,
		&newPOI,
	)
	check("Prepare newPOI", err)
	defer checkDeferred("closeStmt", closeStmt)

	// To call the prepared statement we use the strongly typed function
	err = newPOI(ctx, 48.8016, 2.1204, "Château de Versailles")
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
	closeStmt, err = sqlfunc.Any.QueryRow(
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
		log.Fatalf("getPOICoord should fail with sql.ErrNoRows")
	}

	lat, lon, err := getPOICoord(ctx, "Château de Versailles")
	if err != nil {
		log.Fatalf("getPOICoord should succeed but %q", err)
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
			log.Fatalf("%s: %v", msg, err)
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

func ExampleAnyAPI_Query() {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	var queryNames func(ctx context.Context) (*sql.Rows, error)
	closeQueryNames, err := sqlfunc.Any.Query(
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

func ExampleAnyAPI_Query_withTx() {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("Close", db.Close)

	var queryByNameTx func(ctx context.Context, tx *sql.Tx, name string) (*sql.Rows, error)
	closeQueryByNameTx, err := sqlfunc.Any.Query(
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
	err = sqlfunc.Any.ForEach(rows, func(lat, lon float64) {
		fmt.Printf("(%.4f %.4f)\n", lat, lon)
	})
	check("ForEach", err)

	// Output:
	// (48.8016 2.1204)
}

func ExampleAnyAPI_Query_withArgs() {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	var queryByName func(ctx context.Context, name string) (*sql.Rows, error)
	closeQueryByName, err := sqlfunc.Any.Query(
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
			log.Fatalf("%s: %v", msg, err)
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
			log.Fatalf("%s: %v", msg, err)
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

type panicConn string

func (scs panicConn) PrepareContext(ctx context.Context, query string) (_ *sql.Stmt, _ error) {
	panic(string(scs))
}

func TestExecInvalidSignatures(t *testing.T) {
	CheckInvalidTargets(t, new(struct {
		Any         any                                                   `panic:"fnPtr must be a pointer to a *func* variable"`
		Int         int                                                   `panic:"fnPtr must be a pointer to a *func* variable"`
		NoArg       func()                                                `panic:"func first arg must be a context.Context"`
		NoContext   func() (sql.Result, error)                            `panic:"func first arg must be a context.Context"`
		NoContext2  func(int64) (sql.Result, error)                       `panic:"func first arg must be a context.Context"`
		NoError     func(context.Context)                                 `panic:"func must return (sql.Result, error) or (error)"`
		NoResult    func(context.Context) (sql.Result, int)               `panic:"func must return (sql.Result, error) or (error)"`
		NotResultIE func(context.Context) (int64, error)                  `panic:"func must return (sql.Result, error) or (error)"`
		NotResultEE func(context.Context) (error, error)                  `panic:"func must return (sql.Result, error) or (error)"`
		NotResultRI func(context.Context) (sql.Result, int)               `panic:"func must return (sql.Result, error) or (error)"`
		NotError    func(context.Context) (sql.Result, sql.Result)        `panic:"func must return (sql.Result, error) or (error)"`
		ResultREE   func(context.Context) (sql.Result, error, error)      `panic:"func must return (sql.Result, error) or (error)"`
		ResultRRE   func(context.Context) (sql.Result, sql.Result, error) `panic:"func must return (sql.Result, error) or (error)"`
		// sql.Result is an interface. Can't be returned as pointer.
		ResultIsPtr func(context.Context) (*sql.Result, error) `panic:"func must return (sql.Result, error) or (error)"`

		NotTxPtr func(context.Context, sql.Tx) (sql.Result, error) `panic:"func should take *sql.Tx, not sql.Tx" todo:"should require *sql.Tx, reject sql.Tx"`

		VariadicInts1   func(context.Context, ...int64) (sql.Result, error)         `panic:"func must not be variadic"`
		VariadicInts2   func(context.Context, string, ...int64) (sql.Result, error) `panic:"func must not be variadic"`
		VariadicContext func(...context.Context) (sql.Result, error)                `panic:"func first arg must be a context.Context"`
	}), func(fnPtr any) {
		_, err := sqlfunc.Any.Exec(context.Background(), panicConn("signature validation failure"), "SELECT 1", fnPtr)
		panic(err)
	})

	t.Run("NilPtr", func(t *testing.T) {
		t.Run("Any", func(t *testing.T) {
			for _, v := range []any{
				nil,
				(func(context.Context) (sql.Result, error))(nil),
				(*func(context.Context) (sql.Result, error))(nil),
			} {
				val := reflect.ValueOf(&v).Elem()
				typ := val.Type()
				if !val.IsNil() {
					typ = val.Elem().Type()
				}
				t.Run(typ.String()+"(nil)", func(t *testing.T) {
					MustPanic(t, [...]string{
						"fnPtr must be non-nil",
						"fnPtr must be a pointer to a *func* variable",
					}, func() {
						sqlfunc.Any.Exec(context.Background(), panicConn("signature validation failure"), "SELECT 1", v)
					})
				})
			}
		})
		t.Run("Typed", func(t *testing.T) {
			MustPanic(t, "fnPtr must be non-nil", func() {
				sqlfunc.Scan((*func(context.Context) (sql.Result, error))(nil))
			})
		})
	})
}

func TestQueryInvalidSignatures(t *testing.T) {
	CheckInvalidTargets(t, new(struct {
		Any         any                                                 `panic:"fnPtr must be a pointer to a *func* variable"`
		Int         int                                                 `panic:"fnPtr must be a pointer to a *func* variable"`
		NoArg       func()                                              `panic:"func first arg must be a context.Context"`
		NotContext  func() (*sql.Rows, error)                           `panic:"func first arg must be a context.Context"`
		NotContext2 func(int64) (*sql.Rows, error)                      `panic:"func first arg must be a context.Context"`
		NoError     func(context.Context)                               `panic:"func must return (*sql.Rows, error)"`
		NotRowsIE   func(context.Context) (int64, error)                `panic:"func must return (*sql.Rows, error)"`
		NotRowsEE   func(context.Context) (error, error)                `panic:"func must return (*sql.Rows, error)"`
		NotRowsRI   func(context.Context) (*sql.Rows, int)              `panic:"func must return (*sql.Rows, error)"`
		NotError    func(context.Context) (*sql.Rows, *sql.Rows)        `panic:"func must return (*sql.Rows, error)"`
		ResultREE   func(context.Context) (*sql.Rows, error, error)     `panic:"func must return (*sql.Rows, error)"`
		ResultRRE   func(context.Context) (*sql.Rows, *sql.Rows, error) `panic:"func must return (*sql.Rows, error)"`

		NotTxPtr func(context.Context, sql.Tx) (*sql.Rows, error) `panic:"func should take *sql.Tx, not sql.Tx" todo:"should require *sql.Tx, reject sql.Tx"`

		VariadicInts1   func(context.Context, ...int64) (*sql.Rows, error)         `panic:"func must not be variadic"`
		VariadicInts2   func(context.Context, string, ...int64) (*sql.Rows, error) `panic:"func must not be variadic"`
		VariadicContext func(...context.Context) (*sql.Rows, error)                `panic:"func first arg must be a context.Context"`
	}), func(fnPtr any) {
		_, err := sqlfunc.Any.Query(context.Background(), panicConn("signature validation failure"), "SELECT 1", fnPtr)
		panic(err)
	})

	t.Run("NilPtr", func(t *testing.T) {
		t.Run("Any", func(t *testing.T) {
			for _, v := range []any{
				nil,
				(func(context.Context) (*sql.Rows, error))(nil),
				(*func(context.Context) (*sql.Rows, error))(nil),
			} {
				val := reflect.ValueOf(&v).Elem()
				typ := val.Type()
				if !val.IsNil() {
					typ = val.Elem().Type()
				}
				t.Run(typ.String()+"(nil)", func(t *testing.T) {
					MustPanic(t, [...]string{
						"fnPtr must be non-nil",
						"fnPtr must be a pointer to a *func* variable",
					}, func() {
						sqlfunc.Any.Exec(context.Background(), panicConn("signature validation failure"), "SELECT 1", v)
					})
				})
			}
		})
		t.Run("Typed", func(t *testing.T) {
			MustPanic(t, "fnPtr must be non-nil", func() {
				sqlfunc.Scan((*func(context.Context) (*sql.Rows, error))(nil))
			})
		})
	})
}

func TestQueryRowInvalidSignatures(t *testing.T) {
	CheckInvalidTargets(t, new(struct {
		Any              any                                     `panic:"fnPtr must be a pointer to a *func* variable"`
		Int              int                                     `panic:"fnPtr must be a pointer to a *func* variable"`
		NoArg            func()                                  `panic:"func first arg must be a context.Context"`
		NoContext1       func() (int64, error)                   `panic:"func first arg must be a context.Context"`
		NoContext2       func(int64) (int64, error)              `panic:"func first arg must be a context.Context"`
		NoError          func(context.Context)                   `panic:"func must return either (*sql.Row) or (values..., error)"`
		NoResult         func(context.Context) error             `panic:"func must return at least one column"`
		NoErrorI         func(context.Context) int64             `panic:"func must return an error" todo:"should return an error"`
		NoErrorII        func(context.Context) (int64, int64)    `panic:"func must return an error"`
		ReturnPtrPtr     func(context.Context) (**int64, error)  `panic:"func must not return double pointer" todo:"should reject double pointer"`
		ReturnRowPlusErr func(context.Context) (*sql.Row, error) `panic:"func must return ONLY *sql.Row"`

		// Signature must not overlap with Exec's or Query's signatures
		LikeExec  func(context.Context) (sql.Result, error) `panic:"func must return either (*sql.Row) or (values..., error)"`
		LikeQuery func(context.Context) (*sql.Rows, error)  `panic:"func must return either (*sql.Row) or (values..., error)"`

		NotTxPtr func(context.Context, sql.Tx) (*sql.Row, error) `panic:"func should take *sql.Tx, not sql.Tx" todo:"should require *sql.Tx, reject sql.Tx"`

		VariadicInts1   func(context.Context, ...int64) (int64, error)         `panic:"func must not be variadic"`
		VariadicInts2   func(context.Context, string, ...int64) (int64, error) `panic:"func must not be variadic"`
		VariadicContext func(...context.Context) (int64, error)                `panic:"func first arg must be a context.Context"`
	}), func(fnPtr any) {
		_, err := sqlfunc.Any.QueryRow(context.Background(), panicConn("signature validation failure"), "SELECT 1", fnPtr)
		panic(err)
	})

	t.Run("NilPtr", func(t *testing.T) {
		t.Run("Any", func(t *testing.T) {
			for _, v := range []any{
				nil,
				(func(context.Context) *sql.Row)(nil),
				(*func(context.Context) *sql.Row)(nil),
			} {
				val := reflect.ValueOf(&v).Elem()
				typ := val.Type()
				if !val.IsNil() {
					typ = val.Elem().Type()
				}
				t.Run(typ.String()+"(nil)", func(t *testing.T) {
					MustPanic(t, [...]string{
						"fnPtr must be non-nil",
						"fnPtr must be a pointer to a *func* variable",
					}, func() {
						sqlfunc.Any.Exec(context.Background(), panicConn("signature validation failure"), "SELECT 1", v)
					})
				})
			}
		})
		t.Run("Typed", func(t *testing.T) {
			MustPanic(t, "fnPtr must be non-nil", func() {
				sqlfunc.Scan((*func(context.Context) *sql.Row)(nil))
			})
		})
	})
}

func TestStmt(t *testing.T) {
	benchStmt(TestingTAsB(t))
}

func BenchmarkStmt(b *testing.B) {
	benchStmt(TestingBAsB(b))
}

func benchStmt(b TestingB) {
	check := func(msg string, err error) {
		if err != nil {
			log.Fatalf("%s: %v", msg, err)
		}
	}
	checkDeferred := func(msg string, f func() error) { check(msg, f()) }

	db, err := sql.Open(sqliteDriver, "file:testdata/poi.db?mode=ro&immutable=1")
	check("Open", err)
	defer checkDeferred("db.Close", db.Close)

	b.Run("Query", func(b TestingB) {
		const query = `SELECT name FROM poi ORDER BY name`
		type queryFunc func(ctx context.Context) (*sql.Rows, error)

		b.Run("Manual", func(b TestingB) {
			start := time.Now()
			stmt, err := db.PrepareContext(b.Context(), query)
			b.Log("db.Prepare time:", time.Since(start))
			check("Prepare", err)
			defer checkDeferred("closeStmt", stmt.Close)

			b.ResetTimer()
			for b.Loop() {
				func() error {
					rows, err := stmt.QueryContext(b.Context())
					check("Query", err)
					defer checkDeferred("Query Close", rows.Close)
					for rows.Next() {
						var name string
						if err := rows.Scan(&name); err != nil {
							return err
						}
					}
					return nil
				}()
			}

		})

		runQuery := func(b TestingB) {
			var queryNames queryFunc
			start := time.Now()
			closeQueryNames, err := sqlfunc.Query(
				b.Context(), db,
				query,
				&queryNames,
			)
			b.Log("sqlfunc.Query time:", time.Since(start))

			check("Prepare queryNames", err)
			defer checkDeferred("closeQueryNames", closeQueryNames)

			b.ResetTimer()
			for b.Loop() {
				func() error {
					rows, err := queryNames(b.Context())
					check("queryNames", err)
					defer checkDeferred("queryNames Close", rows.Close)
					for rows.Next() {
						var name string
						if err := rows.Scan(&name); err != nil {
							return err
						}
					}
					return rows.Err()
				}()
			}
		}

		b.Run("First", runQuery)
		b.Run("Second", runQuery)

	})

	b.Run("QueryRow", func(b TestingB) {
		const query = `SELECT name FROM poi WHERE name = ?`
		type queryRowFunc func(ctx context.Context, name string) (string, error)

		b.Run("Manual", func(b TestingB) {
			start := time.Now()
			stmt, err := db.PrepareContext(b.Context(), query)
			b.Log("db.Prepare time:", time.Since(start))
			check("Prepare", err)
			defer checkDeferred("closeStmt", stmt.Close)

			b.ResetTimer()
			for b.Loop() {
				var name string
				err := stmt.QueryRowContext(b.Context(), "Château de Versailles").Scan(&name)
				check("QueryRow", err)
			}
		})

		runQueryRow := func(b TestingB) {
			var queryName queryRowFunc
			start := time.Now()
			closeQueryName, err := sqlfunc.QueryRow(b.Context(), db, query, &queryName)
			b.Log("sqlfunc.QueryRow time:", time.Since(start))
			check("Prepare queryName", err)
			defer checkDeferred("closeQueryName", closeQueryName)

			b.ResetTimer()
			for b.Loop() {
				_, err := queryName(b.Context(), "Château de Versailles")
				check("queryName", err)
			}
		}

		b.Run("First", runQueryRow)
		b.Run("Second", runQueryRow)
	})

	b.Run("Exec", func(b TestingB) {
		db, err := sql.Open(sqliteDriver, ":memory:?cache=shared")
		check("Open", err)
		defer checkDeferred("db.Close", db.Close)

		_, err = db.ExecContext(b.Context(), `CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)`)
		check("Create table", err)

		const query = `INSERT INTO test (name) VALUES (?)`
		type execFunc func(ctx context.Context, name string) (sql.Result, error)

		b.Run("Manual", func(b TestingB) {
			start := time.Now()
			stmt, err := db.PrepareContext(b.Context(), query)
			b.Log("db.Prepare time:", time.Since(start))
			check("Prepare", err)
			defer checkDeferred("closeStmt", stmt.Close)

			b.ResetTimer()
			for b.Loop() {
				_, err := stmt.ExecContext(b.Context(), "test")
				check("Exec", err)
			}
		})

		runExec := func(b TestingB) {
			var insert execFunc
			start := time.Now()
			closeInsert, err := sqlfunc.Exec(b.Context(), db, query, &insert)
			b.Log("sqlfunc.Exec time:", time.Since(start))
			check("Prepare insert", err)
			defer checkDeferred("closeInsert", closeInsert)

			b.ResetTimer()
			for b.Loop() {
				_, err := insert(b.Context(), "test")
				check("insert", err)
			}
		}

		b.Run("First", runExec)
		b.Run("Second", runExec)
	})

}
