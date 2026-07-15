package buildkit

import (
	"net"
	"strings"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	"github.com/pkg/errors"
)

// Frontend opts for build networking, matching the Dockerfile frontend's dockerui
// contract so docker buildx --network / --add-host work with railpack:
//
//	https://github.com/moby/buildkit/blob/v0.31.1/frontend/dockerui/config.go
//	https://github.com/moby/buildkit/blob/v0.31.1/frontend/dockerui/attr.go
//
// buildx maps --network=host → force-network-mode=host and --add-host host:ip →
// add-hosts=host=ip. The Dockerfile frontend reads these and applies them to
// RUN via llb.Network / llb.AddExtraHost; we do the same for plan exec steps.
const (
	keyForceNetwork   = "force-network-mode"
	keyGlobalAddHosts = "add-hosts"
)

// parseNetMode mirrors dockerui.parseNetMode.
// https://github.com/moby/buildkit/blob/v0.31.1/frontend/dockerui/attr.go
func parseNetMode(v string) (pb.NetMode, error) {
	if v == "" {
		return llb.NetModeSandbox, nil
	}
	switch v {
	case "none":
		return llb.NetModeNone, nil
	case "host":
		return llb.NetModeHost, nil
	case "sandbox":
		return llb.NetModeSandbox, nil
	default:
		return 0, errors.Errorf("invalid netmode %s", v)
	}
}

// parseExtraHosts mirrors dockerui.parseExtraHosts (CSV of host=ip pairs).
// https://github.com/moby/buildkit/blob/v0.31.1/frontend/dockerui/attr.go
func parseExtraHosts(v string) ([]llb.HostIP, error) {
	if v == "" {
		return nil, nil
	}

	out := make([]llb.HostIP, 0)
	for field := range strings.SplitSeq(v, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		// dockerui lowercases the field before splitting so hostnames are
		// normalized the same way buildx/Dockerfile would.
		key, val, ok := strings.Cut(strings.ToLower(field), "=")
		if !ok {
			return nil, errors.Errorf("invalid key-value pair %s", field)
		}
		ip := net.ParseIP(val)
		if ip == nil {
			return nil, errors.Errorf("failed to parse IP %s", val)
		}
		out = append(out, llb.HostIP{Host: key, IP: ip})
	}
	return out, nil
}
