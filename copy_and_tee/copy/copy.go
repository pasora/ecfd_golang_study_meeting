package main

import (
	"io"
	"os"
)

func main() {
	copy()
}

func copy() {
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

	fileC, err := os.Create("fileC.txt")
	if err != nil {
		panic(err)
	}
	defer fileC.Close()

	fileD, err := os.Create("fileD.txt")
	if err != nil {
		panic(err)
	}
	defer fileD.Close()

	io.Copy(fileB, fileA)
	io.Copy(fileC, fileB)
	io.Copy(fileD, fileC)
}
