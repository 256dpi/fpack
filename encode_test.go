package fpack

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	enc := "\x01"
	enc += "\xFE"
	enc += "\xFF\xFE"
	enc += "\xFF\xFF\xFF\xFE"
	enc += "\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFE"
	enc += "\xFF"
	enc += "\xFF\xFF"
	enc += "\xFF\xFF\xFF\xFF"
	enc += "\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"
	enc += "\x7F\x7F\xFF\xFF"
	enc += "\x7F\xEF\xFF\xFF\xFF\xFF\xFF\xFF"
	enc += "\x0e"
	enc += "\x80\x04"
	enc += "\x03foo"
	enc += "\x03bar"
	enc += "\x03foo"
	enc += "\x03bar"
	enc += "baz"

	res, _, err := Encode(true, func(enc *Encoder) error {
		enc.Bool(true)
		enc.Int8(math.MaxInt8)
		enc.Int16(math.MaxInt16)
		enc.Int32(math.MaxInt32)
		enc.Int64(math.MaxInt64)
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
	assert.Equal(t, []byte(enc), res)

	assert.Equal(t, 0.0, testing.AllocsPerRun(10, func() {
		_, ref, _ := Encode(true, func(enc *Encoder) error {
			enc.Bool(true)
			enc.Int8(math.MaxInt8)
			enc.Int16(math.MaxInt16)
			enc.Int32(math.MaxInt32)
			enc.Int64(math.MaxInt64)
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

func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ref, err := Encode(true, func(enc *Encoder) error {
			enc.Bool(true)
			enc.Int8(math.MaxInt8)
			enc.Int16(math.MaxInt16)
			enc.Int32(math.MaxInt32)
			enc.Int64(math.MaxInt64)
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
