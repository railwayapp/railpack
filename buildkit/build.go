// called by the build CLI entrypoint and runs the build using the buildkit client
// also used by the integration tests to run builds in a test environment

package buildkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/containerd/platforms"
	"github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/client"
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer"
	_ "github.com/moby/buildkit/client/connhelper/nerdctlcontainer"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/util/appcontext"
	_ "github.com/moby/buildkit/util/grpcutil/encoding/proto"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/tonistiigi/fsutil"
)

const (
	buildkitHostNotSetError = `BUILDKIT_HOST environment variable is not set.

To start a local BuildKit daemon and set the environment variable run:

	docker run --rm --privileged -d --name buildkit moby/buildkit
	export BUILDKIT_HOST='docker-container://buildkit'`

	buildkitInfoError = `failed to get buildkit information.

Most likely the $BUILDKIT_HOST is not running. Here's an example of how to start the build container:

	docker run --rm --privileged -d --name buildkit moby/buildkit

Use 'railpack --verbose' to view more error details`
)

type BuildWithBuildkitClientOptions struct {
	ImageName    string
	DumpLLB      bool
	OutputDir    string
	ProgressMode string
	SecretsHash  string
	Secrets      map[string]string
	Platform     string
	ImportCache  []string
	ExportCache  []string
	CacheKey     string
	GitHubToken  string
}

func BuildWithBuildkitClient(appDir string, plan *plan.BuildPlan, opts BuildWithBuildkitClientOptions) error {
	ctx := appcontext.Context()

	imageName := opts.ImageName
	if imageName == "" {
		imageName = getImageName(appDir)
	}

	buildkitHost := os.Getenv("BUILDKIT_HOST")
	if buildkitHost == "" {
		return errors.New(buildkitHostNotSetError)
	}

	log.Debugf("Connecting to buildkit host: %s", buildkitHost)

	// connecting to the buildkit host does *not* mean the specified build container is running
	c, err := client.New(ctx, buildkitHost)
	if err != nil {
		return fmt.Errorf("failed to connect to buildkit: %w", err)
	}
	defer func() { _ = c.Close() }()

	// Get the buildkit info early so we can ensure we can connect to the buildkit host
	info, err := c.Info(ctx)
	if err != nil {
		log.Debugf("error getting buildkit info: %v", err)
		return errors.New(buildkitInfoError)
	}

	// Parse the platform string using our helper function
	buildPlatform, err := ParsePlatformWithDefaults(opts.Platform)
	if err != nil {
		return fmt.Errorf("failed to parse platform '%s': %w", opts.Platform, err)
	}

	llbState, image, err := ConvertPlanToLLB(plan, ConvertPlanOptions{
		BuildPlatform: buildPlatform,
		SecretsHash:   opts.SecretsHash,
		CacheKey:      opts.CacheKey,
		GitHubToken:   opts.GitHubToken,
	})
	if err != nil {
		return fmt.Errorf("error converting plan to LLB: %w", err)
	}

	imageBytes, err := json.Marshal(image)
	if err != nil {
		return fmt.Errorf("error marshalling image: %w", err)
	}

	def, err := llbState.Marshal(ctx, llb.LinuxAmd64)
	if err != nil {
		return fmt.Errorf("error marshaling LLB state: %w", err)
	}

	if opts.DumpLLB {
		err = llb.WriteTo(def, os.Stdout)
		if err != nil {
			return fmt.Errorf("error writing LLB definition: %w", err)
		}
		return nil
	}

	ch := make(chan *client.SolveStatus)

	var pipeR *io.PipeReader
	var pipeW *io.PipeWriter
	errCh := make(chan error, 1)

	// Only set up pipe and docker load if we're not saving to a directory
	if opts.OutputDir == "" {
		// Create a pipe to connect buildkit output to docker load
		pipeR, pipeW = io.Pipe()
		defer func() { _ = pipeR.Close() }()

		// Pipe the image into `docker load`
		go func() {
			cmd := exec.Command("docker", "load")
			cmd.Stdin = pipeR
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			errCh <- cmd.Run()
		}()
	}

	progressDone := make(chan bool)
	go func() {
		displayCh := make(chan *client.SolveStatus)
		go func() {
			for s := range ch {
				displayCh <- s
			}
			close(displayCh)
		}()

		progressMode := progressui.AutoMode
		switch opts.ProgressMode {
		case "plain":
			progressMode = progressui.PlainMode
		case "tty":
			progressMode = progressui.TtyMode
		}

		display, err := progressui.NewDisplay(os.Stdout, progressMode)
		if err != nil {
			log.Error("failed to create progress display", "error", err)
		}

		_, err = display.UpdateFrom(ctx, displayCh)
		if err != nil {
			log.Error("failed to update progress display", "error", err)
		}
		progressDone <- true
	}()

	appFS, err := fsutil.NewFS(appDir)
	if err != nil {
		return fmt.Errorf("error creating FS: %w", err)
	}

	log.Debugf("Building image for %s with BuildKit %s", platforms.Format(buildPlatform), info.BuildkitVersion.Version)

	secretsMap := make(map[string][]byte)
	for k, v := range opts.Secrets {
		secretsMap[k] = []byte(v)
	}
	secrets := secretsprovider.FromMap(secretsMap)

	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	sessionAttachables := []session.Attachable{
		secrets,
		// buildkit does not use the local auth arguments by default, which prevents private repo access when running `railpack build`
		authprovider.NewDockerAuthProvider(authprovider.DockerAuthProviderConfig{
			AuthConfigProvider: authprovider.LoadAuthConfig(dockerConfig),
		}),
	}

	solveOpts := client.SolveOpt{
		LocalMounts: map[string]fsutil.FS{
			"context": appFS,
		},
		Session: sessionAttachables,
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterDocker,
				Attrs: map[string]string{
					"name":                  imageName,
					"containerimage.config": string(imageBytes),
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return pipeW, nil
				},
			},
		},
	}

	solveOpts.CacheImports = cacheEntriesFromFlags(opts.ImportCache)
	solveOpts.CacheExports = cacheEntriesFromFlags(opts.ExportCache)

	log.Infof("cache imports: %v", solveOpts.CacheImports)
	log.Infof("cache exports: %v", solveOpts.CacheExports)

	// Save the resulting filesystem to a directory
	if opts.OutputDir != "" {
		err = os.MkdirAll(opts.OutputDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating output directory: %w", err)
		}

		solveOpts.Exports = []client.ExportEntry{
			{
				Type:      client.ExporterLocal,
				OutputDir: opts.OutputDir,
			},
		}
	}

	startTime := time.Now()
	_, err = c.Solve(ctx, def, solveOpts, ch)

	// Wait for progress monitoring to complete
	<-progressDone

	if pipeW != nil {
		_ = pipeW.Close()
	}

	if err != nil {
		return fmt.Errorf("failed to solve: %w", err)
	}

	// Only wait for docker load if we used it
	if opts.OutputDir == "" {
		if err := <-errCh; err != nil {
			return fmt.Errorf("docker load failed: %w", err)
		}
	}

	buildDuration := time.Since(startTime)
	log.Infof("Successfully built image in %.2fs", buildDuration.Seconds())

	if opts.OutputDir != "" {
		log.Infof("Saved image filesystem to directory `%s`", opts.OutputDir)
	} else {
		log.Infof("Run with `docker run -it %s`", imageName)
	}

	return nil
}

