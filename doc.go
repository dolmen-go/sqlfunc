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

// Package sqlfunc provides utilities to bind SQL statements to strongly-typed Go functions.
//
// By leveraging generics and reflection, it reduces boilerplate and minimizes common
// [database/sql] errors like incorrect column scanning or argument count mismatches.
//
// # Core Concepts
//
//   - [Exec], [QueryRow], and [Query]: Bind SQL statements to function variables.
//   - [ForEach]: Iterate over [*sql.Rows] using a callback with automatically scanned arguments.
//   - [Scan]: Create reusable row-scanning functions.
//   - [Any]: Similar API, but reduced compile-time constraints for more flexibility.
//
// # Example: Binding a Query
//
// You just have to define the function signature you need as a variable:
//
//	var getPOI func(ctx context.Context, name string) (lat float64, lon float64, err error)
//
// and bind it to an SQL statement:
//
//	close, err := sqlfunc.QueryRow(ctx, db, `SELECT lat, lon FROM poi WHERE name = ?`, &getPOI)
//	if err != nil { /* ... */ }
//	defer close()
//
// You can now use the function:
//
//	lat, lon, err := getPOI(ctx, "Château de Versailles")
//
// # Example: Iterating over Rows
//
//	rows, err := db.QueryContext(ctx, "SELECT name, lat, lon FROM poi")
//	if err != nil { /* ... */ }
//
//	err = sqlfunc.ForEach(rows, func(name string, lat, lon float64) {
//	    fmt.Printf("%s: %.4f, %.4f\n", name, lat, lon)
//	})
//
// # Transaction Support
//
// Generated functions can optionally participate in transactions by accepting a [*sql.Tx]
// as their second argument (after [context.Context]):
//
//	var insertPOI func(ctx context.Context, tx *sql.Tx, name string, lat, lon float64) (sql.Result, error)
//
// # Build Tags
//
//   - sqlfunc_registry_on (default): internal cache of bindings is enabled.
//   - sqlfunc_registry_off: internal cache of bindings is disabled.
//   - sqlfunc_registry_sync: internal cache of bindings is enabled, with
//     new bindings being registered synchronously (instead of in the background).
//     Only useful for reliable benchmarks.
//
// Note: the registry will have its maximum impact when the sqlfunc-gen tool will
// be available. Check the [experiment-gen] branch for progress.
//
// [experiment-gen]: https://github.com/dolmen-go/sqlfunc/commits/experiment-gen/
package sqlfunc
