// for platforms: used by `ghcr.io/railwayapp/railpack-frontend` to accept railpack plans via BuildKit
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

	// Build arg keys
	secretsHash = "secrets-hash"
	cacheKey    = "cache-key"
	githubToken = "github-token"
)

func StartFrontend() {
	log.Info("Starting frontend")

	ctx := appcontext.Context()
	if err := gw.RunFromEnvironment(ctx, Build); err != nil {
		log.Error("error: %+v\n", err)
		os.Exit(1)
	}
}

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

	res, err := c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
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

// validatePlatform checks if the platform is supported and returns the corresponding specs.Platform
func validatePlatform(opts map[string]string) (specs.Platform, error) {
	platformStr := opts["platform"]

	// Error if multiple platforms are specified
	if strings.Contains(platformStr, ",") {
		return specs.Platform{}, fmt.Errorf("multiple platforms are not supported, got: %s", platformStr)
	}

	// Parse the platform using our helper function
	platform, err := ParsePlatformWithDefaults(platformStr)
	if err != nil {
		return specs.Platform{}, fmt.Errorf("invalid platform format: %s. Must be one of: linux/amd64, linux/arm64, etc", platformStr)
	}

	return platform, nil
}

// Read a file from the build context
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

func parseBuildArgs(opts map[string]string) map[string]string {
	buildArgs := make(map[string]string)

	for key, arg := range opts {
		if !strings.HasPrefix(key, "build-arg:") {
			continue
		}

		// Remove the "build-arg:" prefix
		name := strings.TrimPrefix(key, "build-arg:")
		buildArgs[name] = arg
	}

	return buildArgs
}
