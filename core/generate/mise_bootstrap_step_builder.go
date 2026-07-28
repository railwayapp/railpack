package generate

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/internal/utils"
)

const (
	MiseBootstrapBuildStepName   = "packages:mise:bootstrap:build"
	MiseBootstrapRuntimeStepName = "packages:mise:bootstrap:runtime"
	miseBootstrapAssetName       = "generated-mise-bootstrap-toml"
	miseBootstrapSystemConfigDir = "/tmp/railpack-mise-bootstrap/system"
	miseBootstrapConfigDir       = "/tmp/railpack-mise-bootstrap/config"
	miseBootstrapDataDir         = "/tmp/railpack-mise-bootstrap/data"
	miseBootstrapCacheDir        = "/tmp/railpack-mise-bootstrap/cache"
	miseBootstrapStateDir        = "/tmp/railpack-mise-bootstrap/state"
	miseBootstrapBinaryPath      = "/tmp/railpack-mise-bootstrap/mise"
)

type MiseBootstrapStepBuilder struct {
	DisplayName string
	Base        plan.Layer
	Inputs      []plan.Layer
	AptPackages []string
	ConfigFiles []string
	CopyMise    bool
	RunHooks    bool
	// Project packages can require the step even without Railpack-generated packages.
	ApplyProjectPackages bool
	HasProjectHooks      bool
}

// Constructs a package-only bootstrap step for an arbitrary Debian base layer.
func NewMiseBootstrapStepBuilder(
	displayName string,
	base plan.Layer,
	aptPackages []string,
	configFiles []string,
) *MiseBootstrapStepBuilder {
	return &MiseBootstrapStepBuilder{
		DisplayName: displayName,
		Base:        base,
		AptPackages: aptPackages,
		ConfigFiles: configFiles,
		RunHooks:    true,
	}
}

// Reports whether either Railpack or the application requested bootstrap work.
func (b *MiseBootstrapStepBuilder) IsRequired() bool {
	return len(b.AptPackages) > 0 ||
		b.ApplyProjectPackages ||
		(b.RunHooks && b.HasProjectHooks)
}

// Creates the plan step while keeping bootstrap state isolated from tool installs.
func (b *MiseBootstrapStepBuilder) Build(options *BuildStepOptions) (*plan.Step, error) {
	bootstrapPackages := makeMiseBootstrapPackages(b.AptPackages)
	bootstrapToml, err := mise.GenerateMiseBootstrapToml(bootstrapPackages, map[string]any{
		"paranoid":             true,
		"system_packages":      map[string]any{"managers": []string{"apt"}},
		"trusted_config_paths": []string{"/app"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate mise bootstrap config: %w", err)
	}

	step := plan.NewStep(b.DisplayName)
	step.Inputs = append([]plan.Layer{b.Base}, b.Inputs...)
	step.Caches = options.Caches.GetAptCaches()
	step.Secrets = []string{}
	if b.RunHooks && b.HasProjectHooks {
		// Repository hooks commonly use credentials, so their cache must follow secret changes.
		step.Secrets = []string{"*"}
	}
	step.Assets[miseBootstrapAssetName] = bootstrapToml

	for _, file := range b.ConfigFiles {
		step.AddCommands([]plan.Command{
			plan.NewCopyCommand(file),
		})
	}

	misePath := "/usr/local/bin/mise"
	if b.CopyMise {
		misePath = miseBootstrapBinaryPath
		step.AddCommands([]plan.Command{
			plan.CopyCommand{
				Image: RailpackBuilderImage,
				Src:   "/usr/local/bin/mise",
				Dest:  misePath,
			},
		})
	}

	step.AddCommands([]plan.Command{
		plan.NewFileCommand(
			miseBootstrapSystemConfigDir+"/config.toml",
			miseBootstrapAssetName,
			plan.FileOptions{CustomName: "create mise bootstrap config"},
		),
		plan.NewExecCommand(
			miseBootstrapCommand(misePath, b.RunHooks),
			plan.ExecOptions{CustomName: miseBootstrapDisplayName(bootstrapPackages)},
		),
	})

	step.AddCommands([]plan.Command{
		plan.NewExecCommand(
			"rm -rf /tmp/railpack-mise-bootstrap",
			plan.ExecOptions{CustomName: "remove temporary mise bootstrap files"},
		),
	})

	return step, nil
}

func makeMiseBootstrapPackages(aptPackages []string) map[string]string {
	packages := map[string]string{}
	for _, aptPackage := range utils.RemoveDuplicates(aptPackages) {
		if aptPackage == "" || aptPackage == "..." {
			continue
		}

		name, version, hasVersion := strings.Cut(aptPackage, "=")
		if !strings.HasPrefix(name, "apt:") {
			name = "apt:" + name
		}
		if !hasVersion {
			version = "latest"
		}

		packages[name] = version
	}

	return packages
}

func miseBootstrapCommand(misePath string, runHooks bool) string {
	env := []string{
		"MISE_SYSTEM_CONFIG_DIR=" + miseBootstrapSystemConfigDir,
		"MISE_CONFIG_DIR=" + miseBootstrapConfigDir,
		"MISE_DATA_DIR=" + miseBootstrapDataDir,
		"MISE_CACHE_DIR=" + miseBootstrapCacheDir,
		"MISE_STATE_DIR=" + miseBootstrapStateDir,
	}
	sort.Strings(env)

	command := " bootstrap packages apply --manager apt --yes --update"
	if runHooks {
		command = " bootstrap --only packages --yes --update"
	}

	return "env " + strings.Join(env, " ") + " " + misePath + command
}

func miseBootstrapDisplayName(packages map[string]string) string {
	names := slices.Sorted(maps.Keys(packages))
	if len(names) == 0 {
		return "install system packages with mise"
	}

	return "install system packages with mise: " + strings.Join(names, ", ")
}

func miseBootstrapRepositoryLayer() plan.Layer {
	return plan.NewStepLayer(MiseBootstrapBuildStepName, plan.Filter{
		Include: []string{
			"/etc/apt/keyrings",
			"/etc/apt/sources.list.d",
			"/etc/apt/trusted.gpg.d",
			"/usr/share/keyrings",
		},
	})
}
