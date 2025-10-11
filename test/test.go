package main

import "fmt"

func main() {
	// Traditional for loop
	for i := range 5 {
		fmt.Println("Count:", i)
	}

	// For loop as a while loop
	j := 0
	for j < 3 {
		fmt.Println("While-like count:", j)
		j++
	}

	// For-range loop for iterating over collections
	numbers := []int{10, 20, 30}
	for index, value := range numbers {
		fmt.Printf("Index: %d, Value: %d\n", index, value)
	}
}
