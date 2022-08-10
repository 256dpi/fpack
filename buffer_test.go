package fpack

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var hello = []byte("Hello world!")

func TestBuffer(t *testing.T) {
	b := NewBuffer(Global(), 3)
	assert.Equal(t, 0, b.Length())

	n, err := b.WriteAt(hello, 0)
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, 12, b.Length())

	buf := make([]byte, 12)
	n, err = b.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, hello, buf)

	n, err = b.WriteAt(hello[0:10], 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, 15, b.Length())

	buf = make([]byte, 10)
	n, err = b.ReadAt(buf, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, hello[:10], buf)

	off, err := b.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), off)

	buf, err = io.ReadAll(b)
	assert.NoError(t, err)
	assert.Len(t, buf, 15)

	b.Release()
}

func BenchmarkBuffer(b *testing.B) {
	data := make([]byte, 1<<16) // 64 KiB

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b := NewBuffer(Global(), 1<<12) // 4 KiB
		n, err := b.Write(data)
		if err != nil {
			panic(err)
		} else if n != len(data) {
			panic("invalid size")
		}
		b.Release()
	}
}
