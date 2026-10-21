// Shared CLI helper package under cmd/ that sorts lexicographically before
// cmd/server. Discovery of buildable commands in cmd/* must skip non-main
// packages like cmd/common and select cmd/server.
package common

import "fmt"

func FormatBanner(msg string) string {
	return fmt.Sprintf("%s [cmd/server]", msg)
}
