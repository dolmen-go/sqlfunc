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

package sqlfunc

import (
	"context"
	"reflect"
)

// AnyAPI provides a more flexible API, less strictly typed than the main API,
// at the cost of little more runtime overhead and less compile-time safety.
// It is intended for use cases where the main API cannot be used, such as when
// the function signature is not known at compile time.
type AnyAPI [0]struct{}

// Any provides a more flexible API, less strictly typed than the main API,
// at the cost of little more runtime overhead and less compile-time safety.
// It is intended for use cases where the main API cannot be used, such as when
// the function signature is not known at compile time.
var Any AnyAPI

func checkFnPtr(fnPtr any) reflect.Value {
	fnValue := reflect.ValueOf(fnPtr)
	if fnValue.Kind() != reflect.Pointer {
		panic("fnPtr must be a pointer to a *func* variable")
	}
	if fnValue.IsNil() {
		panic("fnPtr must be non-nil")
	}
	return fnValue
}

// Scan is same as [Scan].
func (AnyAPI) Scan(fnPtr any) {
	fnValue := checkFnPtr(fnPtr)
	doScan(fnValue.Type().Elem(), fnValue)
}

// Exec is same as [Exec].
func (AnyAPI) Exec(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return doExec(fnValue.Type().Elem(), ctx, db, query, fnValue)
}

// QueryRow is same as [QueryRow].
func (AnyAPI) QueryRow(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return doQueryRow(fnValue.Type().Elem(), ctx, db, query, fnValue)
}

// Query is same as [Query].
func (AnyAPI) Query(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return doQuery(fnValue.Type().Elem(), ctx, db, query, fnValue)
}
