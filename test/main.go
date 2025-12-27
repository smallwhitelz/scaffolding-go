package main

import "fmt"

func scan(vals ...any) {
	for i, val := range vals {
		fmt.Printf("%d: (%T) %v\n", i, val, val)
	}
}

func main() {
	vals := make([]any, 0, 4)
	vals = append(vals, 1)
	vals = append(vals, "TOM")
	vals = append(vals, false)
	vals = append(vals, 32.67)
	scan(vals)
	scan(vals...)
}
