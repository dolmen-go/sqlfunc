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
	registry.ForEach.r = make(map[reflect.Type]funcForEach)
}

type privateRegistry struct {
	ForEach registryForEach
}

type funcForEach = func(*sql.Rows, interface{}) error

type registryForEach struct {
	disabled uint32
	m        sync.RWMutex
	r        map[reflect.Type]funcForEach
}

func (r *registryForEach) Disable(ig bool) {
	v := uint32(0)
	if ig {
		v = 1
	}
	atomic.StoreUint32(&r.disabled, v)
}

func (r *registryForEach) Get(typ reflect.Type) funcForEach {
	if atomic.LoadUint32(&r.disabled) != 0 {
		return nil
	}
	r.m.RLock()
	defer r.m.RUnlock()
	return r.r[typ]
}

func (r *registryForEach) Register(t interface{}, f funcForEach) {
	if f == nil {
		return // panic?
	}
	r.m.Lock()
	defer r.m.Unlock()
	r.r[reflect.TypeOf(t)] = f
}
