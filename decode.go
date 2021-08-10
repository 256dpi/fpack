package fpack

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"

	"github.com/tidwall/cast"
)

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

// String reads a fixed length prefixed string. If the string is not cloned it
// may change if the decoded byte slice changes.
func (d *Decoder) String(lenSize int, clone bool) string {
	return d.RawString(int(d.Uint(lenSize)), clone)
}

// Bytes reads a fixed length prefixed byte slice. If the byte slice is not
// cloned it may change if the decoded byte slice changes.
func (d *Decoder) Bytes(lenSize int, clone bool) []byte {
	return d.RawBytes(int(d.Uint(lenSize)), clone)
}

// VarString reads a variable length prefixed string. If the string is not
// cloned it may change if the decoded byte slice changes.
func (d *Decoder) VarString(clone bool) string {
	return d.RawString(int(d.VarUint()), clone)
}

// VarBytes reads a variable length prefixed byte slice. If the byte slice is
// not cloned it may change if the decoded byte slice changes.
func (d *Decoder) VarBytes(clone bool) []byte {
	return d.RawBytes(int(d.VarUint()), clone)
}

// DelimString reads a suffix delimited string. If the string is not cloned it
// may change if the decoded byte slice changes.
func (d *Decoder) DelimString(delim string, clone bool) string {
	// skip if errored
	if d.err != nil {
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

// DelimBytes reads a suffix delimited byte slice. If the byte slice is not
// cloned it may change if the decoded byte slice changes.
func (d *Decoder) DelimBytes(delim []byte, clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// find index
	idx := bytes.Index(d.buf, delim)
	if idx < 0 {
		d.err = ErrBufferTooShort
		return nil
	}

	// cast or set bytes
	var bytes []byte
	if clone {
		bytes = make([]byte, idx)
		copy(bytes, d.buf[:idx])
	} else {
		bytes = d.buf[:idx]
	}

	// slice
	d.buf = d.buf[idx+len(delim):]

	return bytes
}

// RawString reads a raw string. If the string is not cloned it may change if
// the decoded byte slice changes.
func (d *Decoder) RawString(length int, clone bool) string {
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

// RawBytes reads a raw byte slice. If the byte slice is not cloned it may
// change if the decoded byte slice changes.
func (d *Decoder) RawBytes(length int, clone bool) []byte {
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
	var bytes []byte
	if clone {
		bytes = make([]byte, length)
		copy(bytes, d.buf[:length])
		d.buf = d.buf[length:]
	} else {
		bytes = d.buf[:length]
		d.buf = d.buf[length:]
	}

	return bytes
}

// Tail reads a tail byte slice. If the byte slice is not cloned it may change
// if the decoded byte slice changes.
func (d *Decoder) Tail(clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// clone or set bytes
	var bytes []byte
	if clone {
		bytes = make([]byte, len(d.buf))
		copy(bytes, d.buf[:len(d.buf)])
		d.buf = d.buf[len(d.buf):]
	} else {
		bytes = d.buf[:len(d.buf)]
		d.buf = d.buf[len(d.buf):]
	}

	return bytes
}

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

// MustDecode wraps Decode but omits error propagation. It will return false if
// the buffer was not long enough to read all data or the buffer has not been
// fully consumed.
func MustDecode(bytes []byte, fn func(dec *Decoder)) bool {
	return Decode(bytes, func(dec *Decoder) error {
		fn(dec)
		return nil
	}) == nil
}
