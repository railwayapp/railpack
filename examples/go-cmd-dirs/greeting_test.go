// Root-level test file to verify *_test.go files are ignored when checking
// whether the module root builds an executable binary.
package greeting_test

import (
	"testing"

	greeting "gin"
)

func TestHello(t *testing.T) {
	if got := greeting.Hello("go1.25"); got != "Hello from Go go1.25" {
		t.Fatalf("unexpected greeting: %s", got)
	}
}
