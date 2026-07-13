package php

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestPhpProvider(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		isPhp     bool
		framework string
	}{
		{
			name:      "vanilla php with index.php",
			path:      "../../../examples/php-vanilla",
			isPhp:     true,
			framework: "",
		},
		{
			name:      "laravel project with composer.json",
			path:      "../../../examples/php-laravel-12-react",
			isPhp:     true,
			framework: "laravel",
		},
		{
			name:      "symfony project",
			path:      "../../../examples/php-symfony",
			isPhp:     true,
			framework: "symfony",
		},
		{
			name:      "non-php project",
			path:      "../../../examples/node-npm",
			isPhp:     false,
			framework: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testingUtils.CreateGenerateContext(t, tt.path)
			provider := PhpProvider{}

			isPhp, err := provider.Detect(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.isPhp, isPhp)

			require.Equal(t, tt.framework, provider.getFramework(ctx))
		})
	}
}
