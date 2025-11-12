package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("This is the shared package %s\n", runtime.Version())
}
