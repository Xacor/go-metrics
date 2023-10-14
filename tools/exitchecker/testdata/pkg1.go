package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("main with os.Exit")
	os.Exit(1) // want "os.Exit call in main func of main package"
}
