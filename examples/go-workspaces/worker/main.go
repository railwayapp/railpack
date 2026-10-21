package main

import (
	"fmt"
	"runtime"

	"example.com/shared"
)

func main() {
	fmt.Println(shared.FormatBanner("worker", runtime.Version()))
}
