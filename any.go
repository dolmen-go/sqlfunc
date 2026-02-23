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
	if fnValue.Kind() != reflect.Ptr {
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
	anyScan(fnValue.Type().Elem(), fnValue)
}

// Exec is same as [Exec].
func (AnyAPI) Exec(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return anyExec(fnValue.Type().Elem(), ctx, db, query, fnValue)
}

// QueryRow is same as [QueryRow].
func (AnyAPI) QueryRow(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return anyQueryRow(fnValue.Type().Elem(), ctx, db, query, fnValue)
}

// Query is same as [Query].
func (AnyAPI) Query(ctx context.Context, db PrepareConn, query string, fnPtr any) (close func() error, err error) {
	fnValue := checkFnPtr(fnPtr)
	return anyQuery(fnValue.Type().Elem(), ctx, db, query, fnValue)
}
