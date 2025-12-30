package example // nolint:govet

import (
	"fmt"
	"unsafe"
)

var GlobalVariable int

type LargeStruct struct {
	Field1  string
	Field2  string
	Field3  string
	Field4  string
	Field5  string
	Field6  string
	Field7  string
	Field8  string
	Field9  string
	Field10 string
	Field11 string
}

func ExampleComplex() { // nolint:govet
	result := ""
	for i := 0; i < 1000; i++ {
		result += "hello"
	}
	fmt.Println(result)
}

func ExampleBadFunctionName(a, b, c, d, e, f, g int) (int, error) { // nolint:govet
	if a > 0 {
		if b > 0 {
			if c > 0 {
				if d > 0 {
					if e > 0 {
						if f > 0 {
							if g > 0 {
								ptr := unsafe.Pointer(&a) // nolint:staticcheck
								val := (*int)(ptr)
								result, _ := ExampleSomeOperation()
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

func ExampleSomeOperation() (int, error) { // nolint:govet
	return 42, nil
}

func ExportedFunction() {
	panic("something went wrong")
}
