package buildkit

import (
	"github.com/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// ParsePlatformWithDefaults parses a platform string and returns the corresponding specs.Platform.
// If the input is empty, it defaults to a Linux platform that matches the host architecture.
//
// This function handles the common case where we need to map host platforms to container platforms.
// We cannot use platforms.DefaultSpec() directly because it returns the host platform
// (e.g., darwin/arm64/v8 on macOS), but container base images only support Linux platforms.
// Instead, we map the host architecture to the corresponding Linux container platform
// to provide optimal performance while ensuring compatibility with container runtimes.
//
// Examples:
//   - "" -> linux/amd64 (on Intel hosts) or linux/arm64/v8 (on ARM hosts)
//   - "linux/amd64" -> linux/amd64
//   - "linux/arm64" -> linux/arm64
//   - "linux/arm64/v8" -> linux/arm64/v8
func ParsePlatformWithDefaults(platformStr string) (specs.Platform, error) {
	if platformStr == "" {
		// Default to Linux platform for container builds
		// We cannot use platforms.DefaultSpec() directly because it returns the host platform
		// (e.g., darwin/arm64/v8 on macOS), but container base images only support Linux platforms.
		// Instead, we map the host architecture to the corresponding Linux container platform
		// to provide optimal performance while ensuring compatibility with container runtimes.
		hostPlatform := platforms.DefaultSpec()
		if hostPlatform.Architecture == "arm64" {
			return specs.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}, nil
		} else {
			return specs.Platform{OS: "linux", Architecture: "amd64"}, nil
		}
	}

	// Parse the user-specified platform string
	return platforms.Parse(platformStr)
}
