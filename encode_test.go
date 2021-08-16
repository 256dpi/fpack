package fpack

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func encodeDummy(enc *Encoder) {
	enc.Skip(3)
	enc.Bool(true)
	enc.Bool(false)
	enc.Int8(math.MinInt8)
	enc.Int8(math.MaxInt8)
	enc.Int16(math.MinInt16)
	enc.Int16(math.MaxInt16)
	enc.Int32(math.MinInt32)
	enc.Int32(math.MaxInt32)
	enc.Int64(math.MinInt64)
	enc.Int64(math.MaxInt64)
	enc.Int(-42, 4)
	enc.Uint8(math.MaxUint8)
	enc.Uint16(math.MaxUint16)
	enc.Uint32(math.MaxUint32)
	enc.Uint64(math.MaxUint64)
	enc.Float32(math.MaxFloat32)
	enc.Float64(math.MaxFloat64)
	enc.VarInt(7)
	enc.VarUint(512)
	enc.String("foo")
	enc.Bytes([]byte("bar"))
	enc.FixString("foo", 1)
	enc.FixBytes([]byte("bar"), 1)
	enc.VarString("foo")
	enc.VarBytes([]byte("bar"))
	enc.DelString("foo", "\x00")
	enc.DelBytes([]byte("bar"), []byte{0})
	enc.Tail([]byte("baz"))
}

func TestMeasure(t *testing.T) {
	length, err := Measure(func(enc *Encoder) error {
		encodeDummy(enc)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, len(dummy), length)
}

func TestEncoderError(t *testing.T) {
	length, err := Measure(func(enc *Encoder) error {
		return io.EOF
	})
	assert.Equal(t, io.EOF, err)
	assert.Zero(t, length)
}

func TestMustMeasure(t *testing.T) {
	length := MustMeasure(encodeDummy)
	assert.Equal(t, len(dummy), length)
}

func TestEncode(t *testing.T) {
	withAndWithoutPool(func(pool *Pool) {
		res, _, err := Encode(pool, func(enc *Encoder) error {
			encodeDummy(enc)
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, dummy, res)
	})
}

func TestEncodeErrors(t *testing.T) {
	withAndWithoutPool(func(pool *Pool) {
		data, ref, err := Encode(pool, func(enc *Encoder) error {
			return io.EOF
		})
		assert.Equal(t, io.EOF, err)
		assert.Empty(t, data)
		assert.Zero(t, ref)

		data, ref, err = Encode(pool, func(enc *Encoder) error {
			if !enc.Counting() {
				return io.EOF
			}
			return nil
		})
		assert.Equal(t, io.EOF, err)
		assert.Empty(t, data)
		assert.Zero(t, ref)
	})
}

func TestEncodeAllocation(t *testing.T) {
	withAndWithoutPool(func(pool *Pool) {
		allocs := 0.0
		if pool == nil {
			allocs = 1.0
		}
		assert.Equal(t, allocs, testing.AllocsPerRun(10, func() {
			_, ref, _ := Encode(pool, func(enc *Encoder) error {
				encodeDummy(enc)
				return nil
			})
			ref.Release()
		}))
	})
}

func TestMustEncode(t *testing.T) {
	data, ref := MustEncode(nil, func(enc *Encoder) {
		enc.VarUint(42)
	})
	assert.Equal(t, "*", string(data))
	ref.Release()
}

func TestEncodeInto(t *testing.T) {
	n, err := EncodeInto(nil, func(enc *Encoder) error {
		enc.VarInt(42)
		return nil
	})
	assert.Equal(t, ErrBufferTooShort, err)
	assert.Zero(t, n)

	n, err = EncodeInto(make([]byte, 10), func(enc *Encoder) error {
		enc.VarUint(42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestMustEncodeInto(t *testing.T) {
	n, ok := MustEncodeInto(nil, func(enc *Encoder) {
		enc.VarUint(42)
	})
	assert.False(t, ok)
	assert.Zero(t, n)

	n, ok = MustEncodeInto(make([]byte, 10), func(enc *Encoder) {
		enc.VarUint(42)
	})
	assert.True(t, ok)
	assert.Equal(t, 1, n)
}

func TestEncodeByteOrder(t *testing.T) {
	buf, _ := MustEncode(nil, func(enc *Encoder) {
		enc.Uint16(42)
	})
	assert.Equal(t, "\x00*", string(buf))

	buf, _ = MustEncode(nil, func(enc *Encoder) {
		enc.UseLittleEndian()
		enc.Uint16(42)
	})
	assert.Equal(t, "*\x00", string(buf))
}

func TestEncodeByteOrderNegative(t *testing.T) {
	buf, _ := MustEncode(nil, func(enc *Encoder) {
		enc.Int16(42)
	})
	assert.Equal(t, "\x00*", string(buf))

	buf, _ = MustEncode(nil, func(enc *Encoder) {
		enc.UseLittleEndian()
		enc.Int16(42)
	})
	assert.Equal(t, "*\x00", string(buf))

	buf, _ = MustEncode(nil, func(enc *Encoder) {
		enc.Int16(-42)
	})
	assert.Equal(t, "\xFF\xD6", string(buf))

	buf, _ = MustEncode(nil, func(enc *Encoder) {
		enc.UseLittleEndian()
		enc.Int16(-42)
	})
	assert.Equal(t, "\xD6\xFF", string(buf))
}

func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ref, err := Encode(Global(), func(enc *Encoder) error {
			encodeDummy(enc)
			return nil
		})
		if err != nil {
			panic(err)
		}

		ref.Release()
	}
}

func withAndWithoutPool(fn func(*Pool)) {
	fn(nil)
	fn(Global())
}
