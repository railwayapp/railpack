package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("This is the api %s\n", runtime.Version())
}
