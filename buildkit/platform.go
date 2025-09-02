package buildkit

import (
	"fmt"
	"runtime"

	"github.com/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type BuildPlatform struct {
	OS           string
	Architecture string
	Variant      string
}

var (
	PlatformLinuxAMD64 = BuildPlatform{
		OS:           "linux",
		Architecture: "amd64",
	}
	PlatformLinuxARM64 = BuildPlatform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}
)

func DetermineBuildPlatformFromHost() BuildPlatform {
	if runtime.GOARCH == "arm64" {
		return PlatformLinuxARM64
	}
	return PlatformLinuxAMD64
}

func ParsePlatform(platformStr string) (BuildPlatform, error) {
	if platformStr == "" {
		return DetermineBuildPlatformFromHost(), nil
	}

	platform, err := platforms.Parse(platformStr)
	if err != nil {
		return BuildPlatform{}, fmt.Errorf("invalid platform format: %s. Must be one of: linux/amd64, linux/arm64, etc", platformStr)
	}

	return BuildPlatform{
		OS:           platform.OS,
		Architecture: platform.Architecture,
		Variant:      platform.Variant,
	}, nil
}

func (p BuildPlatform) String() string {
	if p.Variant != "" {
		return fmt.Sprintf("%s/%s/%s", p.OS, p.Architecture, p.Variant)
	}
	return fmt.Sprintf("%s/%s", p.OS, p.Architecture)
}

func (p BuildPlatform) ToPlatform() specs.Platform {
	return specs.Platform{
		OS:           p.OS,
		Architecture: p.Architecture,
		Variant:      p.Variant,
	}
}
