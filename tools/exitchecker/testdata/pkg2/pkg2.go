package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("main with os.Exit")
}

func Exit() {
	os.Exit(1)
}
