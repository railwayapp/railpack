package build_llb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildEnvironmentMergeDeduplicatesPaths(t *testing.T) {
	env := BuildEnvironment{
		PathList: []string{"/first", "/shared"},
		EnvVars:  map[string]string{},
	}
	other := BuildEnvironment{
		PathList: []string{"/shared", "/last"},
		EnvVars:  map[string]string{},
	}

	env.Merge(other)

	// Keeping /shared in its original position verifies that the first occurrence retains precedence.
	require.Equal(t, []string{"/first", "/shared", "/last"}, env.PathList)
}
