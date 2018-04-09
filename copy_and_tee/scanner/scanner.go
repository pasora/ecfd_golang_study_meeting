package main

import (
	"bufio"
	"fmt"
	"strings"
)

func main() {
	source := "hoge fuga\nfoo bar"
	scanner := bufio.NewScanner(strings.NewReader(source))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		fmt.Print(scanner.Text() + "\n")
	}
}
