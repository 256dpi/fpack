package fpack

import (
	"math"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var classes = []int{2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048}

func TestGlobal(t *testing.T) {
	assert.NotNil(t, Global())
}

func TestNoop(t *testing.T) {
	assert.NotPanics(t, func() {
		Ref{}.Release()
	})
}

func TestBorrow(t *testing.T) {
	buf, ref := Global().Borrow(123, true)
	assert.Equal(t, 123, len(buf))
	assert.Equal(t, 1024, cap(buf))
	ref.Release()

	assert.Equal(t, 0.0, testing.AllocsPerRun(100, func() {
		_, ref := Global().Borrow(123, false)
		ref.Release()
	}))

	pool := NewPool()
	assert.Equal(t, 2.0, testing.AllocsPerRun(100, func() {
		pool.Borrow(123, false)
	}))
}

func TestBorrowCapacity(t *testing.T) {
	buf, ref := Global().Borrow(7, false)
	assert.Equal(t, 7, cap(buf))
	ref.Release()

	buf, ref = Global().Borrow(77, false)
	assert.Equal(t, 1<<10, cap(buf))
	ref.Release()

	for i := 0; i < 16; i++ {
		buf, ref = Global().Borrow(777<<i, false)
		assert.Equal(t, 1<<(10+i), cap(buf))
		ref.Release()
	}

	buf, ref = Global().Borrow(777<<17, false)
	assert.Equal(t, 777<<17, cap(buf))
	ref.Release()
}

func TestDoubleRelease(t *testing.T) {
	runtime.GC()

	_, ref1 := Global().Borrow(123, false)
	ref1.Release()

	_, ref2 := Global().Borrow(123, false)
	assert.NotEqual(t, ref1, ref2)

	assert.PanicsWithValue(t, "fpack: generation mismatch", func() {
		ref1.Release()
	})

	ref2.Release()

	assert.PanicsWithValue(t, "fpack: generation mismatch", func() {
		ref2.Release()
	})
}

func TestLeakedBuffer(t *testing.T) {
	runtime.GC()

	var stack []byte
	Track(func(bytes []byte) {
		stack = bytes
	})

	_, _ = Global().Borrow(123, false)
	runtime.GC()
	assert.NotEmpty(t, stack)

	Track(nil)
	stack = nil

	_, _ = Global().Borrow(123, false)
	runtime.GC()
	assert.Empty(t, stack)
}

func TestGenerationOverflow(t *testing.T) {
	global.gen = math.MaxUint64
	_, ref := Global().Borrow(123, false)
	ref.Release()
}

func TestRefInterface(t *testing.T) {
	var _ interface {
		Release()
	} = Ref{}
}

func TestClone(t *testing.T) {
	buf, ref := Global().Clone([]byte("foo"))
	assert.Equal(t, []byte("foo"), buf)
	ref.Release()
}

func TestConcat(t *testing.T) {
	buf, ref := Global().Concat([]byte("foo"), []byte("123"), []byte("bar"))
	assert.Equal(t, []byte("foo123bar"), buf)
	ref.Release()
}

func BenchmarkBorrow(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ref := Global().Borrow(123, false)
		ref.Release()
	}
}

func BenchmarkBorrowZero(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ref := Global().Borrow(1<<16, true)
		ref.Release()
	}
}

func BenchmarkPoolClasses(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}

	for _, class := range classes {
		b.Run(strconv.Itoa(class), func(b *testing.B) {
			list := make([][]byte, b.N)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buf, ref := Global().Borrow(class, false)
				list[i] = buf
				ref.Release()
			}
		})
	}
}

func BenchmarkMakeClasses(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}

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
