package fpack

import (
	"sync"
	"sync/atomic"
)

var index []int
var pools []*sync.Pool
var generation uint64

type buffer struct {
	mutex sync.Mutex
	gen   uint64
	pool  int8
	slice []byte
}

func init() {
	// create 16 pools from 1KB to 32MB
	for i := 0; i < 16; i++ {
		num := int8(i)
		size := 1 << (i + 10)
		index = append(index, size)
		pools = append(pools, &sync.Pool{
			New: func() interface{} {
				return &buffer{
					pool:  num,
					slice: make([]byte, size),
				}
			},
		})
	}
}

var zeroRef Ref

// Ref is a reference to a borrowed slice. A zero reference represents a
// no-op reference.
type Ref struct {
	gen uint64
	buf *buffer
}

// Release will release the borrowed slice. The function should be called at
// most once as it protects from double releasing and other invalid access.
func (r Ref) Release() {
	// treat zero refs as no-ops
	if r == zeroRef {
		return
	}

	// acquire mutex
	r.buf.mutex.Lock()
	defer r.buf.mutex.Unlock()

	// check generation
	if r.gen != r.buf.gen {
		panic("fpack: generation mismatch")
	}

	// set count
	r.buf.gen = 0

	// return
	pools[r.buf.pool].Put(r.buf)
}

// Borrow will return a slice that has the specified length. If the requested
// length is too small or too long a slice will be allocated. To recycle the
// slice, it must be released by calling Release() on the returned Ref value.
// Always release any returned value, even if the slice grows, it is possible
// to return the originally requested slice.
//
// Note: For values up to 8 bytes (64 bits) the internal Go arena allocator is
// used by calling make(). From benchmarks this seems to be faster than calling
// the pool to borrow and return a value.
func Borrow(len int) ([]byte, Ref) {
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
		return make([]byte, len), Ref{}
	}

	// get next non zero generation
	var gen uint64
	for gen == 0 {
		gen = atomic.AddUint64(&generation, 1)
	}

	// otherwise get from pool
	buf := pools[pool].Get().(*buffer)

	// set generation
	buf.gen = gen

	// prepare slice
	slice := buf.slice[0:len]

	// prepare ref
	ref := Ref{
		gen: gen,
		buf: buf,
	}

	return slice, ref
}

// Clone will copy the provided slice into a borrowed slice.
func Clone(slice []byte) ([]byte, Ref) {
	// borrow buffer
	buf, ref := Borrow(len(slice))

	// copy bytes
	copy(buf, slice)

	return buf, ref
}

// Concat will concatenate the provided byte slices using a borrowed slice.
func Concat(slices ...[]byte) ([]byte, Ref) {
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
