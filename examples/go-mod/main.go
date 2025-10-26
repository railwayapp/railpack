package main

import (
	"fmt"
	"runtime"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"
)

func main() {
	say, err := cowsay.Say(
		fmt.Sprintf("Hello from Go %s", runtime.Version()),
		cowsay.Type("default"),
		cowsay.BallonWidth(40),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(say)
}
