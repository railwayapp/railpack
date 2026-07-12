// for platforms: used by `ghcr.io/railwayapp/railpack-frontend` as a buildkit frontend
// the buildkit library consumes buildkit input and exposes it to us via the client.Client interface
// note that `frontend` and `build` are completely separate paths

package buildkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/exporter/containerimage/exptypes"
	"github.com/moby/buildkit/frontend/gateway/client"
	gw "github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	// The default local mount of where to look for the config file
	// This is "dockerfile" because that is commonly used for the config file mount
	configMountName = "dockerfile"

	// default filename for the serialized Railpack plan
	defaultRailpackPlan = "railpack-plan.json"

	// railpack build args
	secretsHash = "secrets-hash"
	cacheKey    = "cache-key"
	githubToken = "github-token"

	// buildctl --import-cache is serialized into this frontend opt by the BuildKit client
	// `docker buildx` uses a different arg name, but the buildkit frontend normalizes the opt name the frontend receives
	keyCacheImports = "cache-imports"
)

func StartFrontend() {
	log.Info("starting frontend")

	ctx := appcontext.Context()
	if err := gw.RunFromEnvironment(ctx, Build); err != nil {
		log.Error("error: %+v\n", err)
		os.Exit(1)
	}
}

// handler for the buildkit gateway to let us read the railplan plan and generate a buildkit solve
func Build(ctx context.Context, c client.Client) (*client.Result, error) {
	opts := c.BuildOpts().Opts
	buildArgs := parseBuildArgs(opts)

	cacheKey := buildArgs[cacheKey]
	secretsHash := buildArgs[secretsHash]
	githubToken := buildArgs[githubToken]

	// TODO: Support building for multiple platforms
	buildPlatform, err := validatePlatform(opts)
	if err != nil {
		return nil, err
	}

	plan, err := readRailpackPlan(ctx, c)
	if err != nil {
		return nil, err
	}

	_, err = json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling plan: %w", err)
	}

	llbState, image, err := ConvertPlanToLLB(plan, ConvertPlanOptions{
		BuildPlatform: buildPlatform,
		SecretsHash:   secretsHash,
		CacheKey:      cacheKey,
		SessionID:     c.BuildOpts().SessionID,
		GitHubToken:   githubToken,
	})
	if err != nil {
		return nil, fmt.Errorf("error converting plan to LLB: %w", err)
	}

	def, err := llbState.Marshal(ctx)
	if err != nil {
		return nil, fmt.Errorf("error marshalling LLB state: %w", err)
	}

	imageBytes, err := json.Marshal(image)
	if err != nil {
		return nil, fmt.Errorf("error marshalling image: %w", err)
	}

	// buildkit does not auto-apply --import-cache to this solve, we need to parse the frontend opt and set the CacheImports explicitly
	// cache exports are applied automatically for us since they do not impact the solve
	cacheImports, err := parseCacheImports(opts)
	if err != nil {
		return nil, err
	}
	// NOTE logs are swallowed and outputted to the buildkit container logs, not the buildctl logs
	log.Infof("frontend cache imports: %v", cacheImports)

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition:   def.ToPB(),
		CacheImports: cacheImports,
	})
	if err != nil {
		return nil, err
	}

	res.AddMeta(exptypes.ExporterImageConfigKey, imageBytes)

	return res, nil
}

func readRailpackPlan(ctx context.Context, c client.Client) (*plan.BuildPlan, error) {
	opts := c.BuildOpts().Opts
	filename := opts["filename"]
	if filename == "" {
		filename = defaultRailpackPlan
	}

	fileContents, err := readFile(ctx, c, filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read railpack plan")
	}

	plan := plan.NewBuildPlan()
	err = json.Unmarshal([]byte(fileContents), plan)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse railpack plan")
	}

	return plan, nil
}

// checks if the platform is supported and returns the corresponding platform specs
func validatePlatform(opts map[string]string) (specs.Platform, error) {
	platformStr := opts["platform"]

	// Error if multiple platforms are specified
	if strings.Contains(platformStr, ",") {
		return specs.Platform{}, fmt.Errorf("multiple platforms are not supported, got: %s", platformStr)
	}

	platform, err := ParsePlatformWithDefaults(platformStr)
	if err != nil {
		return specs.Platform{}, fmt.Errorf("invalid platform format: %s. Must be one of: linux/amd64, linux/arm64, etc", platformStr)
	}

	return platform, nil
}

// Read a file from the build context. The frontend does not have a full `App` struct, which is why we have this helper
// to read railpack-plan.json.
func readFile(ctx context.Context, c client.Client, filename string) (string, error) {
	// Create a Local source for the dockerfile
	src := llb.Local(configMountName,
		llb.FollowPaths([]string{filename}),
		llb.SessionID(c.BuildOpts().SessionID),
		llb.WithCustomName("load build definition from "+filename),
	)

	srcDef, err := src.Marshal(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal local source")
	}

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: srcDef.ToPB(),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve dockerfile")
	}

	ref, err := res.SingleRef()
	if err != nil {
		return "", err
	}

	content, err := ref.ReadFile(ctx, client.ReadRequest{
		Filename: filename,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to read file")
	}

	fileContents := string(content)

	return fileContents, nil
}

// Extracts Docker/buildx --build-arg values from frontend opts.
//
// Docker/buildx always namespaces those as "build-arg:<name>" (e.g.
// --build-arg cache-key=x → opts["build-arg:cache-key"]). bare buildctl
// --opt cache-key=x is opts["cache-key"] and is not returned here; use
// --opt build-arg:cache-key=x for the same shape as Docker.
func parseBuildArgs(opts map[string]string) map[string]string {
	buildArgs := make(map[string]string)

	for key, arg := range opts {
		if !strings.HasPrefix(key, "build-arg:") {
			continue
		}

		name := strings.TrimPrefix(key, "build-arg:")
		buildArgs[name] = arg
	}

	return buildArgs
}

// reads the "cache-imports" frontend opt set by the BuildKit
func parseCacheImports(opts map[string]string) ([]client.CacheOptionsEntry, error) {
	cacheImportsStr := opts[keyCacheImports]
	if cacheImportsStr == "" {
		return nil, nil
	}

	// Same JSON shape as control API CacheOptionsEntry (Type / Attrs).
	var entries []struct {
		Type  string            `json:"Type"`
		Attrs map[string]string `json:"Attrs"`
	}
	if err := json.Unmarshal([]byte(cacheImportsStr), &entries); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal %s (%q)", keyCacheImports, cacheImportsStr)
	}

	cacheImports := make([]client.CacheOptionsEntry, 0, len(entries))
	for _, e := range entries {
		cacheImports = append(cacheImports, client.CacheOptionsEntry{Type: e.Type, Attrs: e.Attrs})
	}
	return cacheImports, nil
}
