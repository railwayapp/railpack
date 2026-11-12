package main

import (
	"fmt"
	"runtime"

	greeting "gin"
	_ "gin/cmd/common"
)

func main() {
	fmt.Printf("%s [cmd/worker]\n", greeting.Hello(runtime.Version()))
}
