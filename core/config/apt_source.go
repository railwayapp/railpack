package config

import (
	"fmt"
	"net/url"
	"strings"
)

// These values reach a deb822 file and a single-quoted `sh -c` command, where any of
// these characters would split a field or end the quoted string.
const aptSourceForbiddenChars = " \t\r\n'\"\\"

// AptSource is a third-party Apt repository, used to install packages Debian does not ship.
type AptSource struct {
	URI           string   `json:"uri" jsonschema:"description=The base URL of the Apt repository (e.g. https://cli.github.com/packages)"`
	Suite         string   `json:"suite" jsonschema:"description=The repository suite (e.g. stable). A value ending in / selects a flat repository and must omit components"`
	Components    []string `json:"components,omitempty" jsonschema:"description=The repository components (e.g. main). Required unless suite ends in /"`
	SigningKeyURL string   `json:"signingKeyUrl" jsonschema:"description=The HTTPS URL of the public key that signs the repository"`
}

func (s AptSource) validate() error {
	if err := validateAptSourceURL("uri", s.URI, "http", "https"); err != nil {
		return err
	}

	// A signed repository may use http, but the key is the root of trust and may not.
	if err := validateAptSourceURL("signingKeyUrl", s.SigningKeyURL, "https"); err != nil {
		return err
	}

	if s.Suite == "" || strings.ContainsAny(s.Suite, aptSourceForbiddenChars) {
		return fmt.Errorf("invalid suite %q", s.Suite)
	}

	// Apt treats a suite ending in / as a path to a flat repository, which has no components.
	isFlat := strings.HasSuffix(s.Suite, "/")
	if isFlat && len(s.Components) > 0 {
		return fmt.Errorf("suite %q is a flat repository and cannot have components", s.Suite)
	}
	if !isFlat && len(s.Components) == 0 {
		return fmt.Errorf("suite %q requires at least one component", s.Suite)
	}

	for _, component := range s.Components {
		if component == "" || strings.ContainsAny(component, aptSourceForbiddenChars) {
			return fmt.Errorf("invalid component %q", component)
		}
	}

	return nil
}

func validateAptSourceURL(field string, value string, allowedSchemes ...string) error {
	if strings.ContainsAny(value, aptSourceForbiddenChars) {
		return fmt.Errorf("invalid %s %q", field, value)
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid %s %q", field, value)
	}

	for _, scheme := range allowedSchemes {
		if parsed.Scheme == scheme {
			return nil
		}
	}

	return fmt.Errorf("%s %q must use %s", field, value, strings.Join(allowedSchemes, " or "))
}
