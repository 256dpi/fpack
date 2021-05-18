# fpack

[![Test](https://github.com/256dpi/fpack/actions/workflows/test.yml/badge.svg)](https://github.com/256dpi/fpack/actions/workflows/test.yml)

**A functional approach to encoding and decoding byte sequences.**

## Example

```go
// encode
buf, ref, err := Encode(true, func(enc *Encoder) error {
    enc.Uint8(42)
    enc.String("Hello World!", 2)
    return nil
})
if err != nil {
    panic(err)
}

// ensure release
defer ref.Release()

// decode
var num uint8
var str string
err = Decode(buf, func(dec *Decoder) error {
    num = dec.Uint8()
    str = dec.String(2, false)
    return nil
})
if err != nil {
    panic(err)
}

// print
fmt.Println(len(buf))
fmt.Println(num)
fmt.Println(str)

// Output:
// 15
// 42
// Hello World!
```
