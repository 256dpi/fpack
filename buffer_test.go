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

	off, err := b.Seek(5, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), off)

	n, err = b.Write(hello[0:10])
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, 15, b.Length())

	buf = make([]byte, 10)
	n, err = b.ReadAt(buf, 5)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, hello[:10], buf)

	off, err = b.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), off)

	buf, err = io.ReadAll(b)
	assert.NoError(t, err)
	assert.Len(t, buf, 15)

	off, err = b.Seek(-2, io.SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, int64(13), off)

	buf, err = io.ReadAll(b)
	assert.NoError(t, err)
	assert.Len(t, buf, 2)

	off, err = b.Seek(-1, io.SeekStart)
	assert.Equal(t, ErrInvalidOffset, err)

	_, err = b.WriteAt(hello, 22)
	assert.NoError(t, err)
	assert.Equal(t, 34, b.Length())

	buf = make([]byte, 7)
	_, err = b.ReadAt(buf, 15)
	assert.NoError(t, err)
	assert.Equal(t, make([]byte, 7), buf)

	var chunks [][]byte
	b.Range(0, 11, func(offset int, data []byte) {
		chunks = append(chunks, append([]byte{byte(offset)}, data...))
	})
	assert.Equal(t, [][]byte{
		{0, 'H', 'e', 'l'},
		{3, 'l', 'o', 'H'},
		{6, 'e', 'l', 'l'},
		{9, 'o', ' '},
	}, chunks)

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
