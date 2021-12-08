package fpack

import (
	"math/bits"
	"sync"
	"sync/atomic"
)

var global = NewPool()

// Global returns the global pool.
func Global() *Pool {
	return global
}

type buffer struct {
	gen   uint64
	pool  int8
	slice []byte
}

// Pool is dynamic slice length pool.
type Pool struct {
	gen   uint64
	pools []*sync.Pool
}

// NewPool creates and returns a new pool.
func NewPool() *Pool {
	// create 16 pools from 1 KB to 32 MB
	var pools []*sync.Pool
	for i := 0; i < 16; i++ {
		num := int8(i)
		size := 1 << (i + 10)
		pools = append(pools, &sync.Pool{
			New: func() interface{} {
				return &buffer{
					pool:  num,
					slice: make([]byte, size),
				}
			},
		})
	}

	return &Pool{
		pools: pools,
	}
}

var zeroRef Ref

// Ref is a reference to a borrowed slice. A zero reference represents a
// no-op reference.
type Ref struct {
	pool *Pool
	gen  uint64
	buf  *buffer
}

// Release will release the borrowed slice. The function should be called at
// most once and will panic otherwise.
func (r Ref) Release() {
	// treat zero refs as no-ops
	if r == zeroRef {
		return
	}

	// reset and check generation
	if !atomic.CompareAndSwapUint64(&r.buf.gen, r.gen, 0) {
		panic("fpack: generation mismatch")
	}

	// recycle buffer
	r.pool.pools[r.buf.pool].Put(r.buf)
}

// Borrow will return a slice that has the specified length. If the requested
// length is too small or too long a slice will be allocated. To recycle the
// slice, it must be released by calling Release() on the returned Ref value.
// Always release any returned value, even if the slice grows, it is possible
// to at least return the originally requested slice.
//
// Note: For values up to 8 bytes (64 bits) the internal Go arena allocator is
// used by calling make(). From benchmarks this seems to be faster than calling
// the pool to borrow and return a value.
func (p *Pool) Borrow(len int) ([]byte, Ref) {
	// determine pool
	pool := bits.Len64(uint64(len)) - 10
	if pool < 0 {
		pool = 0
	} else if pool >= 16 {
		pool = -1
	}

	// allocate if too small or too big
	if len < 9 || pool == -1 {
		return make([]byte, len), Ref{}
	}

	// get next non zero generation
	var gen = atomic.AddUint64(&p.gen, 1)
	if gen == 0 {
		gen = atomic.AddUint64(&p.gen, 1)
	}

	// get from pool
	buf := p.pools[pool].Get().(*buffer)

	// set generation
	buf.gen = gen

	// prepare slice
	slice := buf.slice[0:len]

	// prepare ref
	ref := Ref{
		pool: p,
		gen:  gen,
		buf:  buf,
	}

	return slice, ref
}

// Clone will copy the provided slice into a borrowed slice.
func (p *Pool) Clone(slice []byte) ([]byte, Ref) {
	// borrow buffer
	buf, ref := p.Borrow(len(slice))

	// copy bytes
	copy(buf, slice)

	return buf, ref
}

// Concat will concatenate the provided byte slices using a borrowed slice.
func (p *Pool) Concat(slices ...[]byte) ([]byte, Ref) {
	// compute total length
	var total int
	for _, s := range slices {
		total += len(s)
	}

	// borrow buffer
	buf, ref := p.Borrow(total)

	// copy bytes
	var pos int
	for _, s := range slices {
		pos += copy(buf[pos:], s)
	}

	return buf, ref
}
