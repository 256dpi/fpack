package fpack

// Arena is a basic arena allocator that allocates fixed size buffers to provide
// memory for many small buffers. The memory must be returned all at once when
// no longer needed.
type Arena struct {
	Pool  *Pool
	Size  int
	buf   []byte
	refs  []Ref
	_refs [128]Ref
}

// Get will return a buffer of the provided length.
func (a *Arena) Get(length int) []byte {
	// check refs
	if a.refs == nil {
		a.refs = a._refs[:0]
	}

	// check size
	if length == 0 {
		return []byte{}
	} else if length > a.Size {
		buf, ref := a.Pool.Borrow(length)
		a.refs = append(a.refs, ref)
		return buf
	}

	// ensure buf
	if a.buf == nil || len(a.buf) < length {
		buf, ref := a.Pool.Borrow(a.Size)
		a.buf = buf
		a.refs = append(a.refs, ref)
	}

	// get fragment
	frag := a.buf[:length]
	a.buf = a.buf[length:]

	return frag
}

// Clone will return a copy of the provided buffer.
func (a *Arena) Clone(buf []byte) []byte {
	// clone buffer
	clone := a.Get(len(buf))
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
