package fpack

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"

	"github.com/tidwall/cast"
)

// ErrBufferTooShort is returned if the provided buffer is too short.
var ErrBufferTooShort = errors.New("buffer too short")

// Decoder manages data decoding.
type Decoder struct {
	buf []byte
	err error
}

// NewDecoder will return a decoder for the provided buffer.
func NewDecoder(buf []byte) *Decoder {
	return &Decoder{
		buf: buf,
	}
}

// Reset will reset the decoder and set the provided byte slice.
func (d *Decoder) Reset(buf []byte) {
	d.buf = buf
	d.err = nil
}

// Error will return the error.
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

// Int8 reads a one byte integer.
func (d *Decoder) Int8() int8 {
	return int8(d.Int(1))
}

// Int16 reads a two byte integer.
func (d *Decoder) Int16() int16 {
	return int16(d.Int(2))
}

// Int32 reads a four byte integer.
func (d *Decoder) Int32() int32 {
	return int32(d.Int(4))
}

// Int64 reads an eight byte integer.
func (d *Decoder) Int64() int64 {
	return d.Int(8)
}

// Int read a one, two, four or eight byte integer.
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

	// read
	var u uint64
	switch size {
	case 1:
		u = uint64(d.buf[0])
	case 2:
		u = uint64(binary.BigEndian.Uint16(d.buf))
	case 4:
		u = uint64(binary.BigEndian.Uint32(d.buf))
	case 8:
		u = binary.BigEndian.Uint64(d.buf)
	}

	// slice
	d.buf = d.buf[size:]

	// convert
	i := int64(u >> 1)
	if u&1 != 0 {
		i = ^i
	}

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
		u = uint64(binary.BigEndian.Uint16(d.buf))
	case 4:
		u = uint64(binary.BigEndian.Uint32(d.buf))
	case 8:
		u = binary.BigEndian.Uint64(d.buf)
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
	if n == 0 {
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
	if n == 0 {
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
	// skip if errored
	if d.err != nil {
		return ""
	}

	// read length
	length := d.Uint(lenSize)
	if d.err != nil {
		return ""
	}

	// check length
	if len(d.buf) < int(length) {
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

// Bytes reads a fixed length prefixed byte slice. If the byte slice is not
// cloned it may change if the decoded byte slice changes.
func (d *Decoder) Bytes(lenSize int, clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// read length
	length := d.Uint(lenSize)
	if d.err != nil {
		return nil
	}

	// check length
	if len(d.buf) < int(length) {
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

// VarString reads a variable length prefixed string. If the string is not
// cloned it may change if the decoded byte slice changes.
func (d *Decoder) VarString(clone bool) string {
	// skip if errored
	if d.err != nil {
		return ""
	}

	// read length
	length, n := binary.Uvarint(d.buf)
	if n == 0 {
		d.err = ErrBufferTooShort
		return ""
	}

	// slice
	d.buf = d.buf[n:]

	// check length
	if len(d.buf) < int(length) {
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

// VarBytes reads a variable length prefixed byte slice. If the byte slice is
// not cloned it may change if the decoded byte slice changes.
func (d *Decoder) VarBytes(clone bool) []byte {
	// skip if errored
	if d.err != nil {
		return nil
	}

	// read length
	length, n := binary.Uvarint(d.buf)
	if n == 0 {
		d.err = ErrBufferTooShort
		return nil
	}

	// slice
	d.buf = d.buf[n:]

	// check length
	if len(d.buf) < int(length) {
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

// Tail reads a tail byte slice.
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
// was not long enough to read all data or any error returned by the callback.
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
	if err == nil {
		err = dec.Error()
	}

	return err
}

// MustDecode wraps Decode but omits error propagation. It will return false if
// the buffer was not long enough to read all data.
func MustDecode(bytes []byte, fn func(dec *Decoder)) bool {
	return Decode(bytes, func(dec *Decoder) error {
		fn(dec)
		return nil
	}) != ErrBufferTooShort
}
