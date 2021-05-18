package fpack

import "fmt"

func Example() {
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
		dec.Uint8(&num)
		dec.String(&str, 2, false)
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
}
