package buildkit

import (
	"context"
	"net"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestConvertPlanToLLB_NetworkModeAndExtraHosts(t *testing.T) {
	t.Parallel()

	// Deploy inputs need include filters so the step state (and its exec) is
	// wired into the final image LLB rather than being dropped by copy/merge.
	buildPlan := &plan.BuildPlan{
		Steps: []plan.Step{
			{
				Name:   "build",
				Inputs: []plan.Layer{plan.NewImageLayer("alpine:latest")},
				Commands: []plan.Command{
					plan.NewExecCommand("curl my-other-container"),
				},
			},
		},
		Deploy: plan.Deploy{
			Base: plan.NewImageLayer("alpine:latest"),
			Inputs: []plan.Layer{
				plan.NewStepLayer("build", plan.NewIncludeFilter([]string{"."})),
			},
		},
	}

	t.Run("defaults are sandbox with no extra hosts", func(t *testing.T) {
		state, _, err := ConvertPlanToLLB(buildPlan, ConvertPlanOptions{
			BuildPlatform: specs.Platform{OS: "linux", Architecture: "amd64"},
		})
		require.NoError(t, err)

		execs := marshalExecOps(t, state)
		require.NotEmpty(t, execs)
		for _, exec := range execs {
			require.Equal(t, pb.NetMode_UNSET, exec.Network)
			require.Empty(t, exec.Meta.ExtraHosts)
		}
	})

	t.Run("host network and extra hosts applied to execs", func(t *testing.T) {
		state, _, err := ConvertPlanToLLB(buildPlan, ConvertPlanOptions{
			BuildPlatform: specs.Platform{OS: "linux", Architecture: "amd64"},
			NetworkMode:   llb.NetModeHost,
			ExtraHosts: []llb.HostIP{
				{Host: "my-other-container", IP: net.ParseIP("192.168.107.2")},
			},
		})
		require.NoError(t, err)

		execs := marshalExecOps(t, state)
		require.NotEmpty(t, execs)

		var found bool
		for _, exec := range execs {
			if exec.Network != pb.NetMode_HOST {
				continue
			}
			found = true
			require.Len(t, exec.Meta.ExtraHosts, 1)
			require.Equal(t, "my-other-container", exec.Meta.ExtraHosts[0].Host)
			require.Equal(t, "192.168.107.2", exec.Meta.ExtraHosts[0].IP)
		}
		require.True(t, found, "expected at least one exec op with host network mode")
	})
}

func marshalExecOps(t *testing.T, state *llb.State) []*pb.ExecOp {
	t.Helper()

	def, err := state.Marshal(context.Background())
	require.NoError(t, err)

	var execs []*pb.ExecOp
	for _, dt := range def.Def {
		var op pb.Op
		require.NoError(t, op.Unmarshal(dt))
		if exec := op.GetExec(); exec != nil {
			execs = append(execs, exec)
		}
	}
	return execs
}
