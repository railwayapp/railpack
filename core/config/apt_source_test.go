package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func validAptSource() AptSource {
	return AptSource{
		URI:           "https://example.com/debian",
		Suite:         "stable",
		Components:    []string{"main"},
		SigningKeyURL: "https://example.com/key.asc",
	}
}

func TestAptSourceValidationAcceptsSupportedRepositories(t *testing.T) {
	require.NoError(t, validAptSource().validate())

	// Signatures protect packages fetched over http, so repositories may use it.
	httpRepo := validAptSource()
	httpRepo.URI = "http://example.com/debian"
	require.NoError(t, httpRepo.validate())

	flatRepo := validAptSource()
	flatRepo.Suite = "./"
	flatRepo.Components = nil
	require.NoError(t, flatRepo.validate())
}

func TestAptSourceValidationRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(s *AptSource)
	}{
		{"missing uri", func(s *AptSource) { s.URI = "" }},
		{"uri without host", func(s *AptSource) { s.URI = "https://" }},
		{"unsupported uri scheme", func(s *AptSource) { s.URI = "ftp://example.com/debian" }},
		{"uri with a field break", func(s *AptSource) { s.URI = "https://example.com/\nSigned-By: /tmp/key" }},
		{"missing key", func(s *AptSource) { s.SigningKeyURL = "" }},
		{"key over http", func(s *AptSource) { s.SigningKeyURL = "http://example.com/key.asc" }},
		{"key that ends the shell quote", func(s *AptSource) { s.SigningKeyURL = "https://example.com/k'; id; '" }},
		{"missing suite", func(s *AptSource) { s.Suite = "" }},
		{"suite with a space", func(s *AptSource) { s.Suite = "stable testing" }},
		{"missing components", func(s *AptSource) { s.Components = nil }},
		{"empty component", func(s *AptSource) { s.Components = []string{""} }},
		{"flat repository with components", func(s *AptSource) { s.Suite = "./" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := validAptSource()
			tt.mutate(&source)
			require.Error(t, source.validate())
		})
	}
}

func TestConfigValidateNamesTheInvalidAptSource(t *testing.T) {
	invalid := validAptSource()
	invalid.Suite = ""

	build := EmptyConfig()
	build.BuildAptSources = []AptSource{validAptSource(), invalid}
	require.ErrorContains(t, build.Validate(), "buildAptSources[1]")

	deploy := EmptyConfig()
	deploy.Deploy.AptSources = []AptSource{invalid}
	require.ErrorContains(t, deploy.Validate(), "deploy.aptSources[0]")

	require.NoError(t, EmptyConfig().Validate())
}
