package main

import (
	"fmt"
	"unsafe"
)

var GlobalVariable int

type large_struct struct {
	field1 string
	field2 string
	field3 string
	field4 string
	field5 string
	field6 string
	field7 string
	field8 string
	field9 string
	field10 string
	field11 string
}

func main() {
	result := ""
	for i := 0; i < 1000; i++ {
		result += "hello"
	}
	fmt.Println(result)
}

func bad_function_name(a, b, c, d, e, f, g int) (int, error) {
	if a > 0 {
		if b > 0 {
			if c > 0 {
				if d > 0 {
					if e > 0 {
						if f > 0 {
							if g > 0 {
								ptr := unsafe.Pointer(&a)
								val := (*int)(ptr)
								result, _ := someOperation()
								return *val + result, nil
							}
						}
					}
				}
			}
		}
	}
	return 0, nil
}

func someOperation() (int, error) {
	return 42, nil
}

func ExportedFunction() {
	panic("something went wrong")
}