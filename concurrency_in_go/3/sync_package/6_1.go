package main

import (
	"sync"
	"fmt"
)

func main() {
	var count int

	increment := func() {
		count++
	}

	decrement := func() {
		count--
	}

	var once sync.Once

	var increments sync.WaitGroup
	var decrements sync.WaitGroup
	increments.Add(100)
	decrements.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
			once.Do(decrement)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count)
}
