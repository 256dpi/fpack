package fpack

import (
	"errors"
	"io"
	"sync"
)

// ErrInvalidOffset is return for offsets that under or overflow the buffer.
var ErrInvalidOffset = errors.New("invalid offset")

type chunk struct {
	buf []byte
	ref Ref
}

// Buffer is basic buffer that dynamically allocates needed chunks.
type Buffer struct {
	pool    *Pool
	alloc   int
	offset  int
	length  int
	chunks  []chunk
	_chunks [128]chunk
	mutex   sync.Mutex
}

var bufferPool = sync.Pool{
	New: func() any {
		return &Buffer{}
	},
}

// NewBuffer will return a new buffer that uses the provided pool and allocation
// size to dynamically allocate chunks as needed to hold the data.
func NewBuffer(pool *Pool, alloc int) *Buffer {
	// get buffer
	b := bufferPool.Get().(*Buffer)

	// prepare buffer
	b.pool = pool
	b.alloc = alloc

	// set chunks
	b.chunks = b._chunks[:0]

	return b
}

// Length returns the buffer length.
func (b *Buffer) Length() int {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.length
}

// Seek implements the io.Seeker interface.
func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// apply seek
	var newOffset int
	switch whence {
	case io.SeekStart:
		newOffset = int(offset)
	case io.SeekCurrent:
		newOffset = b.offset + int(offset)
	case io.SeekEnd:
		newOffset = b.length + int(offset)
	}
	if newOffset < 0 {
		return 0, ErrInvalidOffset
	}

	// set offset
	b.offset = newOffset

	return int64(b.offset), nil
}

// Write implements the io.Writer interface.
func (b *Buffer) Write(buf []byte) (int, error) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// write data
	err := b.write(b.offset, buf)
	if err != nil {
		return 0, err
	}

	// adjust offset
	b.offset += len(buf)

	return len(buf), nil
}

// WriteAt implements the io.WriterAt interface.
func (b *Buffer) WriteAt(buf []byte, off int64) (int, error) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// write data
	err := b.write(int(off), buf)
	if err != nil {
		return 0, err
	}

	return len(buf), nil
}

// Read implements the io.Reader interface.
func (b *Buffer) Read(buf []byte) (int, error) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// read data
	n, err := b.read(b.offset, buf)
	if err != nil {
		return 0, err
	}

	// adjust offset
	b.offset += n

	return n, nil
}

// ReadAt implements the io.ReaderAt interface.
func (b *Buffer) ReadAt(buf []byte, off int64) (int, error) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// read data
	n, err := b.read(int(off), buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Range will iterate over the buffer in the given range and call the provided
// function with the offset and data for each chunk.
func (b *Buffer) Range(offset, length int, fn func(offset int, data []byte)) {
	// acquire mutex
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// check offset
	if offset < 0 {
		return
	}

	// limit length
	if offset+length > b.length {
		length = b.length - offset
	}

	// iterate
	b.iterate(offset, offset+length, fn)
}

// Release will release the buffer and all memory.
func (b *Buffer) Release() {
	// release refs
	for _, chunk := range b.chunks {
		chunk.ref.Release()
	}

	// recycle buffer
	*b = Buffer{}
	bufferPool.Put(b)
}

func (b *Buffer) write(off int, buf []byte) error {
	// check offset
	if off < 0 {
		return ErrInvalidOffset
	}

	// get length
	length := b.length

	// grow buffer
	b.grow(off + len(buf))

	// zero gap
	b.iterate(length, off, func(_ int, chunk []byte) {
		for i := range chunk {
			chunk[i] = 0
		}
	})

	// write data
	b.iterate(off, off+len(buf), func(loc int, chunk []byte) {
		copy(chunk, buf[loc:])
	})

	return nil
}

func (b *Buffer) read(off int, buf []byte) (int, error) {
	// check offset
	if off < 0 {
		return 0, ErrInvalidOffset
	} else if off >= b.length {
		return 0, io.EOF
	}

	// limit read
	if off+len(buf) > b.length {
		buf = buf[:b.length-off]
	}

	// read data
	b.iterate(off, off+len(buf), func(loc int, chunk []byte) {
		copy(buf[loc:], chunk)
	})

	return len(buf), nil
}

func (b *Buffer) grow(length int) {
	// check length
	if length <= b.length {
		return
	}

	// determine blocks
	n := (length / b.alloc) + 1 - len(b.chunks)

	// append chunks
	for i := 0; i < n; i++ {
		buf, ref := b.pool.Borrow(b.alloc, false)
		b.chunks = append(b.chunks, chunk{
			buf: buf,
			ref: ref,
		})
	}

	// adjust length
	if length > b.length {
		b.length = length
	}
}

func (b *Buffer) iterate(start, end int, fn func(loc int, chunk []byte)) {
	// range over chunks
	for pos := start; pos < end; {
		// determine index and position
		idx := pos / b.alloc
		off := pos % b.alloc

		// get chunk
		chunk := b.chunks[idx]

		// get part
		part := chunk.buf[off:]

		// limit part
		if len(part) > end-pos {
			part = part[:end-pos]
		}

		// yield part
		fn(pos-start, part)

		// increment
		idx++
		pos += len(part)
	}
}
