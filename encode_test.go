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
	enc.TimeUnix(now)
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

func TestMeasureErrors(t *testing.T) {
	length, err := Measure(func(enc *Encoder) error {
		return io.EOF
	})
	assert.Equal(t, io.EOF, err)
	assert.Zero(t, length)

	table := []func(*Encoder){
		func(enc *Encoder) {
			enc.Skip(0)
		},
		func(enc *Encoder) {
			enc.Int(0, 0)
		},
		func(enc *Encoder) {
			enc.Uint(0, 0)
		},
		func(enc *Encoder) {
			enc.VarInt(0)
		},
		func(enc *Encoder) {
			enc.VarUint(0)
		},
		func(enc *Encoder) {
			enc.FixString("", 0)
		},
		func(enc *Encoder) {
			enc.FixBytes(nil, 0)
		},
		func(enc *Encoder) {
			enc.VarString("")
		},
		func(enc *Encoder) {
			enc.VarBytes(nil)
		},
		func(enc *Encoder) {
			enc.DelString("", "")
		},
		func(enc *Encoder) {
			enc.DelBytes(nil, nil)
		},
		func(enc *Encoder) {
			enc.Tail(nil)
		},
	}

	for _, item := range table {
		n, err := Measure(func(enc *Encoder) error {
			enc.err = io.EOF
			item(enc)
			return nil
		})
		assert.Equal(t, io.EOF, err)
		assert.Zero(t, n)
	}
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

func TestEncodeNumbers(t *testing.T) {
	table := []func(*Encoder){
		func(enc *Encoder) {
			enc.Int(math.MinInt8, 1)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt8, 1)
		},
		func(enc *Encoder) {
			enc.Int(math.MinInt16, 2)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt16, 2)
		},
		func(enc *Encoder) {
			enc.Int(math.MinInt32, 4)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt32, 4)
		},
		func(enc *Encoder) {
			enc.Int(math.MinInt64, 8)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt64, 8)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint8, 1)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint16, 2)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint32, 4)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint64, 8)
		},
	}

	for _, item := range table {
		data, _, err := Encode(nil, func(enc *Encoder) error {
			item(enc)
			return nil
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	}
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

		table := []func(*Encoder){
			func(enc *Encoder) {
				enc.Skip(0)
			},
			func(enc *Encoder) {
				enc.Int(0, 0)
			},
			func(enc *Encoder) {
				enc.Uint(0, 0)
			},
			func(enc *Encoder) {
				enc.VarInt(0)
			},
			func(enc *Encoder) {
				enc.VarUint(0)
			},
			func(enc *Encoder) {
				enc.FixString("", 0)
			},
			func(enc *Encoder) {
				enc.FixBytes(nil, 0)
			},
			func(enc *Encoder) {
				enc.VarString("")
			},
			func(enc *Encoder) {
				enc.VarBytes(nil)
			},
			func(enc *Encoder) {
				enc.DelString("", "")
			},
			func(enc *Encoder) {
				enc.DelBytes(nil, nil)
			},
			func(enc *Encoder) {
				enc.Tail(nil)
			},
		}

		for _, item := range table {
			data, ref, err := Encode(pool, func(enc *Encoder) error {
				enc.err = io.EOF
				item(enc)
				return nil
			})
			assert.Equal(t, io.EOF, err)
			assert.Empty(t, data)
			assert.Zero(t, ref)
		}

		data, ref, err = Encode(pool, func(enc *Encoder) error {
			if !enc.Counting() {
				enc.err = io.EOF
			}
			return nil
		})
		assert.Equal(t, io.EOF, err)
		assert.Empty(t, data)
		assert.Zero(t, ref)
	})
}

func TestEncodeNumberOverflow(t *testing.T) {
	table := []func(*Encoder){
		func(enc *Encoder) {
			enc.Int(math.MinInt8-1, 1)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt8+1, 1)
		},
		func(enc *Encoder) {
			enc.Int(math.MinInt16-1, 2)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt16+1, 2)
		},
		func(enc *Encoder) {
			enc.Int(math.MinInt32-1, 4)
		},
		func(enc *Encoder) {
			enc.Int(math.MaxInt32+1, 4)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint8+1, 1)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint16+1, 2)
		},
		func(enc *Encoder) {
			enc.Uint(math.MaxUint32+1, 4)
		},
	}

	for i, item := range table {
		data, ref, err := Encode(nil, func(enc *Encoder) error {
			item(enc)
			return nil
		})
		assert.Error(t, err, i)
		assert.Equal(t, ErrNumberOverflow, err)
		assert.Empty(t, data)
		assert.Zero(t, ref)
	}
}

func TestEncodeInvalidSize(t *testing.T) {
	data, ref, err := Encode(nil, func(enc *Encoder) error {
		enc.Int(0, 3)
		return nil
	})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Zero(t, ref)
	assert.Equal(t, ErrInvalidSize, err)

	data, ref, err = Encode(nil, func(enc *Encoder) error {
		enc.Uint(0, 3)
		return nil
	})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Zero(t, ref)
	assert.Equal(t, ErrInvalidSize, err)
}

func TestEncodeEmptyDelimiter(t *testing.T) {
	data, ref, err := Encode(nil, func(enc *Encoder) error {
		enc.DelString("", "")
		return nil
	})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Zero(t, ref)
	assert.Equal(t, ErrEmptyDelimiter, err)

	data, ref, err = Encode(nil, func(enc *Encoder) error {
		enc.DelBytes(nil, nil)
		return nil
	})
	assert.Error(t, err)
	assert.Empty(t, data)
	assert.Zero(t, ref)
	assert.Equal(t, ErrEmptyDelimiter, err)
}

func TestEncodeAllocation(t *testing.T) {
	withAndWithoutPool(func(pool *Pool) {
		allocs := 0.0
		if pool == nil {
			allocs = 1.0
		}
		assert.Equal(t, allocs, testing.AllocsPerRun(10, func() {
			_, ref, err := Encode(pool, func(enc *Encoder) error {
				encodeDummy(enc)
				return nil
			})
			if err != nil {
				panic(err)
			}
			ref.Release()
		}))
	})
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

func TestEncodeByteOrder(t *testing.T) {
	buf, _, err := Encode(nil, func(enc *Encoder) error {
		enc.Uint16(42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "\x00*", string(buf))

	buf, _, err = Encode(nil, func(enc *Encoder) error {
		enc.UseLittleEndian()
		enc.Uint16(42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "*\x00", string(buf))
}

func TestEncodeByteOrderNegative(t *testing.T) {
	buf, _, err := Encode(nil, func(enc *Encoder) error {
		enc.Int16(42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "\x00*", string(buf))

	buf, _, err = Encode(nil, func(enc *Encoder) error {
		enc.UseLittleEndian()
		enc.Int16(42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "*\x00", string(buf))

	buf, _, err = Encode(nil, func(enc *Encoder) error {
		enc.Int16(-42)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "\xFF\xD6", string(buf))

	buf, _, err = Encode(nil, func(enc *Encoder) error {
		enc.UseLittleEndian()
		enc.Int16(-42)
		return nil
	})
	assert.NoError(t, err)
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
