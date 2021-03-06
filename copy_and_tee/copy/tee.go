package main

import (
	"io"
	"os"
)

func tee() {
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

	tee1 := io.TeeReader(fileA, fileB)
	tee2 := io.TeeReader(tee1, fileC)
	tee3 := io.TeeReader(tee2, fileD)

	buf := make([]byte, 0, 4097)
	for {
		n, err := tee3.Read(buf)
		if n == 0 || err != nil {
			break
		}
	}
}
