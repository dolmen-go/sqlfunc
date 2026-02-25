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

package registry

import (
	"database/sql"
	"reflect"
	"sync"
	"sync/atomic"
)

func init() {
	ForEach.init()
	Scan.init()
	Stmt.init()
}

var (
	ForEach registryOf[FuncForEach]
	Scan    registryOf[FuncScan]
	// Stmt is the shared registry for Exec, QueryRow, Query.
	// This is possible because the shapes of the return types never overlap.
	Stmt registryOf[FuncStmt]
)

type (
	FuncForEach = func(*sql.Rows, any) error
	FuncScan    = reflect.Value
	FuncStmt    = func(stmt *sql.Stmt) reflect.Value // Exec, QueryRow, Query
)

type registryOf[T any] struct {
	disabled uint32
	m        sync.RWMutex
	r        map[reflect.Type]T
}

func (r *registryOf[T]) init() {
	r.r = make(map[reflect.Type]T)
}

func (r *registryOf[T]) Disable(ig bool) {
	v := uint32(0)
	if ig {
		v = 1
	}
	atomic.StoreUint32(&r.disabled, v)
}

func (r *registryOf[T]) Get(typ reflect.Type) T {
	if atomic.LoadUint32(&r.disabled) != 0 {
		var v T
		return v
	}
	r.m.RLock()
	defer r.m.RUnlock()
	return r.r[typ]
}

func (r *registryOf[T]) Register(typ reflect.Type, v T) {
	r.m.Lock()
	defer r.m.Unlock()
	r.r[typ] = v
}
