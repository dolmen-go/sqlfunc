# sqlfunc - Stronger typing for database/sql prepared statements

[![Go Reference](https://pkg.go.dev/badge/github.com/dolmen-go/sqlfunc.svg)](https://pkg.go.dev/github.com/dolmen-go/sqlfunc)
[![CI](https://github.com/dolmen-go/sqlfunc/actions/workflows/test.yml/badge.svg)](https://github.com/dolmen-go/sqlfunc/actions)
[![Coverage](https://codecov.io/gh/dolmen-go/sqlfunc/branch/master/graph/badge.svg)](https://app.codecov.io/gh/dolmen-go/sqlfunc)
[![Go Report Card](https://goreportcard.com/badge/github.com/dolmen-go/sqlfunc)](https://goreportcard.com/report/github.com/dolmen-go/sqlfunc)

## Status

Production ready.

Check [code coverage by the testsuite](https://app.codecov.io/gh/dolmen-go/sqlfunc).

### Known issues

* There is a speed/memory penalty in using the `sqlfunc` wrappers
  (check `go test -bench B -benchmem github.com/dolmen-go/sqlfunc`).
  It is recommended to do your own benchmarks. There are plans to fix
  that (add a code generator to reduce cost of runtime `reflect`),
  but no release date planned for this complex feature.

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
