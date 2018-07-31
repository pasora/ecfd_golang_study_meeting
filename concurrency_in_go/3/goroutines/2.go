package main

import "fmt"

func main() {
	sayHello := func() {
		fmt.Println("hello")
	}
	go sayHello()
}
