package node

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReactRouterSPAConfig(t *testing.T) {
	tests := []struct {
		name   string
		config string
		isSPA  bool
	}{
		{
			name:   "multiline config",
			config: "export default {\n  ssr: false,\n} satisfies Config;",
			isSPA:  true,
		},
		{
			name:   "inline config",
			config: "export default { ssr: false } satisfies Config;",
			isSPA:  true,
		},
		{
			name:   "quoted property",
			config: "export default { 'ssr': false } satisfies Config;",
			isSPA:  true,
		},
		{
			name:   "server rendering enabled",
			config: "export default { ssr: true } satisfies Config;",
			isSPA:  false,
		},
		{
			name:   "server rendering omitted",
			config: "export default {} satisfies Config;",
			isSPA:  false,
		},
		{
			name:   "dynamic server rendering",
			config: "export default { ssr: process.env.SSR === 'true' } satisfies Config;",
			isSPA:  false,
		},
		{
			name:   "commented setting",
			config: "export default {\n  // ssr: false,\n} satisfies Config;",
			isSPA:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isSPA, reactRouterSPAConfigRegex.MatchString(tt.config))
		})
	}
}
