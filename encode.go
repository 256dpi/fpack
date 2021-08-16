package fpack

import (
	"encoding/binary"
	"math"
	"sync"
)

var encoderPool = sync.Pool{
	New: func() interface{} {
		return NewEncoder()
	},
}

// Encode will encode data using the provided encoding function. The function
// is run once to assess the length of the buffer and once to encode the data.
// Any error returned by the callback is returned immediately.
func Encode(borrow bool, fn func(enc *Encoder) error) ([]byte, Ref, error) {
	buf, _, ref, err := encode(nil, &borrow, fn)
	return buf, ref, err
}

// EncodeInto will encode data into the specified byte slice using the provided
// encoding function. The function is run once to assess the length of the
// buffer and once to encode the data. Any error returned by the callback is
// returned immediately. If the provided buffer is too small ErrBufferTooShort
// is returned.
func EncodeInto(buf []byte, fn func(enc *Encoder) error) (int, error) {
	_, n, _, err := encode(buf, nil, fn)
	return n, err
}

func encode(buf []byte, borrow *bool, fn func(enc *Encoder) error) ([]byte, int, Ref, error) {
	// borrow
	enc := encoderPool.Get().(*Encoder)

	// recycle
	defer func() {
		enc.Reset(nil)
		encoderPool.Put(enc)
	}()

	// count
	err := fn(enc)
	if err != nil {
		return nil, 0, Ref{}, err
	}

	// get length
	length := enc.Length()

	// check length
	if borrow == nil && len(buf) < length {
		return nil, 0, Ref{}, ErrBufferTooShort
	}

	// get buffer
	var ref Ref
	if borrow != nil {
		if *borrow {
			buf, ref = Borrow(length)
			buf = buf[:enc.len]
		} else {
			buf = make([]byte, length)
		}
	}

	// reset encoder
	enc.Reset(buf)

	// encode
	err = fn(enc)
	if err != nil {
		ref.Release()
		return nil, 0, Ref{}, err
	}

	return buf, length, ref, nil
}

// MustEncode wraps Encode but omits the error propagation.
func MustEncode(borrow bool, fn func(enc *Encoder)) ([]byte, Ref) {
	// encode without error
	data, ref, _ := Encode(borrow, func(enc *Encoder) error {
		fn(enc)
		return nil
	})

	return data, ref
}

// MustEncodeInto wraps EncodeInto but omits the error propagation. It will
// return false if the buffer was not long enough to write all data.
func MustEncodeInto(buf []byte, fn func(enc *Encoder)) (int, bool) {
	// encode without error
	n, err := EncodeInto(buf, func(enc *Encoder) error {
		fn(enc)
		return nil
	})

	return n, err != ErrBufferTooShort
}

// Encoder manages data encoding.
type Encoder struct {
	bo  binary.ByteOrder
	b10 [10]byte
	len int
	buf []byte
}

// NewEncoder will return an encoder.
func NewEncoder() *Encoder {
	return &Encoder{
		bo: binary.BigEndian,
	}
}

// Reset will reset the encoder. Pass nil so set the encoder to counting mode.
func (e *Encoder) Reset(buf []byte) {
	e.bo = binary.BigEndian
	e.len = 0
	e.buf = buf
}

// UseLittleEndian will set the used binary byte order to little endian.
func (e *Encoder) UseLittleEndian() {
	e.bo = binary.LittleEndian
}

// Counting returns whether the encoder is counting.
func (e *Encoder) Counting() bool {
	return e.buf == nil
}

// Length will return the accumulated length.
func (e *Encoder) Length() int {
	return e.len
}

// Skip the specified amount of bytes.
func (e *Encoder) Skip(num int) {
	// handle length
	if e.buf == nil {
		e.len += num
		return
	}

	// write zeros
	for i := 0; i < num; i++ {
		e.buf[i] = 0
	}

	// slice
	e.buf = e.buf[num:]
}

// Bool writes a boolean.
func (e *Encoder) Bool(yes bool) {
	if yes {
		e.Uint8(1)
	} else {
		e.Uint8(0)
	}
}

// Int8 writes a one byte signed integer (two's complement).
func (e *Encoder) Int8(num int8) {
	e.Int(int64(num), 1)
}

// Int16 writes a two byte signed integer (two's complement).
func (e *Encoder) Int16(num int16) {
	e.Int(int64(num), 2)
}

// Int32 writes a four byte signed integer (two's complement).
func (e *Encoder) Int32(num int32) {
	e.Int(int64(num), 4)
}

// Int64 writes an eight byte signed integer (two's complement).
func (e *Encoder) Int64(num int64) {
	e.Int(num, 8)
}

