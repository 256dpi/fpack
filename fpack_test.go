package fpack

import "fmt"

func Example() {
	// encode
	buf, ref, err := Encode(Global(), func(enc *Encoder) error {
		enc.Uint8(42)
		enc.FixString("Hello World!", 2)
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
		str = dec.FixString(2, false)
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
