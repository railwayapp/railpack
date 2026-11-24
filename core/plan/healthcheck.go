package plan

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/invopop/jsonschema"
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
)

const (
	// Dockerfile HEALTHCHECK defaults.
	// https://docs.docker.com/reference/dockerfile/#healthcheck
	DefaultHealthcheckInterval      = "30s"
	DefaultHealthcheckTimeout       = "30s"
	DefaultHealthcheckStartPeriod   = "0s"
	DefaultHealthcheckStartInterval = "5s"
	DefaultHealthcheckRetries       = 3
)

// TCP port numbers occupy 16 bits; zero is reserved rather than a service port.
const maxHealthcheckPort = 65535

// The URL is supplied as $1, so its contents are never parsed as shell code.
const urlHealthcheckCommand = `probe_url=$1
if [[ "$probe_url" == *'${PORT}'* ]]; then
    probe_url=${probe_url//\$\{PORT\}/"${PORT:?PORT must be set for this healthcheck}"}
fi
status=$(curl --disable --silent --show-error --fail --globoff --noproxy '*' \
    --output /dev/null --write-out '%{http_code}' -- "$probe_url") || exit 1
# A 2xx response proves success; redirects are not followed or treated as healthy.
[[ "$status" == 2[0-9][0-9] ]]`

// Container healthcheck from railpack.json, using either a shell command or an HTTP(S) URL.
type Healthcheck struct {
	Command       string `json:"command,omitempty" jsonschema:"minLength=1,description=Shell command the runtime runs. Exit 0 is healthy. Set exactly one of command or url."`
	URL           string `json:"url,omitempty" jsonschema:"minLength=1,pattern=^https?://,description=HTTP(S) URL to GET. Only 2xx responses are healthy. Supports ${PORT} at runtime. Set exactly one of command or url."`
	Interval      string `json:"interval,omitempty" jsonschema:"description=Time between checks as a Go duration such as 30s. Defaults to 30s."`
	Timeout       string `json:"timeout,omitempty" jsonschema:"description=Maximum time one check may run as a Go duration. Defaults to 30s."`
	StartPeriod   string `json:"startPeriod,omitempty" jsonschema:"description=Startup grace period as a Go duration. Failures during this period do not count toward retries. Defaults to 0s."`
	StartInterval string `json:"startInterval,omitempty" jsonschema:"description=Time between checks during the start period as a Go duration. Defaults to 5s."`
	Retries       int    `json:"retries,omitempty" jsonschema:"description=Consecutive failures before the container is unhealthy. Defaults to 3."`
}

// Require one probe type while sharing timing options between both forms.
func (Healthcheck) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.OneOf = []*jsonschema.Schema{
		{Required: []string{"command"}},
		{Required: []string{"url"}},
	}
}

// Converts to Docker image metadata, applying defaults and reporting every invalid field.
func (h *Healthcheck) ToImageConfig() (*dockerspec.HealthcheckConfig, error) {
	if h == nil {
		return nil, nil
	}
	config := &dockerspec.HealthcheckConfig{
		// The Debian runtime uses dash for /bin/sh; healthcheck commands support bash syntax.
		Test:    []string{"CMD", "/bin/bash", "-c", h.Command},
		Retries: cmp.Or(h.Retries, DefaultHealthcheckRetries),
	}
	var errs []error
	if (h.Command == "") == (h.URL == "") {
		errs = append(errs, errors.New("deploy.healthcheck must set exactly one of command or url"))
	}
	if h.URL != "" {
		if err := validateHealthcheckURL(h.URL); err != nil {
			errs = append(errs, err)
		} else {
			// bash -c consumes the first argument after the script as $0, leaving the URL in $1.
			config.Test = []string{"CMD", "/bin/bash", "-c", urlHealthcheckCommand, "railpack-healthcheck", h.URL}
		}
	}
	if h.Retries < 0 {
		errs = append(errs, errors.New("deploy.healthcheck.retries must be greater than zero"))
	}

	// Duration constraints live with the fields so one pass can report every failure.
	fields := []struct {
		name string
		raw  string
		dest *time.Duration
	}{
		{"interval", cmp.Or(h.Interval, DefaultHealthcheckInterval), &config.Interval},
		{"timeout", cmp.Or(h.Timeout, DefaultHealthcheckTimeout), &config.Timeout},
		{"startPeriod", cmp.Or(h.StartPeriod, DefaultHealthcheckStartPeriod), &config.StartPeriod},
		{"startInterval", cmp.Or(h.StartInterval, DefaultHealthcheckStartInterval), &config.StartInterval},
	}
	for _, field := range fields {
		duration, err := parseHealthcheckDuration(field.raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("deploy.healthcheck.%s is invalid: %w", field.name, err))
			continue
		}
		*field.dest = duration
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return config, nil
}

// Empty and zero match an unset Dockerfile duration. Docker rejects any other value below 1ms.
func parseHealthcheckDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", value, err)
	}
	if duration < 0 || (duration > 0 && duration < time.Millisecond) {
		return 0, fmt.Errorf("duration %q must be 0 or at least 1ms", value)
	}

	return duration, nil
}

// Validate literal URLs at plan time; curl validates templates after runtime expansion.
func validateHealthcheckURL(address string) error {
	if (!strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://")) || strings.ContainsAny(address, " \t\r\n") {
		return errors.New("deploy.healthcheck.url must be an absolute http:// or https:// URL")
	}
	if strings.Contains(address, "${PORT}") {
		// The final port is unknown until the container starts; do not substitute a build-time value.
		return nil
	}
	parsed, err := url.Parse(address)
	if err != nil || parsed.Hostname() == "" {
		return errors.New("deploy.healthcheck.url must be an absolute http:// or https:// URL")
	}
	if port := parsed.Port(); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil || value < 1 || value > maxHealthcheckPort {
			return fmt.Errorf("deploy.healthcheck.url port must be between 1 and %d", maxHealthcheckPort)
		}
	}
	return nil
}
