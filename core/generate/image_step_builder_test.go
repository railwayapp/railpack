package generate

import (
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestImageStepBuilderReusesBuildBootstrapRepositories(t *testing.T) {
	builder := &ImageStepBuilder{
		DisplayName: "packages:image",
		ResolveStepImage: func(_ *BuildStepOptions) string {
			return "debian:bookworm-slim"
		},
		AptPackages: []string{"curl"},
	}
	options := &BuildStepOptions{
		Caches: NewCacheContext(),
		MiseBootstrapProject: MiseBootstrapProjectConfig{
			ConfigFiles:     []string{"mise.toml"},
			HasAptPackages:  true,
			HasPackageHooks: true,
		},
	}
	buildPlan := plan.NewBuildPlan()

	err := builder.Build(buildPlan, options)
	require.NoError(t, err)
	require.Len(t, buildPlan.Steps, 1)

	step := buildPlan.Steps[0]
	require.Equal(t, []plan.Layer{
		plan.NewImageLayer("debian:bookworm-slim"),
		miseBootstrapRepositoryLayer(),
	}, step.Inputs)
	require.Empty(t, step.Secrets)

	bootstrapCommand, ok := step.Commands[3].(plan.ExecCommand)
	require.True(t, ok)
	require.Contains(
		t,
		bootstrapCommand.Cmd,
		"bootstrap packages apply --manager apt --yes --update",
	)
}
