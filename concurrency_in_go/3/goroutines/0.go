package main

import "fmt"

func main() {
	go sayHello()
	// continue doing other things
}

func sayHello() {
	fmt.Println("hello")
}
