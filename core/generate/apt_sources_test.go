package generate

import (
	"testing"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/stretchr/testify/require"
)

var testAptSource = config.AptSource{
	URI:           "https://example.com/debian",
	Suite:         "stable",
	Components:    []string{"main", "contrib"},
	SigningKeyURL: "https://example.com/key.asc",
}

func commandTypes(commands []plan.Command) []string {
	types := []string{}
	for _, command := range commands {
		types = append(types, command.CommandType())
	}
	return types
}

func TestAptCommandsRegisterSourcesBeforeInstallingPackages(t *testing.T) {
	step := plan.NewStep("apt")

	(&BuildStepOptions{}).addAptCommands(step, []config.AptSource{testAptSource}, []string{"example"})

	// `apt-get update` in the final exec must run with the key and source file already in place.
	require.Equal(t, []string{"exec", "file", "exec"}, commandTypes(step.Commands))

	keyCommand := step.Commands[0].(plan.ExecCommand)
	require.Contains(t, keyCommand.Cmd, "/usr/lib/apt/apt-helper download-file https://example.com/key.asc /etc/apt/keyrings/railpack-")

	fileCommand := step.Commands[1].(plan.FileCommand)
	require.Regexp(t, `^/etc/apt/sources\.list\.d/railpack-[0-9a-f]{12}\.sources$`, fileCommand.Path)
	require.Contains(t, step.Assets[fileCommand.Name], "URIs: https://example.com/debian\nSuites: stable\nComponents: main contrib\nSigned-By: /etc/apt/keyrings/railpack-")

	require.Contains(t, step.Commands[2].(plan.ExecCommand).Cmd, "apt-get install -y example")
}

func TestAptCommandsWriteSourcesWithoutInstallingWhenNoPackagesAreConfigured(t *testing.T) {
	step := plan.NewStep("apt")

	(&BuildStepOptions{}).addAptCommands(step, []config.AptSource{testAptSource}, nil)

	require.Equal(t, []string{"exec", "file"}, commandTypes(step.Commands))
}

func TestAptCommandsAddADuplicateSourceOnce(t *testing.T) {
	step := plan.NewStep("apt")

	(&BuildStepOptions{}).addAptCommands(step, []config.AptSource{testAptSource, testAptSource}, nil)

	require.Equal(t, []string{"exec", "file"}, commandTypes(step.Commands))
	require.Len(t, step.Assets, 1)
}

func TestRenderedFlatAptSourceOmitsComponents(t *testing.T) {
	flat := config.AptSource{URI: "https://example.com/repo", Suite: "./", SigningKeyURL: "https://example.com/key.gpg"}

	require.Equal(t,
		"Types: deb\nURIs: https://example.com/repo\nSuites: ./\nSigned-By: /etc/apt/keyrings/key.gpg\n",
		renderAptSource(flat, "/etc/apt/keyrings/key.gpg"),
	)
}

func TestAptKeyExtensionReflectsThePublishedKeyFormat(t *testing.T) {
	require.Equal(t, ".gpg", aptKeyExtension("https://example.com/keyring.gpg"))
	require.Equal(t, ".gpg", aptKeyExtension("https://example.com/keyring.gpg?v=2"))
	require.Equal(t, ".asc", aptKeyExtension("https://example.com/key.asc"))
}

func TestConfiguredAptSourcesReachTheirApplicableBuilders(t *testing.T) {
	ctx := CreateTestContext(t, "../../examples/node-npm")
	deploySource := testAptSource
	deploySource.Suite = "testing"

	cfg := config.EmptyConfig()
	cfg.BuildAptSources = []config.AptSource{testAptSource}
	cfg.Deploy.AptSources = []config.AptSource{deploySource}
	ctx.Config = cfg

	ctx.applyConfig()

	require.Equal(t, []config.AptSource{testAptSource}, ctx.GetMiseStepBuilder().AptSources)
	require.Equal(t, []config.AptSource{deploySource}, ctx.Deploy.AptSources)
}
