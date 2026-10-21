// Library module in the workspace imported by example.com/api via go.work
// without a replace directive in api/go.mod.
package shared

import "fmt"

func FormatBanner(service, version string) string {
	return fmt.Sprintf("This is the %s %s (from shared)", service, version)
}