// determine the image name from the app dir path
func getImageName(appDir string) string {
	parts := strings.Split(appDir, string(os.PathSeparator))
	name := parts[len(parts)-1]

	// TODO how could this happen in practice?
	if name == "" {
		name = "railpack-app" // Fallback if path ends in separator
	}

	// Docker requires image names to be lowercase
	return strings.ToLower(name)
}

// Converts docker buildx-style cache flag values (e.g. type=registry,ref=...)
// into BuildKit CacheOptionsEntry values. Empty strings are skipped.
//
// Intentionally hand-rolled instead of using github.com/docker/buildx/util/buildflags
// (or pulling buildx solely for ParseCacheEntry). BuildKit has no public parser for
// these strings; buildx does, but the dependency cost outweighs the small amount of
// logic we need for the type=... form we document.
func cacheEntriesFromFlags(entries []string) []client.CacheOptionsEntry {
	var out []client.CacheOptionsEntry
	for _, entry := range entries {
		if entry == "" {
			continue
		}
		cacheType, attrs := extractCacheType(parseKeyValue(entry))
		out = append(out, client.CacheOptionsEntry{
			Type:  cacheType,
			Attrs: attrs,
		})
	}
	return out
}

// parse comma-separated key=value strings into a map, ignoring entries without an "="
func parseKeyValue(s string) map[string]string {
	attrs := make(map[string]string)
	parts := strings.SplitSeq(s, ",")
	for part := range parts {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}
		attrs[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return attrs
}

func extractCacheType(attrs map[string]string) (string, map[string]string) {
	cacheType := attrs["type"]

	cleanedAttrs := maps.Clone(attrs)
	delete(cleanedAttrs, "type")

	return cacheType, cleanedAttrs
}
