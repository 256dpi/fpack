package fpack

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sample = bytes.Repeat([]byte{'x'}, 128)

func TestArena(t *testing.T) {
	arena := Arena{
		Pool: Global(),
		Size: 64,
	}

	buf1 := arena.Get(0, false)
	assert.Len(t, buf1, 0)

	buf2 := arena.Get(42, false)
	assert.Len(t, buf2, 42)

	buf3 := arena.Get(42, false)
	assert.Len(t, buf3, 42)

	for i := range buf2 {
		buf2[i] = 'z'
	}
	assert.Equal(t, buf3, make([]byte, 42))

	buf4 := arena.Clone(sample)
	assert.Len(t, buf4, len(sample))

	arena.Release()
	assert.Panics(t, func() {
		arena.Release()
	})

	assert.Equal(t, 0.0, testing.AllocsPerRun(100, func() {
		arena = Arena{
			Pool: Global(),
			Size: 64,
		}
		arena.Get(32, false)
		arena.Release()
	}))
}

func BenchmarkArena(b *testing.B) {
	b.ReportAllocs()

	var arena Arena

	for i := 0; i < b.N; i++ {
		arena = Arena{
			Pool: Global(),
			Size: 64,
		}
		arena.Get(32, false)
		arena.Release()
	}
}
