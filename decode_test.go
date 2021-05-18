package fpack

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
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

	bytes := []byte(enc)

	var bol bool
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var u64 uint64
	var f32 float32
	var f64 float64
	var vi int64
	var vu uint64
	var fs string
	var vs string
	var fb []byte
	var vb []byte
	var tail []byte
	err := Decode(bytes, func(dec *Decoder) error {
		dec.Bool(&bol)
		dec.Int8(&i8)
		dec.Int16(&i16)
		dec.Int32(&i32)
		dec.Int64(&i64)
		dec.Uint8(&u8)
		dec.Uint16(&u16)
		dec.Uint32(&u32)
		dec.Uint64(&u64)
		dec.Float32(&f32)
		dec.Float64(&f64)
		dec.VarInt(&vi)
		dec.VarUint(&vu)
		dec.String(&fs, 1, false)
		dec.Bytes(&fb, 1, false)
		dec.VarString(&vs, false)
		dec.VarBytes(&vb, false)
		dec.Tail(&tail, false)
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, bol)
	assert.Equal(t, int8(math.MaxInt8), i8)
	assert.Equal(t, int16(math.MaxInt16), i16)
	assert.Equal(t, int32(math.MaxInt32), i32)
	assert.Equal(t, int64(math.MaxInt64), i64)
	assert.Equal(t, uint8(math.MaxUint8), u8)
	assert.Equal(t, uint16(math.MaxUint16), u16)
	assert.Equal(t, uint32(math.MaxUint32), u32)
	assert.Equal(t, uint64(math.MaxUint64), u64)
	assert.Equal(t, int64(7), vi)
	assert.Equal(t, uint64(512), vu)
	assert.Equal(t, "foo", fs)
	assert.Equal(t, []byte("bar"), fb)
	assert.Equal(t, "foo", vs)
	assert.Equal(t, []byte("bar"), vb)
	assert.Equal(t, []byte("baz"), tail)

	assert.Equal(t, 0.0, testing.AllocsPerRun(10, func() {
		var bol bool
		var i8 int8
		var i16 int16
		var i32 int32
		var i64 int64
		var u8 uint8
		var u16 uint16
		var u32 uint32
		var u64 uint64
		var f32 float32
		var f64 float64
		var vi int64
		var vu uint64
		var fs string
		var vs string
		var fb []byte
		var vb []byte
		var tail []byte
		_ = Decode(bytes, func(dec *Decoder) error {
			dec.Bool(&bol)
			dec.Int8(&i8)
			dec.Int16(&i16)
			dec.Int32(&i32)
			dec.Int64(&i64)
			dec.Uint8(&u8)
			dec.Uint16(&u16)
			dec.Uint32(&u32)
			dec.Uint64(&u64)
			dec.Float32(&f32)
			dec.Float64(&f64)
			dec.VarInt(&vi)
			dec.VarUint(&vu)
			dec.String(&fs, 1, false)
			dec.Bytes(&fb, 1, false)
			dec.VarString(&vs, false)
			dec.VarBytes(&vb, false)
			dec.Tail(&tail, false)
			return nil
		})
	}))
}

func BenchmarkDecode(b *testing.B) {
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

	bytes := []byte(enc)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var bol bool
		var i8 int8
		var i16 int16
		var i32 int32
		var i64 int64
		var u8 uint8
		var u16 uint16
		var u32 uint32
		var u64 uint64
		var f32 float32
		var f64 float64
		var vi int64
		var vu uint64
		var fs string
		var vs string
		var fb []byte
		var vb []byte
		var tail []byte
		err := Decode(bytes, func(dec *Decoder) error {
			dec.Bool(&bol)
			dec.Int8(&i8)
			dec.Int16(&i16)
			dec.Int32(&i32)
			dec.Int64(&i64)
			dec.Uint8(&u8)
			dec.Uint16(&u16)
			dec.Uint32(&u32)
			dec.Uint64(&u64)
			dec.Float32(&f32)
			dec.Float64(&f64)
			dec.VarInt(&vi)
			dec.VarUint(&vu)
			dec.String(&fs, 1, false)
			dec.VarString(&vs, false)
			dec.Bytes(&fb, 1, false)
			dec.VarBytes(&vb, false)
			dec.Tail(&tail, false)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
