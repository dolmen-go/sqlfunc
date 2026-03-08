# sqlfunc - Stronger typing for [database/sql](https://pkg.go.dev/database/sql) prepared statements

[![Go Reference](https://pkg.go.dev/badge/github.com/dolmen-go/sqlfunc.svg)](https://pkg.go.dev/github.com/dolmen-go/sqlfunc)
[![CI](https://github.com/dolmen-go/sqlfunc/actions/workflows/test.yml/badge.svg)](https://github.com/dolmen-go/sqlfunc/actions)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=coverage)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/dolmen-go/sqlfunc)](https://goreportcard.com/report/github.com/dolmen-go/sqlfunc)

<!--
[![Coverage](https://codecov.io/gh/dolmen-go/sqlfunc/branch/master/graph/badge.svg)](https://app.codecov.io/gh/dolmen-go/sqlfunc)
-->

`sqlfunc` is a Go library that simplifies the use of `database/sql` by binding SQL queries directly to strongly-typed Go functions. By leveraging Go generics and reflection, it provides an idiomatic and type-safe API that reduces boilerplate and minimizes common errors like incorrect column scanning or argument count mismatches.

### Key Features

- **Strongly-Typed Query Binding**: Bind `INSERT`, `UPDATE`, `DELETE`, and `SELECT` statements directly to Go function variables. The function signature defines the SQL parameters and the expected result types.
- **[`Exec`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Exec), [`QueryRow`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#QueryRow), and [`Query`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Query)**: Dedicated functions for different SQL operations, ensuring that the bound functions return the appropriate results (`sql.Result`, individual columns, or `*sql.Rows`).
- **[`ForEach`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#ForEach) Iteration**: A high-level helper for iterating over `*sql.Rows` that automatically scans columns into the arguments of a provided callback function.
- **Flexible [`Scan`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Scan) Functions**: Generate reusable functions to scan a single row into pointers or return values, improving readability and reuse.
- **Transaction Support**: Generated functions can optionally accept a `*sql.Tx` as an argument, allowing them to participate seamlessly in transactions.
- **`Any` API**: Provides a more flexible, reflection-heavy API for use cases where function signatures are determined at runtime.
- **Performance-Oriented Registry**: Includes a registry system to cache reflective implementations, designed with future code generation in mind to achieve performance comparable to manual `database/sql` code.

### Example: Binding a Query to a Function

```go
import (
    "context"
    "database/sql"
    "github.com/dolmen-go/sqlfunc"
)

// Define a function variable with the desired signature
var getPOI func(ctx context.Context, name string) (lat float64, lon float64, err error)

// Bind the SQL query to the function variable
closeStmt, err := sqlfunc.QueryRow(ctx, db,
    "SELECT lat, lon FROM poi WHERE name = ?",
    &getPOI,
)
if err != nil {
    // handle error
}
defer closeStmt()

// Execute the query using the strongly-typed function
lat, lon, err := getPOI(ctx, "Château de Versailles")
```

### Example: Using `ForEach` for Iteration

```go
rows, err := db.QueryContext(ctx, "SELECT name, lat, lon FROM poi")
if err != nil { /* ... */ }

// Scan rows directly into callback arguments
err = sqlfunc.ForEach(rows, func(name string, lat, lon float64) {
    fmt.Printf("%s: %.4f, %.4f\n", name, lat, lon)
})
```

## Documentation

* [Package `sqlfunc`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc)
* [`sqlfunc.ForEach`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#ForEach)
* [`sqlfunc.Query`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Query)
* [`sqlfunc.QueryRow`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#QueryRow)
* [`sqlfunc.Exec`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Exec)
* [`sqlfunc.Scan`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#Scan)
* [`sqlfunc.Any.ForEach`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#AnyAPI.ForEach)
* [`sqlfunc.Any.Query`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#AnyAPI.Query)
* [`sqlfunc.Any.QueryRow`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#AnyAPI.QueryRow)
* [`sqlfunc.Any.Exec`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#AnyAPI.Exec)
* [`sqlfunc.Any.Scan`](https://pkg.go.dev/github.com/dolmen-go/sqlfunc#AnyAPI.Scan)

## Status

Production ready.

Check [code coverage by the testsuite](https://app.codecov.io/gh/dolmen-go/sqlfunc).

### Known issues

* There is a speed/memory penalty in using the `sqlfunc` wrappers
  (check `go test -bench B -benchmem github.com/dolmen-go/sqlfunc`).
  It is recommended to do your own benchmarks. However there is **work in
  progress** to add a code generator to reduce cost of runtime `reflect`.
  Check the [`experiment-gen`](https://github.com/dolmen-go/sqlfunc/commits/experiment-gen/)
  branch.

### Quality reports

See quality reports on [SonarQube](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc).

[![SonarQube Cloud](https://sonarcloud.io/images/project_badges/sonarcloud-highlight.svg)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)


[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=coverage)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=bugs)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=dolmen-go_sqlfunc&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=dolmen-go_sqlfunc)

## License

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
