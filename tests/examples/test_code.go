package main

import "fmt"

func main() {
	var name string = "world"
	fmt.Printf("Hello %s\n", name)
}

func badFunction(a, b, c, d, e, f int) int {
	if a > 0 {
		if b > 0 {
			if c > 0 {
				if d > 0 {
					if e > 0 {
						if f > 0 {
							return a + b + c + d + e + f
						}
					}
				}
			}
		}
	}
	return 0
}

func ExportedFunctionWithoutDocs() {
	// This function lacks documentation
}