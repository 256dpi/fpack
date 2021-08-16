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
	buf, ref := Borrow(123)
	assert.Equal(t, 123, len(buf))
	assert.Equal(t, 1024, cap(buf))
	ref.Release()

	assert.Equal(t, 2.0, testing.AllocsPerRun(100, func() {
		Borrow(123)
	}))

	assert.Equal(t, 0.0, testing.AllocsPerRun(100, func() {
		_, ref := Borrow(123)
		ref.Release()
	}))
}

func TestBorrowCapacity(t *testing.T) {
	buf, ref := Borrow(7)
	assert.Equal(t, 7, cap(buf))
	ref.Release()

	buf, ref = Borrow(77)
	assert.Equal(t, 1<<10, cap(buf))
	ref.Release()

	for i := 0; i < 16; i++ {
		buf, ref = Borrow(777 << i)
		assert.Equal(t, 1<<(10+i), cap(buf))
		ref.Release()
	}

	buf, ref = Borrow(777 << 17)
	assert.Equal(t, 777<<17, cap(buf))
	ref.Release()
}

func TestDoubleRelease(t *testing.T) {
	runtime.GC()

	_, ref1 := Borrow(123)
	ref1.Release()

	_, ref2 := Borrow(123)
	assert.NotEqual(t, ref1, ref2)

	assert.PanicsWithValue(t, "fpack: generation mismatch", func() {
		ref1.Release()
	})

	ref2.Release()

	assert.PanicsWithValue(t, "fpack: generation mismatch", func() {
		ref2.Release()
	})
}

func TestGenerationOverflow(t *testing.T) {
	globalPool.generation = math.MaxUint64
	_, ref := Borrow(123)
	ref.Release()
}

func TestRefInterface(t *testing.T) {
	var _ interface {
		Release()
	} = Ref{}
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
	for i := 0; i < b.N; i++ {
		_, ref := Borrow(123)
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
				buf, ref := Borrow(class)
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
