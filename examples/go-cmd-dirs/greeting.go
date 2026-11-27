/*
Package greeting is a root-level library imported by cmd/server.

The usage snippet below intentionally starts a line with "package main" inside a
block comment so package detection must parse Go syntax rather than grepping lines:

package main

func main() {
	println(greeting.Hello("go1.25"))
}
*/
package greeting

import "fmt"

func Hello(version string) string {
	return fmt.Sprintf("Hello from Go %s", version)
}
