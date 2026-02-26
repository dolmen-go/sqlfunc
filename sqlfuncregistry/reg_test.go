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

package sqlfuncregistry_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dolmen-go/sqlfunc/internal/registry"
	"github.com/dolmen-go/sqlfunc/sqlfuncregistry"
)

// Compilte-time check that the registry functions have the expected signature.
// This enforces the API contract whatever the build tags.
var (
	_ func(registry.FuncForEach)  = sqlfuncregistry.ForEach[func(int64) error]
	_ func(func(int64) error)     = sqlfuncregistry.Scan[func(int64) error]
	_ func(registry.FuncExec)     = sqlfuncregistry.Exec[func(context.Context) (*sql.Rows, error)]
	_ func(registry.FuncQueryRow) = sqlfuncregistry.QueryRow[func(context.Context) (int64, error)]
	_ func(registry.FuncQuery)    = sqlfuncregistry.Query[func(context.Context) (*sql.Rows, error)]
)

func Test(*testing.T) {}
