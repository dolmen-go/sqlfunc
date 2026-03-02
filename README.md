# sqlfunc - Stronger typing for [database/sql](https://pkg.go.dev/database/sql) prepared statements

[![Go Reference](https://pkg.go.dev/badge/github.com/dolmen-go/sqlfunc.svg)](https://pkg.go.dev/github.com/dolmen-go/sqlfunc)
[![CI](https://github.com/dolmen-go/sqlfunc/actions/workflows/test.yml/badge.svg)](https://github.com/dolmen-go/sqlfunc/actions)
[![Coverage](https://codecov.io/gh/dolmen-go/sqlfunc/branch/master/graph/badge.svg)](https://app.codecov.io/gh/dolmen-go/sqlfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/dolmen-go/sqlfunc)](https://goreportcard.com/report/github.com/dolmen-go/sqlfunc)

## Demo

```go
var queryPersonsByZip func(ctx context.Context, zipCode string) (*sql.Rows, error)

close, _ = sqlfunc.Query(db, ``+
  `SELECT name, age `+
  `FROM person `+
  `WHERE zipcode = ?`,
  &queryPersonsByZip)
defer close()

rows, _ = queryPersonsByZip(ctx, "10017")

_ = sqlfunc.ForEach(rows, func(name string, age int) {
    fmt.Printf("Name: %s, Age: %d\n", name, age)
})
```

Note: error handling is replaced with `_` for this demo.

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
