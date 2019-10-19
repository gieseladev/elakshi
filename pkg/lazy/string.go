package lazy

import (
	"sync"
	"sync/atomic"
)

// StringL is a function returning a lazily evaluated string.
type StringL func() string

type lazyString struct {
	f     func() string
	value string
	mux   sync.Mutex
	done  uint32
}

func (v *lazyString) Get() string {
	// check if already evaluated without using the lock
	if atomic.LoadUint32(&v.done) == 1 {
		return v.value
	}

	v.mux.Lock()
	defer v.mux.Unlock()

	if v.done == 0 {
		v.value = v.f()
		v.done = 1
		// remove f to save memory
		v.f = nil
	}
	return v.value
}

// StringFunc creates a lazy string which evaluates to the return value of f.
// f is called exactly once and the result is stored.
func StringFunc(f func() string) StringL {
	return (&lazyString{f: f}).Get
}

// StringConst creates a lazy string which evaluates to the value of s.
func StringConst(s string) StringL {
	return func() string {
		return s
	}
}
