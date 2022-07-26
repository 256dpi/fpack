package fpack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sample = bytes.Repeat([]byte{'x'}, 128)

func TestArena(t *testing.T) {
	arena := NewArena(Global(), 64)

	buf1 := arena.Get(0, false)
	assert.Len(t, buf1, 0)
	assert.Equal(t, 0, arena.Length())

	buf2 := arena.Get(42, false)
	assert.Len(t, buf2, 42)
	assert.Equal(t, 42, arena.Length())

	buf3 := arena.Get(42, false)
	assert.Len(t, buf3, 42)
	assert.Equal(t, 84, arena.Length())

	for i := range buf2 {
		buf2[i] = 'z'
	}
	assert.Equal(t, buf3, make([]byte, 42))

	buf4 := arena.Clone(sample)
	assert.Len(t, buf4, len(sample))
	assert.Equal(t, 212, arena.Length())

	arena.Release()

	assert.Equal(t, 0.0, testing.AllocsPerRun(100, func() {
		arena = NewArena(Global(), 64)
		arena.Get(32, false)
		arena.Release()
	}))
}

func BenchmarkArena(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		arena := NewArena(Global(), 64)
		arena.Get(32, false)
		arena.Release()
	}
}