// Int writes a one, two, four or eight byte signed integer (two's complement).
func (e *Encoder) Int(n int64, size int) {
	// convert
	un := uint64(n)

	// handle length
	if e.buf == nil {
		e.len += size
		return
	}

	// write number
	switch size {
	case 1:
		e.buf[0] = uint8(un)
	case 2:
		e.bo.PutUint16(e.buf, uint16(un))
	case 4:
		e.bo.PutUint32(e.buf, uint32(un))
	case 8:
		e.bo.PutUint64(e.buf, un)
	}

	// slice
	e.buf = e.buf[size:]
}

// Uint8 writes a one byte unsigned integer.
func (e *Encoder) Uint8(num uint8) {
	e.Uint(uint64(num), 1)
}

// Uint16 writes a two byte unsigned integer.
func (e *Encoder) Uint16(num uint16) {
	e.Uint(uint64(num), 2)
}

// Uint32 writes a four byte unsigned integer.
func (e *Encoder) Uint32(num uint32) {
	e.Uint(uint64(num), 4)
}

// Uint64 writes an eight byte unsigned integer.
func (e *Encoder) Uint64(num uint64) {
	e.Uint(num, 8)
}

// Uint writes a one, two, four or eight byte unsigned integer.
func (e *Encoder) Uint(num uint64, size int) {
	// handle length
	if e.buf == nil {
		e.len += size
		return
	}

	// write number
	switch size {
	case 1:
		e.buf[0] = uint8(num)
	case 2:
		e.bo.PutUint16(e.buf, uint16(num))
	case 4:
		e.bo.PutUint32(e.buf, uint32(num))
	case 8:
		e.bo.PutUint64(e.buf, num)
	}

	// slice
	e.buf = e.buf[size:]
}

// Float32 writes a four byte float.
func (e *Encoder) Float32(num float32) {
	e.Uint32(math.Float32bits(num))
}

// Float64 writes an eight byte float.
func (e *Encoder) Float64(num float64) {
	e.Uint64(math.Float64bits(num))
}

// VarInt writes a variable signed integer.
func (e *Encoder) VarInt(num int64) {
	// handle length
	if e.buf == nil {
		e.len += binary.PutVarint(e.b10[:], num)
		return
	}

	// write number
	n := binary.PutVarint(e.buf, num)
	e.buf = e.buf[n:]
}

// VarUint writes a variable unsigned integer.
func (e *Encoder) VarUint(num uint64) {
	// handle length
	if e.buf == nil {
		e.len += binary.PutUvarint(e.b10[:], num)
		return
	}

	// write number
	n := binary.PutUvarint(e.buf, num)
	e.buf = e.buf[n:]
}

// String writes a raw string.
func (e *Encoder) String(str string) {
	// handle length
	if e.buf == nil {
		e.len += len(str)
		return
	}

	// write string
	n := copy(e.buf, str)
	e.buf = e.buf[n:]
}

// Bytes writes a raw byte slice.
func (e *Encoder) Bytes(buf []byte) {
	// handle length
	if e.buf == nil {
		e.len += len(buf)
		return
	}

	// write bytes
	n := copy(e.buf, buf)
	e.buf = e.buf[n:]
}

// FixString writes a fixed length prefixed string.
func (e *Encoder) FixString(str string, lenSize int) {
	e.Uint(uint64(len(str)), lenSize)
	e.String(str)
}

// FixBytes writes a fixed length prefixed byte slice.
func (e *Encoder) FixBytes(buf []byte, lenSize int) {
	e.Uint(uint64(len(buf)), lenSize)
	e.Bytes(buf)
}

// VarString writes a variable length prefixed string.
func (e *Encoder) VarString(str string) {
	e.VarUint(uint64(len(str)))
	e.String(str)
}

// VarBytes writes a variable length prefixed byte slice.
func (e *Encoder) VarBytes(buf []byte) {
	e.VarUint(uint64(len(buf)))
	e.Bytes(buf)
}

// DelString writes a suffix delimited string.
func (e *Encoder) DelString(str, delim string) {
	e.String(str)
	e.String(delim)
}

// DelBytes writes a suffix delimited byte slice.
func (e *Encoder) DelBytes(buf, delim []byte) {
	e.Bytes(buf)
	e.Bytes(delim)
}

// Tail writes a tail byte slice.
func (e *Encoder) Tail(buf []byte) {
	// handle length
	if e.buf == nil {
		e.len += len(buf)
		return
	}

	// write bytes
	n := copy(e.buf, buf)
	e.buf = e.buf[n:]
}
