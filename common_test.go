package fpack

import "time"

var dummy []byte

var now = time.Date(2022, 05, 7, 18, 43, 0, 0, time.UTC)

func init() {
	enc := "\x00\x00\x00"
	enc += "\x01"
	enc += "\x00"
	enc += "\x80"
	enc += "\x7F"
	enc += "\x80\x00"
	enc += "\x7F\xFF"
	enc += "\x80\x00\x00\x00"
	enc += "\x7F\xFF\xFF\xFF"
	enc += "\x80\x00\x00\x00\x00\x00\x00\x00"
	enc += "\x7F\xFF\xFF\xFF\xFF\xFF\xFF\xFF"
	enc += "\xFF\xFF\xFF\xD6"
	enc += "\xFF"
	enc += "\xFF\xFF"
	enc += "\xFF\xFF\xFF\xFF"
	enc += "\xFF\xFF\xFF\xFF\xFF\xFF\xFF\xFF"
	enc += "\x7F\x7F\xFF\xFF"
	enc += "\x7F\xEF\xFF\xFF\xFF\xFF\xFF\xFF"
	enc += "\x0e"
	enc += "\x80\x04"
	enc += "\x00\x00\x00\x00\x62\x76\xbd\xb4"
	enc += "foo"
	enc += "bar"
	enc += "\x03foo"
	enc += "\x03bar"
	enc += "\x03foo"
	enc += "\x03bar"
	enc += "foo\x00"
	enc += "bar\x00"
	enc += "baz"
	dummy = []byte(enc)
}
