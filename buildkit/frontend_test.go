package buildkit

import (
	"testing"

	"github.com/moby/buildkit/solver/pb"
	"github.com/stretchr/testify/require"
)

func TestParseBuildArgs(t *testing.T) {
	opts := map[string]string{
		"build-arg:FOO": "bar",
		"platform":      "linux/amd64",
		"build-arg:BAZ": "qux",
		"filename":      "Dockerfile",
	}

	got := parseBuildArgs(opts)

	want := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	if len(got) != len(want) {
		t.Errorf("got %d build args, want %d", len(got), len(want))
	}

	for k, v := range want {
		if got[k] != v {
			t.Errorf("build arg %q = %q, want %q", k, got[k], v)
		}
	}
}

func TestParseCacheImports(t *testing.T) {
	t.Parallel()

	t.Run("empty opts", func(t *testing.T) {
		imports, err := parseCacheImports(nil)
		require.NoError(t, err)
		require.Empty(t, imports)
	})

	t.Run("cache-imports JSON", func(t *testing.T) {
		imports, err := parseCacheImports(map[string]string{
			keyCacheImports: `[{"Type":"registry","Attrs":{"ref":"host.docker.internal:7890/node-bun:cache"}}]`,
		})
		require.NoError(t, err)
		require.Len(t, imports, 1)
		require.Equal(t, "registry", imports[0].Type)
		require.Equal(t, "host.docker.internal:7890/node-bun:cache", imports[0].Attrs["ref"])
	})

	t.Run("multiple cache-imports", func(t *testing.T) {
		imports, err := parseCacheImports(map[string]string{
			keyCacheImports: `[{"Type":"registry","Attrs":{"ref":"a:cache"}},{"Type":"gha","Attrs":{"scope":"my-scope"}}]`,
		})
		require.NoError(t, err)
		require.Len(t, imports, 2)
		require.Equal(t, "registry", imports[0].Type)
		require.Equal(t, "a:cache", imports[0].Attrs["ref"])
		require.Equal(t, "gha", imports[1].Type)
		require.Equal(t, "my-scope", imports[1].Attrs["scope"])
	})

	t.Run("invalid cache-imports JSON", func(t *testing.T) {
		_, err := parseCacheImports(map[string]string{
			keyCacheImports: `not-json`,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), keyCacheImports)
	})
}

func TestParseGitContext(t *testing.T) {
	tests := []struct {
		name           string
		context        string
		wantIdentifier string
	}{
		{
			name:           "commit_sha",
			context:        "https://github.com/org/repo.git#0123456789abcdef0123456789abcdef01234567",
			wantIdentifier: "git://github.com/org/repo.git#0123456789abcdef0123456789abcdef01234567",
		},
		{
			name:           "ref_and_subdir",
			context:        "https://github.com/org/repo.git#main:apps/api",
			wantIdentifier: "git://github.com/org/repo.git#main:apps/api",
		},
		{
			name:           "pull_request_ref",
			context:        "https://github.com/org/repo.git#refs/pull/42/head",
			wantIdentifier: "git://github.com/org/repo.git#refs/pull/42/head",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, err := parseGitContext(map[string]string{keyContext: tt.context})
			require.NoError(t, err)
			require.NotNil(t, st, "parseGitContext(%q) returned no git source", tt.context)

			sources := sourceOps(marshalOps(t, *st))
			require.Len(t, sources, 1)
			require.Equal(t, tt.wantIdentifier, sources[0].Identifier, "parseGitContext(%q) source", tt.context)
		})
	}
}

func TestParseGitContextKeepGitDir(t *testing.T) {
	tests := []struct {
		name        string
		keepGitDir  string
		wantKeepGit string
	}{
		{name: "unset", keepGitDir: "", wantKeepGit: ""},
		{name: "true", keepGitDir: "true", wantKeepGit: "true"},
		{name: "false", keepGitDir: "false", wantKeepGit: ""},
		{name: "invalid", keepGitDir: "yes", wantKeepGit: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, err := parseGitContext(map[string]string{
				keyContext:              "https://github.com/org/repo.git#main",
				keyContextKeepGitDirArg: tt.keepGitDir,
			})
			require.NoError(t, err)
			require.NotNil(t, st)

			sources := sourceOps(marshalOps(t, *st))
			require.Len(t, sources, 1)
			require.Equal(t, tt.wantKeepGit, sources[0].Attrs[pb.AttrKeepGitDir], "parseGitContext with %s=%q", keyContextKeepGitDirArg, tt.keepGitDir)
		})
	}
}

func TestParseGitContextNotGit(t *testing.T) {
	for _, context := range []string{"", ".", "./app", "some-local-dir"} {
		st, err := parseGitContext(map[string]string{keyContext: context})
		require.NoError(t, err, "parseGitContext(%q)", context)
		require.Nil(t, st, "parseGitContext(%q) returned a git source, want nil", context)
	}
}
