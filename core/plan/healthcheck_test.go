package plan

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthcheckImageConfigReportsEveryInvalidField(t *testing.T) {
	_, err := (&Healthcheck{
		Interval:    "1us",
		Timeout:     "nope",
		StartPeriod: "-1s",
		Retries:     -1,
	}).ToImageConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, "deploy.healthcheck must set exactly one of command or url")
	require.ErrorContains(t, err, "deploy.healthcheck.interval is invalid")
	require.ErrorContains(t, err, "deploy.healthcheck.timeout is invalid")
	require.ErrorContains(t, err, "deploy.healthcheck.startPeriod is invalid")
	require.ErrorContains(t, err, "deploy.healthcheck.retries must be greater than zero")
}

func TestHealthcheckImageConfigUsesDockerfileDefaults(t *testing.T) {
	healthcheck := &Healthcheck{Command: "true"}
	config, err := healthcheck.ToImageConfig()
	require.NoError(t, err)
	require.Equal(t, 30*time.Second, config.Interval)
	require.Equal(t, 30*time.Second, config.Timeout)
	require.Equal(t, time.Duration(0), config.StartPeriod)
	require.Equal(t, 5*time.Second, config.StartInterval)
	require.Equal(t, DefaultHealthcheckRetries, config.Retries)
	require.Empty(t, healthcheck.Interval)
	require.Zero(t, healthcheck.Retries)
}

func TestHealthcheckImageConfigOmitted(t *testing.T) {
	var healthcheck *Healthcheck
	config, err := healthcheck.ToImageConfig()
	require.NoError(t, err)
	require.Nil(t, config)
}

func TestHealthcheckImageConfigPreservesOverrides(t *testing.T) {
	config, err := (&Healthcheck{
		Command:  "echo healthy",
		Interval: "12s", Timeout: "4s", StartPeriod: "1m", StartInterval: "2s", Retries: 7,
	}).ToImageConfig()
	require.NoError(t, err)
	require.Equal(t, []string{"CMD", "/bin/bash", "-c", "echo healthy"}, config.Test)
	require.Equal(t, 12*time.Second, config.Interval)
	require.Equal(t, 4*time.Second, config.Timeout)
	require.Equal(t, time.Minute, config.StartPeriod)
	require.Equal(t, 2*time.Second, config.StartInterval)
	require.Equal(t, 7, config.Retries)
}

func TestHealthcheckRejectsInvalidProbeConfiguration(t *testing.T) {
	for _, tt := range []struct {
		name        string
		healthcheck Healthcheck
	}{
		{"missing probe", Healthcheck{}},
		{"both probes", Healthcheck{Command: "true", URL: "http://localhost/status"}},
		{"relative URL", Healthcheck{URL: "/status"}},
		{"unsupported scheme", Healthcheck{URL: "file:///status"}},
		{"missing host", Healthcheck{URL: "http:///status"}},
		{"invalid port", Healthcheck{URL: "http://localhost:nope/status"}},
		{"out of range port", Healthcheck{URL: "http://localhost:65536/status"}},
		{"invalid escape", Healthcheck{URL: "http://localhost/%zz"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.healthcheck.ToImageConfig()
			require.Error(t, err)
		})
	}
}

func TestHealthcheckURLResponseStatus(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent, http.StatusFound, http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.Header().Set("Location", "/healthy")
				w.WriteHeader(status)
			}))
			defer server.Close()
			_, err := runURLHealthcheck(t, server.URL)
			if status >= http.StatusOK && status < http.StatusMultipleChoices {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestHealthcheckURLConnectionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()
	_, err := runURLHealthcheck(t, server.URL)
	require.Error(t, err)
}

func TestHealthcheckURLVerifiesTLSCertificates(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	_, err := runURLHealthcheck(t, server.URL)
	require.Error(t, err)

	certificate := filepath.Join(t.TempDir(), "ca.pem")
	require.NoError(t, os.WriteFile(certificate, pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE", Bytes: server.Certificate().Raw,
	}), 0600))
	t.Setenv("CURL_CA_BUNDLE", certificate)
	_, err = runURLHealthcheck(t, server.URL)
	require.NoError(t, err)
}

func TestHealthcheckURLQuotesDataAndExpandsPort(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "unexpected-command")
	// Shell substitutions and curl glob syntax must reach the server as URL data.
	query := "single='&double=\"&dollar=$(touch${IFS}" + marker + ")&backtick=`id`&glob=[a-z]"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, query, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	parsed, err := url.Parse(server.URL)
	require.NoError(t, err)
	t.Setenv("PORT", parsed.Port())
	// A proxy environment must not send a local health probe outside the container.
	t.Setenv("http_proxy", "http://127.0.0.1:1")
	t.Setenv("ALL_PROXY", "http://127.0.0.1:1")
	output, err := runURLHealthcheck(t, "http://127.0.0.1:${PORT}/?"+query)
	require.NoError(t, err, output)
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err))
}

func runURLHealthcheck(t *testing.T, address string) (string, error) {
	t.Helper()
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl is required to exercise HTTP probes")
	}
	config, err := (&Healthcheck{URL: address}).ToImageConfig()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, config.Test[1], config.Test[2:]...).CombinedOutput()
	return string(output), err
}

func TestHealthcheckURLRejectsInvalidRuntimePort(t *testing.T) {
	for _, tt := range []struct {
		name string
		port string
	}{
		{"unset", ""},
		{"nonnumeric", "not-a-port"},
		{"out of range", strconv.Itoa(maxHealthcheckPort + 1)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)
			output, err := runURLHealthcheck(t, "http://localhost:${PORT}/status")
			require.Error(t, err, output)
		})
	}
}

func TestHealthcheckURLTreatsRuntimePortAsData(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "unexpected-command")
	// An environment value must not gain shell syntax when inserted into the URL.
	t.Setenv("PORT", "$(touch "+marker+")")
	_, err := runURLHealthcheck(t, "http://localhost:${PORT}/status")
	require.Error(t, err)
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err))
}
