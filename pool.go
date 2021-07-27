package fpack

import (
	"sync"
)

var index []int
var pools []*sync.Pool

func init() {
	// create 16 pools from 1KB to 32MB
	for i := 0; i < 16; i++ {
		size := 1 << (i + 10)
		num := len(pools)
		index = append(index, size)
		pools = append(pools, &sync.Pool{
			New: func() interface{} {
				return &Ref{
					pool:  num,
					slice: make([]byte, size),
				}
			},
		})
	}
}

// Ref is a reference to a borrowed slice.
type Ref struct {
	done  bool
	pool  int
	slice []byte
}

// Release will release the slice.
func (r *Ref) Release() {
	// return if unavailable
	if r == nil || r.done {
		return
	}

	// return
	pools[r.pool].Put(r)
	r.done = true
}

var noop = &Ref{done: true}

// Noop returns a no-op ref.
func Noop() *Ref {
	return noop
}

// Borrow will return a slice that has the specified length. If the requested
// length is too small or too long a slice will be allocated. To recycle the
// slice, it must be released by calling Release() on the returned ref value.
// Always release any returned value, even if the slice grows, it is possible
// to return the originally requested slice.
//
// Note: For values up to 8 bytes (64 bits) the internal Go arena allocator is
// used by calling make(). From benchmarks this seems to be faster than calling
// the pool to borrow and return a value.
func Borrow(len int) ([]byte, *Ref) {
	// select pool
	pool := -1
	for i, max := range index {
		if len < max {
			pool = i
			break
		}
	}

	// allocate if too small or too big
	if len < 9 || pool == -1 {
		return make([]byte, len), noop
	}

	// otherwise get from pool
	ref := pools[pool].Get().(*Ref)
	ref.done = false

	return ref.slice[0:len], ref
}

// Clone will copy the provided slice into a borrowed slice.
func Clone(slice []byte) ([]byte, *Ref) {
	// borrow buffer
	buf, ref := Borrow(len(slice))

	// copy bytes
	copy(buf, slice)

	return buf, ref
}

// Concat will concatenate the provided byte slices using a borrowed slice.
func Concat(slices ...[]byte) ([]byte, *Ref) {
	// compute total length
	var total int
	for _, s := range slices {
		total += len(s)
	}

	// borrow buffer
	buf, ref := Borrow(total)

	// copy bytes
	var pos int
	for _, s := range slices {
		pos += copy(buf[pos:], s)
	}

	return buf, ref
}
