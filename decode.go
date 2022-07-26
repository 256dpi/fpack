package fpack

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"

	"github.com/tidwall/cast"
)

var decoderPool = sync.Pool{
	New: func() interface{} {
		return NewDecoder(nil)
	},
}

// Decode will decode data using the provided decoding function. The function is
// run once to decode the data. It will return ErrBufferTooShort if the buffer
// was not long enough to read all data, ErrRemainingBytes if the provided
// buffers has not been full consumed or any error returned by the callback.
func Decode(bytes []byte, fn func(dec *Decoder) error) error {
	// borrow
	dec := decoderPool.Get().(*Decoder)
	dec.Reset(bytes)

	// recycle
	defer func() {
		dec.Reset(nil)
		decoderPool.Put(dec)
	}()

	// decode
	err := fn(dec)
	if err != nil {
		return err
	}

	// check error
	err = dec.Error()
	if err != nil {
		return err
	}

	// check length
	if dec.Length() != 0 {
		return ErrRemainingBytes
	}

	return nil
}

// Decoder manages data decoding.
type Decoder struct {
	bo  binary.ByteOrder
	buf []byte
	err error
}

// NewDecoder will return a new decoder.
func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		bo:  binary.BigEndian,
		buf: buf,
	}
}

// Reset will reset the decoder.
func (d *Decoder) Reset(buf []byte) {
	d.bo = binary.BigEndian
	d.buf = buf
	d.err = nil
}

// UseLittleEndian will set the used binary byte order to little endian.
func (d *Decoder) UseLittleEndian() {
	d.bo = binary.LittleEndian
}

// Length returns the remaining length of the buffer.
func (d *Decoder) Length() int {
	return len(d.buf)
}

// Error will return the current error.
func (d *Decoder) Error() error {
	return d.err
}

// Remaining returns whether more bytes can be decoded.
func (d *Decoder) Remaining() bool {
	return len(d.buf) > 0 && d.err == nil
}

// Skip the specified amount of bytes.
func (d *Decoder) Skip(num int) {
	// skip if errored
	if d.err != nil {
		return
	}

	// check length
	if len(d.buf) < num {
		d.err = ErrBufferTooShort
		return
	}

	// slice
	d.buf = d.buf[num:]
}

// Bool reads a boolean.
func (d *Decoder) Bool() bool {
	return d.Uint8() == 1
}

// Int8 reads a one byte signed integer (two's complement).
func (d *Decoder) Int8() int8 {
	return int8(d.Int(1))
}

// Int16 reads a two byte signed integer (two's complement).
func (d *Decoder) Int16() int16 {
	return int16(d.Int(2))
}

// Int32 reads a four byte signed integer (two's complement).
func (d *Decoder) Int32() int32 {
	return int32(d.Int(4))
}

// Int64 reads an eight byte signed integer (two's complement).
func (d *Decoder) Int64() int64 {
	return d.Int(8)
}

// Int read a one, two, four or eight byte signed integer (two's complement).
func (d *Decoder) Int(size int) int64 {
	// skip if errored
	if d.err != nil {
		return 0
	}

	// check length
	if len(d.buf) < size {
		d.err = ErrBufferTooShort
		return 0
	}

	// read and convert
	var i int64
	switch size {
	case 1:
		i = int64(int8(d.buf[0]))
	case 2:
		i = int64(int16(d.bo.Uint16(d.buf)))
	case 4:
		i = int64(int32(d.bo.Uint32(d.buf)))
	case 8:
		i = int64(d.bo.Uint64(d.buf))
	default:
		d.err = ErrInvalidSize
		return 0
	}

	// slice
	d.buf = d.buf[size:]

	return i
}

// Uint8 reads a one byte unsigned integer.
func (d *Decoder) Uint8() uint8 {
	return uint8(d.Uint(1))
}

// Uint16 reads a two byte unsigned integer.
func (d *Decoder) Uint16() uint16 {
	return uint16(d.Uint(2))
}

// Uint32 reads a four byte unsigned integer.
func (d *Decoder) Uint32() uint32 {
	return uint32(d.Uint(4))
}

// Uint64 reads an eight byte unsigned integer.
func (d *Decoder) Uint64() uint64 {
	return d.Uint(8)
}

// Uint reads a one, two, four or eight byte unsigned integer.
func (d *Decoder) Uint(size int) uint64 {
	// skip if errored
	if d.err != nil {
		return 0
	}

	// check length
	if len(d.buf) < size {
		d.err = ErrBufferTooShort
		return 0
	}

	// read
	var u uint64
	switch size {
	case 1:
		u = uint64(d.buf[0])
	case 2:
		u = uint64(d.bo.Uint16(d.buf))
	case 4:
		u = uint64(d.bo.Uint32(d.buf))
	case 8:
		u = d.bo.Uint64(d.buf)
	default:
		d.err = ErrInvalidSize
		return 0
	}

	// slice
	d.buf = d.buf[size:]

	return u
}

