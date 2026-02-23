package sqlfunc

import (
	"database/sql"
	"reflect"
	"sync"
	"sync/atomic"
)

// Ř is the private registry used by the sqlfunc monomorphizer.
// var Ř privateRegistry
var registry privateRegistry

func init() {
	registry.ForEach.init()
	registry.Scan.init()
}

type privateRegistry struct {
	ForEach registryOf[funcForEach]
	Scan    registryOf[funcScan]
}

type (
	funcForEach = func(*sql.Rows, any) error
	funcScan    = reflect.Value
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

func (r *registryOf[T]) Register(t any, v T) {
	r.m.Lock()
	defer r.m.Unlock()
	r.r[reflect.TypeOf(t)] = v
}
