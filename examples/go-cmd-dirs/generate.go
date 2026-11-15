//go:build ignore

// Helper script invoked via `go run generate.go` (or `go generate`).
// Declares package main at the module root, but is excluded from normal `go build`
// invocations by the ignore build constraint and must not shadow ./cmd/server.
package main

import "fmt"

func main() {
	fmt.Println("code generation helper")
}
