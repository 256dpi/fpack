package fpack

// Arena is a basic arena allocator that allocates fixed size buffers to provide
// memory for many small buffers.
type Arena struct {
	pool  *Pool
	size  int
	len   int
	buf   []byte
	refs  []Ref
	_refs [128]Ref
}

// NewArena creates and returns a new arena using the specified pool and buffer
// size.
func NewArena(pool *Pool, size int) *Arena {
	return &Arena{
		pool: pool,
		size: size,
	}
}

// Length returns the total length of the arena.
func (a *Arena) Length() int {
	return a.len
}

// Get will return a buffer of the provided length.
func (a *Arena) Get(length int, zero bool) []byte {
	// ensure refs
	if a.refs == nil {
		a.refs = a._refs[:0]
	}

	// increment
	a.len += length

	// check size
	if length == 0 {
		return []byte{}
	} else if length > a.size {
		buf, ref := a.pool.Borrow(length, zero)
		a.refs = append(a.refs, ref)
		return buf
	}

	// ensure buf
	if a.buf == nil || len(a.buf) < length {
		buf, ref := a.pool.Borrow(a.size, false)
		a.buf = buf
		a.refs = append(a.refs, ref)
	}

	// get fragment
	frag := a.buf[:length]
	a.buf = a.buf[length:]

	// zero fragment if requested
	if zero {
		for i := range frag {
			frag[i] = 0
		}
	}

	return frag
}

// Clone will return a copy of the provided buffer.
func (a *Arena) Clone(buf []byte) []byte {
	// clone buffer
	clone := a.Get(len(buf), false)
	copy(clone, buf)

	return clone
}

// Release will release all returned buffers.
func (a *Arena) Release() {
	// release refs
	for _, ref := range a.refs {
		ref.Release()
	}
}
