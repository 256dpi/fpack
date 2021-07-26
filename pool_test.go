package fpack

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var classes = []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048}

func TestNoop(t *testing.T) {
	assert.Equal(t, noop, Noop())
}

func TestBorrow(t *testing.T) {
	buf, ref := Borrow(123)
	assert.Equal(t, 123, len(buf))
	assert.Equal(t, maxSize, cap(buf))
	assert.False(t, ref.done)

	ref.Release()
	assert.True(t, ref.done)

	ref.Release()
	assert.True(t, ref.done)

	assert.Equal(t, 0.0, testing.AllocsPerRun(10, func() {
		_, ref := Borrow(123)
		ref.Release()
	}))
}

func TestClone(t *testing.T) {
	buf, ref := Clone([]byte("foo"))
	assert.Equal(t, []byte("foo"), buf)
	ref.Release()
}

func TestConcat(t *testing.T) {
	buf, ref := Concat([]byte("foo"), []byte("123"), []byte("bar"))
	assert.Equal(t, []byte("foo123bar"), buf)
	ref.Release()
}

func BenchmarkPool(b *testing.B) {
	for _, class := range classes {
		b.Run(strconv.Itoa(class), func(b *testing.B) {
			list := make([][]byte, b.N)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf, ref := Borrow(class)
				list[i] = buf
				ref.Release()
			}
		})
	}
}

func BenchmarkMake(b *testing.B) {
	for _, class := range classes {
		b.Run(strconv.Itoa(class), func(b *testing.B) {
			list := make([][]byte, b.N)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				list[i] = make([]byte, class)
			}
		})
	}
}
