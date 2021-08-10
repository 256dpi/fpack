package fpack

import (
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	testDecode(t, false)
	testDecode(t, true)
}

func testDecode(t *testing.T, clone bool) {
	var bt bool
	var bf bool
	var ni8 int8
	var i8 int8
	var ni16 int16
	var i16 int16
	var ni32 int32
	var i32 int32
	var ni64 int64
	var i64 int64
	var ni int64
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var u64 uint64
	var f32 float32
	var f64 float64
	var vi int64
	var vu uint64
	var fs string
	var fb []byte
	var vs string
	var vb []byte
	var ds string
	var db []byte
	var rs string
	var rb []byte
	var tail []byte
	err := Decode(dummy, func(dec *Decoder) error {
		dec.Skip(3)
		bt = dec.Bool()
		bf = dec.Bool()
		ni8 = dec.Int8()
		i8 = dec.Int8()
		ni16 = dec.Int16()
		i16 = dec.Int16()
		ni32 = dec.Int32()
		i32 = dec.Int32()
		ni64 = dec.Int64()
		i64 = dec.Int64()
		ni = dec.Int(4)
		u8 = dec.Uint8()
		u16 = dec.Uint16()
		u32 = dec.Uint32()
		u64 = dec.Uint64()
		f32 = dec.Float32()
		f64 = dec.Float64()
		vi = dec.VarInt()
		vu = dec.VarUint()
		fs = dec.String(1, clone)
		fb = dec.Bytes(1, clone)
		vs = dec.VarString(clone)
		vb = dec.VarBytes(clone)
		ds = dec.DelimString("\x00", clone)
		db = dec.DelimBytes([]byte{0}, clone)
		rs = dec.RawString(3, clone)
		rb = dec.RawBytes(3, clone)
		tail = dec.Tail(clone)
		return nil
	})
	assert.NoError(t, err)
	assert.True(t, bt)
	assert.False(t, bf)
	assert.Equal(t, int8(math.MinInt8), ni8)
	assert.Equal(t, int8(math.MaxInt8), i8)
	assert.Equal(t, int16(math.MinInt16), ni16)
	assert.Equal(t, int16(math.MaxInt16), i16)
	assert.Equal(t, int32(math.MinInt32), ni32)
	assert.Equal(t, int32(math.MaxInt32), i32)
	assert.Equal(t, int64(math.MinInt64), ni64)
	assert.Equal(t, int64(math.MaxInt64), i64)
	assert.Equal(t, int64(-42), ni)
	assert.Equal(t, uint8(math.MaxUint8), u8)
	assert.Equal(t, uint16(math.MaxUint16), u16)
	assert.Equal(t, uint32(math.MaxUint32), u32)
	assert.Equal(t, uint64(math.MaxUint64), u64)
	assert.Equal(t, float32(math.MaxFloat32), f32)
	assert.Equal(t, math.MaxFloat64, f64)
	assert.Equal(t, int64(7), vi)
	assert.Equal(t, uint64(512), vu)
	assert.Equal(t, "foo", fs)
	assert.Equal(t, []byte("bar"), fb)
	assert.Equal(t, "foo", vs)
	assert.Equal(t, []byte("bar"), vb)
	assert.Equal(t, "foo", ds)
	assert.Equal(t, []byte("bar"), db)
	assert.Equal(t, "foo", rs)
	assert.Equal(t, []byte("bar"), rb)
	assert.Equal(t, []byte("baz"), tail)
}

func TestDecodeRemaining(t *testing.T) {
	err := Decode([]byte{42, 84}, func(dec *Decoder) error {
		assert.True(t, dec.Remaining())
		dec.Uint8()
		assert.True(t, dec.Remaining())
		dec.Uint8()
		assert.False(t, dec.Remaining())
		dec.Uint8()
		assert.False(t, dec.Remaining())
		return nil
	})
	assert.Error(t, err)
}

func TestDecodeErrors(t *testing.T) {
	table := []func(*Decoder){
		func(dec *Decoder) {
			dec.Skip(3)
		},
		func(dec *Decoder) {
			dec.Int(8)
		},
		func(dec *Decoder) {
			dec.Uint(8)
		},
		func(dec *Decoder) {
			dec.VarInt()
		},
		func(dec *Decoder) {
			dec.VarUint()
		},
		func(dec *Decoder) {
			dec.String(8, true)
		},
		func(dec *Decoder) {
			dec.Bytes(8, true)
		},
		func(dec *Decoder) {
			dec.VarString(true)
		},
		func(dec *Decoder) {
			dec.VarBytes(true)
		},
		func(dec *Decoder) {
			dec.Tail(true)
		},
	}

	for _, item := range table {
		err := Decode(nil, func(dec *Decoder) error {
			dec.err = io.EOF
			item(dec)
			return nil
		})
		assert.Equal(t, io.EOF, err)
	}
}

