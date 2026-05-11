package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/dolmen-go/sqlfunc"
)

func main() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	_, err = db.ExecContext(ctx, "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	var insertUser func(context.Context, string) (sql.Result, error)
	_, err = sqlfunc.Any.Exec(ctx, db, "INSERT INTO users (name) VALUES (?)", &insertUser)
	if err != nil {
		log.Fatal(err)
	}

	_, err = insertUser(ctx, "Alice")
	if err != nil {
		log.Fatal(err)
	}

	err = sqlfunc.Any.ForEach(nil, func(name string) {
		fmt.Println("User:", name)
	})
	// We don't actually run it as we don't have rows here, but we want to see it generated.
}
