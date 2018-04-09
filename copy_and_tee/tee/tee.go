package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	fileA, err := os.Open("fileA.txt")
	if err != nil {
		panic(err)
	}
	defer fileA.Close()

	fileB, err := os.Create("fileB.txt")
	if err != nil {
		panic(err)
	}
	defer fileB.Close()

	tee := io.TeeReader(fileA, fileB)

	scanner := bufio.NewScanner(tee)
	for scanner.Scan() {
		fmt.Print(scanner.Text())
	}
}
