package fpack

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	testEncode(t, true)
	testEncode(t, false)
}

func testEncode(t *testing.T, borrow bool) {
	res, _, err := Encode(borrow, func(enc *Encoder) error {
		enc.Skip(3)
		enc.Bool(true)
		enc.Bool(false)
		enc.Int8(math.MaxInt8)
		enc.Int16(math.MaxInt16)
		enc.Int32(math.MaxInt32)
		enc.Int64(math.MaxInt64)
		enc.Int64(math.MinInt64)
		enc.Uint8(math.MaxUint8)
		enc.Uint16(math.MaxUint16)
		enc.Uint32(math.MaxUint32)
		enc.Uint64(math.MaxUint64)
		enc.Float32(math.MaxFloat32)
		enc.Float64(math.MaxFloat64)
		enc.VarInt(7)
		enc.VarUint(512)
		enc.String("foo", 1)
		enc.Bytes([]byte("bar"), 1)
		enc.VarString("foo")
		enc.VarBytes([]byte("bar"))
		enc.Tail([]byte("baz"))
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, dummy, res)
}

func TestEncodeErrors(t *testing.T) {
	data, ref, err := Encode(true, func(enc *Encoder) error {
		return io.EOF
	})
	assert.Equal(t, io.EOF, err)
	assert.Empty(t, data)
	assert.Nil(t, ref)

	data, ref, err = Encode(false, func(enc *Encoder) error {
		return io.EOF
	})
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, ref)
	assert.Empty(t, data)

	data, ref, err = Encode(true, func(enc *Encoder) error {
		if !enc.Counting() {
			return io.EOF
		}
		return nil
	})
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, ref)
	assert.Empty(t, data)

	data, ref, err = Encode(false, func(enc *Encoder) error {
		if !enc.Counting() {
			return io.EOF
		}
		return nil
	})
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, ref)
	assert.Empty(t, data)
}

func TestEncodeAllocation(t *testing.T) {
	assert.Equal(t, 0.0, testing.AllocsPerRun(10, func() {
		_, ref, _ := Encode(true, func(enc *Encoder) error {
			enc.Skip(3)
			enc.Bool(true)
			enc.Bool(false)
			enc.Int8(math.MaxInt8)
			enc.Int16(math.MaxInt16)
			enc.Int32(math.MaxInt32)
			enc.Int64(math.MaxInt64)
			enc.Int64(math.MinInt64)
			enc.Uint8(math.MaxUint8)
			enc.Uint16(math.MaxUint16)
			enc.Uint32(math.MaxUint32)
			enc.Uint64(math.MaxUint64)
			enc.Float32(math.MaxFloat32)
			enc.Float64(math.MaxFloat64)
			enc.VarInt(7)
			enc.VarUint(512)
			enc.String("foo", 1)
			enc.Bytes([]byte("bar"), 1)
			enc.VarString("foo")
			enc.VarBytes([]byte("bar"))
			enc.Tail([]byte("baz"))
			return nil
		})
		ref.Release()
	}))
}

func TestMustEncode(t *testing.T) {
	data, ref := MustEncode(true, func(enc *Encoder) {
		enc.VarUint(42)
	})
	assert.Equal(t, "*", string(data))
	ref.Release()
}

func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ref, err := Encode(true, func(enc *Encoder) error {
			enc.Skip(3)
			enc.Bool(true)
			enc.Bool(false)
			enc.Int8(math.MaxInt8)
			enc.Int16(math.MaxInt16)
			enc.Int32(math.MaxInt32)
			enc.Int64(math.MaxInt64)
			enc.Int64(math.MinInt64)
			enc.Uint8(math.MaxUint8)
			enc.Uint16(math.MaxUint16)
			enc.Uint32(math.MaxUint32)
			enc.Uint64(math.MaxUint64)
			enc.Float32(math.MaxFloat32)
			enc.Float64(math.MaxFloat64)
			enc.VarInt(7)
			enc.VarUint(512)
			enc.String("foo", 1)
			enc.Bytes([]byte("bar"), 1)
			enc.VarString("foo")
			enc.VarBytes([]byte("bar"))
			enc.Tail([]byte("baz"))
			return nil
		})
		if err != nil {
			panic(err)
		}

		ref.Release()
	}
}
