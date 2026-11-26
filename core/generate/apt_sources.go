package generate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	aptSourcesDir  = "/etc/apt/sources.list.d"
	aptKeyringsDir = "/etc/apt/keyrings"
	// Ships with apt, so keys can be fetched on images without curl or gnupg (such as the runtime image).
	aptHelperPath = "/usr/lib/apt/apt-helper"
)

// Adds configured repositories to a step before installing packages, since `apt-get
// update` must see the new sources. Sources without packages are still written so that
// a user-defined step can install from them.
func (o *BuildStepOptions) addAptCommands(step *plan.Step, sources []config.AptSource, pkgs []string) {
	addedSources := []string{}

	for _, source := range sources {
		id := aptSourceID(source)
		// The same source may be configured twice when build and deploy share an image.
		if slices.Contains(addedSources, id) {
			continue
		}
		addedSources = append(addedSources, id)

		keyPath := path.Join(aptKeyringsDir, id+aptKeyExtension(source.SigningKeyURL))
		assetName := id + ".sources"
		step.Assets[assetName] = renderAptSource(source, keyPath)

		step.AddCommands([]plan.Command{
			plan.NewExecShellCommand(
				fmt.Sprintf("mkdir -p %s && %s download-file %s %s", aptKeyringsDir, aptHelperPath, source.SigningKeyURL, keyPath),
				plan.ExecOptions{CustomName: "download apt signing key: " + source.SigningKeyURL},
			),
			plan.NewFileCommand(path.Join(aptSourcesDir, assetName), assetName, plan.FileOptions{
				CustomName: "add apt source: " + source.URI,
			}),
		})
	}

	if len(pkgs) > 0 {
		step.AddCommands([]plan.Command{o.NewAptInstallCommand(pkgs)})
	}
}

// Derives a stable name from the source's contents so layers stay cacheable across
// builds and reordering the config does not change the generated files.
func aptSourceID(source config.AptSource) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		source.URI,
		source.Suite,
		strings.Join(source.Components, " "),
		source.SigningKeyURL,
	}, "\n")))

	return "railpack-" + hex.EncodeToString(hash[:6])
}

// gpgv (Debian before trixie) relies on the extension to tell binary keys from armored
// ones. Binary keys are conventionally published as .gpg and anything else is armored.
func aptKeyExtension(keyURL string) string {
	parsed, err := url.Parse(keyURL)
	if err == nil && strings.HasSuffix(parsed.Path, ".gpg") {
		return ".gpg"
	}

	return ".asc"
}

// Signed-By scopes the key to this repository, unlike a key in the global trusted.gpg.d.
func renderAptSource(source config.AptSource, keyPath string) string {
	lines := []string{
		"Types: deb",
		"URIs: " + source.URI,
		"Suites: " + source.Suite,
	}

	if len(source.Components) > 0 {
		lines = append(lines, "Components: "+strings.Join(source.Components, " "))
	}

	lines = append(lines, "Signed-By: "+keyPath)

	return strings.Join(lines, "\n") + "\n"
}
