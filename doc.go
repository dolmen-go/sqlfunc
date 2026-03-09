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

// Package sqlfunc provides utilities to wrap SQL prepared statements with strongly typed Go functions.
//
// You just have to define the function signature you need:
//
//	var whoami func(context.Context) (string, error)
//
// and the SQL statement that this function wraps:
//
//	close, err := sqlfunc.QueryRow(ctx, db, `SELECT USER()`, &whoami)  // MySQL example
//	defer close()
//
// You can now use the function:
//
//	user, err := whoami(ctx)
//	fmt.Println("Connected as", user)
//
// # Build tags
//
//   - sqlfunc_registry_on (default): internal cache is enabled.
//   - sqlfunc_registry_off: internal cache is disabled.
//
// Note: the registry will have its maximum impact when the sqlfunc-gen tool will
// be available. Check the [experiment-gen] branch for progress.
//
// [experiment-gen]: https://github.com/dolmen-go/sqlfunc/commits/experiment-gen/
package sqlfunc
