package generate

import (
	"testing"

	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

func TestMakeMiseBootstrapPackages(t *testing.T) {
	packages := makeMiseBootstrapPackages([]string{
		"...",
		"curl",
		"curl",
		"libssl3=3.0.0",
		"apt:gcc:arm64",
	})

	require.Equal(t, map[string]string{
		"apt:curl":      "latest",
		"apt:gcc:arm64": "latest",
		"apt:libssl3":   "3.0.0",
	}, packages)
}

func TestMiseBootstrapStepBuilder(t *testing.T) {
	options := &BuildStepOptions{
		Caches: NewCacheContext(),
	}
	builder := NewMiseBootstrapStepBuilder(
		MiseBootstrapRuntimeStepName,
		plan.NewImageLayer("debian:bookworm-slim"),
		[]string{"curl"},
		[]string{"mise.toml"},
	)
	builder.CopyMise = true
	builder.RunHooks = false

	step, err := builder.Build(options)
	require.NoError(t, err)
	require.Equal(t, MiseBootstrapRuntimeStepName, step.Name)
	require.Equal(t, []plan.Layer{plan.NewImageLayer("debian:bookworm-slim")}, step.Inputs)
	require.Equal(t, []string{"apt", "apt-lists"}, step.Caches)
	require.Empty(t, step.Secrets)
	require.Contains(t, step.Assets[miseBootstrapAssetName], `"apt:curl" = "latest"`)
	require.Contains(t, step.Assets[miseBootstrapAssetName], `managers = ["apt"]`)

	require.IsType(t, plan.CopyCommand{}, step.Commands[0])
	require.Equal(t, plan.NewCopyCommand("mise.toml"), step.Commands[0])

	miseCopy, ok := step.Commands[1].(plan.CopyCommand)
	require.True(t, ok)
	require.Equal(t, RailpackBuilderImage, miseCopy.Image)
	require.Equal(t, "/usr/local/bin/mise", miseCopy.Src)
	require.Equal(t, miseBootstrapBinaryPath, miseCopy.Dest)

	bootstrapCommand, ok := step.Commands[3].(plan.ExecCommand)
	require.True(t, ok)
	require.Contains(t, bootstrapCommand.Cmd, miseBootstrapBinaryPath)
	require.Contains(t, bootstrapCommand.Cmd, "bootstrap packages apply --manager apt --yes --update")

	cleanupCommand, ok := step.Commands[4].(plan.ExecCommand)
	require.True(t, ok)
	require.Equal(t, "rm -rf /tmp/railpack-mise-bootstrap", cleanupCommand.Cmd)
}

func TestMiseBootstrapStepBuilderRunsProjectHooks(t *testing.T) {
	options := &BuildStepOptions{
		Caches: NewCacheContext(),
	}
	builder := NewMiseBootstrapStepBuilder(
		MiseBootstrapBuildStepName,
		plan.NewImageLayer(RailpackBuilderImage),
		nil,
		[]string{"mise.toml"},
	)
	builder.ApplyProjectPackages = true
	builder.HasProjectHooks = true

	step, err := builder.Build(options)
	require.NoError(t, err)
	require.Equal(t, []string{"*"}, step.Secrets)

	bootstrapCommand, ok := step.Commands[2].(plan.ExecCommand)
	require.True(t, ok)
	require.Contains(t, bootstrapCommand.Cmd, "bootstrap --only packages --yes --update")

	cleanupCommand, ok := step.Commands[3].(plan.ExecCommand)
	require.True(t, ok)
	require.Equal(t, "rm -rf /tmp/railpack-mise-bootstrap", cleanupCommand.Cmd)
}

func TestMiseBootstrapStepBuilderRequiresHooksOnlyWhenTheyRun(t *testing.T) {
	builder := NewMiseBootstrapStepBuilder(
		MiseBootstrapRuntimeStepName,
		plan.NewImageLayer("debian:bookworm-slim"),
		nil,
		nil,
	)
	builder.HasProjectHooks = true
	builder.RunHooks = false
	require.False(t, builder.IsRequired())

	builder.RunHooks = true
	require.True(t, builder.IsRequired())
}