// Float32 reads a four byte float.
func (d *Decoder) Float32() float32 {
	return math.Float32frombits(d.Uint32())
}

// Float64 reads an eight byte float.
func (d *Decoder) Float64() float64 {
	return math.Float64frombits(d.Uint64())
}

// VarUint reads a variable unsigned integer.
func (d *Decoder) VarUint() uint64 {
	// skip if errored
	if d.err != nil {
		return 0
	}

	// read
	num, n := binary.Uvarint(d.buf)
	if n <= 0 {
		d.err = ErrBufferTooShort
		return 0
	}

	// slice
	d.buf = d.buf[n:]

	return num
}

// VarInt reads a variable signed integer.
func (d *Decoder) VarInt() int64 {
	// skip if errored
	if d.err != nil {
		return 0
	}

	// read
	num, n := binary.Varint(d.buf)
	if n <= 0 {
		d.err = ErrBufferTooShort
		return 0
	}

	// slice
	d.buf = d.buf[n:]

	return num
}

// String reads a raw string. If the string is not cloned it may change if
// the source byte slice changes.
func (d *Decoder) String(length int, clone bool) string {
	// skip if errored
	if d.err != nil {
		return ""
	}

	// check length
	if len(d.buf) < length {
		d.err = ErrBufferTooShort
		return ""
	}

	// cast or set string
	var str string
	if clone {
		str = string(d.buf[:length])
		d.buf = d.buf[length:]
	} else {
		str = cast.ToString(d.buf[:length])
		d.buf = d.buf[length:]
	}

	return str
}

// Bytes reads a raw byte slice. If the byte slice is not cloned it may
// change if the source byte slice changes.
func (d *Decoder) Bytes(length int, clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// check length
	if len(d.buf) < length {
		d.err = ErrBufferTooShort
		return nil
	}

	// clone or set bytes
	var buf []byte
	if clone {
		buf = make([]byte, length)
		copy(buf, d.buf[:length])
		d.buf = d.buf[length:]
	} else {
		buf = d.buf[:length]
		d.buf = d.buf[length:]
	}

	return buf
}

// FixString reads a fixed length prefixed string. If the string is not cloned it
// may change if the source byte slice changes.
func (d *Decoder) FixString(lenSize int, clone bool) string {
	return d.String(int(d.Uint(lenSize)), clone)
}

// FixBytes reads a fixed length prefixed byte slice. If the byte slice is not
// cloned it may change if the source byte slice changes.
func (d *Decoder) FixBytes(lenSize int, clone bool) []byte {
	return d.Bytes(int(d.Uint(lenSize)), clone)
}

// VarString reads a variable length prefixed string. If the string is not
// cloned it may change if the source byte slice changes.
func (d *Decoder) VarString(clone bool) string {
	return d.String(int(d.VarUint()), clone)
}

// VarBytes reads a variable length prefixed byte slice. If the byte slice is
// not cloned it may change if the source byte slice changes.
func (d *Decoder) VarBytes(clone bool) []byte {
	return d.Bytes(int(d.VarUint()), clone)
}

// DelString reads a suffix delimited string. If the string is not cloned it
// may change if the source byte slice changes.
func (d *Decoder) DelString(delim string, clone bool) string {
	// skip if errored
	if d.err != nil {
		return ""
	}

	// check delimiter
	if len(delim) == 0 {
		d.err = ErrEmptyDelimiter
		return ""
	}

	// find index
	idx := bytes.Index(d.buf, cast.ToBytes(delim))
	if idx < 0 {
		d.err = ErrBufferTooShort
		return ""
	}

	// cast or set string
	var str string
	if clone {
		str = string(d.buf[:idx])
	} else {
		str = cast.ToString(d.buf[:idx])
	}

	// slice
	d.buf = d.buf[idx+len(delim):]

	return str
}

// DelBytes reads a suffix delimited byte slice. If the byte slice is not
// cloned it may change if the source byte slice changes.
func (d *Decoder) DelBytes(delim []byte, clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// check delimiter
	if len(delim) == 0 {
		d.err = ErrEmptyDelimiter
		return nil
	}

	// find index
	idx := bytes.Index(d.buf, delim)
	if idx < 0 {
		d.err = ErrBufferTooShort
		return nil
	}

	// cast or set bytes
	var buf []byte
	if clone {
		buf = make([]byte, idx)
		copy(buf, d.buf[:idx])
	} else {
		buf = d.buf[:idx]
	}

	// slice
	d.buf = d.buf[idx+len(delim):]

	return buf
}

// Tail reads a tail byte slice. If the byte slice is not cloned it may change
// if the source byte slice changes.
func (d *Decoder) Tail(clone bool) []byte {
	return d.Bytes(len(d.buf), clone)
}