func TestDecodeShortBuffer(t *testing.T) {
	table := []func(*Decoder){
		func(dec *Decoder) {
			dec.Skip(3)
		},
		func(dec *Decoder) {
			dec.Int(8)
		},
		func(dec *Decoder) {
			dec.Uint(8)
		},
		func(dec *Decoder) {
			dec.VarInt()
		},
		func(dec *Decoder) {
			dec.VarUint()
		},
		func(dec *Decoder) {
			dec.String(8, true)
		},
		func(dec *Decoder) {
			dec.Bytes(8, true)
		},
		func(dec *Decoder) {
			dec.VarString(true)
		},
		func(dec *Decoder) {
			dec.VarBytes(true)
		},
	}

	for i, item := range table {
		err := Decode(nil, func(dec *Decoder) error {
			item(dec)
			return nil
		})
		assert.Equal(t, ErrBufferTooShort, err, i)
	}

	table = []func(*Decoder){
		func(dec *Decoder) {
			dec.String(1, true)
		},
		func(dec *Decoder) {
			dec.Bytes(1, true)
		},
		func(dec *Decoder) {
			dec.VarString(true)
		},
		func(dec *Decoder) {
			dec.VarBytes(true)
		},
	}

	for i, item := range table {
		err := Decode([]byte{42}, func(dec *Decoder) error {
			item(dec)
			return nil
		})
		assert.Equal(t, ErrBufferTooShort, err, i)
	}
}

func TestDecodeRemainingBytes(t *testing.T) {
	err := Decode([]byte{42, 84}, func(dec *Decoder) error {
		dec.Uint8()
		return nil
	})
	assert.Equal(t, ErrRemainingBytes, err)
}

func TestDecodeAllocation(t *testing.T) {
	assert.Equal(t, 0.0, testing.AllocsPerRun(10, func() {
		_ = Decode(dummy, func(dec *Decoder) error {
			dec.Skip(3)
			dec.Bool()
			dec.Bool()
			dec.Int8()
			dec.Int16()
			dec.Int32()
			dec.Int64()
			dec.Int64()
			dec.Uint8()
			dec.Uint16()
			dec.Uint32()
			dec.Uint64()
			dec.Float32()
			dec.Float64()
			dec.VarInt()
			dec.VarUint()
			dec.String(1, false)
			dec.Bytes(1, false)
			dec.VarString(false)
			dec.VarBytes(false)
			dec.DelimString("\x00", false)
			dec.DelimBytes([]byte{0}, false)
			dec.RawString(3, false)
			dec.RawBytes(3, false)
			dec.Tail(false)
			return nil
		})
	}))
}

func TestDecodeByteOrder(t *testing.T) {
	MustDecode([]byte("\x00*"), func(dec *Decoder) {
		assert.Equal(t, uint16(42), dec.Uint16())
	})

	MustDecode([]byte("*\x00"), func(dec *Decoder) {
		dec.UseLittleEndian()
		assert.Equal(t, uint16(42), dec.Uint16())
	})
}

func TestMustDecode(t *testing.T) {
	var num uint64
	ok := MustDecode([]byte("*"), func(dec *Decoder) {
		num = dec.VarUint()
	})
	assert.True(t, ok)
	assert.Equal(t, uint64(42), num)

	ok = MustDecode([]byte(""), func(dec *Decoder) {
		num = dec.VarUint()
	})
	assert.False(t, ok)
	assert.Zero(t, num)
}

func BenchmarkDecode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := Decode(dummy, func(dec *Decoder) error {
			dec.Skip(3)
			dec.Bool()
			dec.Bool()
			dec.Int8()
			dec.Int8()
			dec.Int16()
			dec.Int16()
			dec.Int32()
			dec.Int32()
			dec.Int64()
			dec.Int64()
			dec.Int(4)
			dec.Uint8()
			dec.Uint16()
			dec.Uint32()
			dec.Uint64()
			dec.Float32()
			dec.Float64()
			dec.VarInt()
			dec.VarUint()
			dec.String(1, false)
			dec.Bytes(1, false)
			dec.VarString(false)
			dec.VarBytes(false)
			dec.DelimString("\x00", false)
			dec.DelimBytes([]byte{0}, false)
			dec.RawString(3, false)
			dec.RawBytes(3, false)
			dec.Tail(false)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
