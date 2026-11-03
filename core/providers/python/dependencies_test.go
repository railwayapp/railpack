package python

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyStringMatchesRequestedExtras(t *testing.T) {
	tests := []struct {
		name                string
		dependencyString    string
		requestedDependency string
		matchesRequest      bool
	}{
		{
			name:                "multiple extras",
			dependencyString:    "fastapi[other,standard]>=0.115.12",
			requestedDependency: "fastapi[standard]",
			matchesRequest:      true,
		},
		{
			name:                "extra order and whitespace",
			dependencyString:    " FastAPI [ Other , STANDARD ] >=0.115.12 ",
			requestedDependency: "fastapi[standard,other]",
			matchesRequest:      true,
		},
		{
			name:                "extra names use package name normalization",
			dependencyString:    "package[Extra_Name]>=1",
			requestedDependency: "package[extra-name]",
			matchesRequest:      true,
		},
		{
			name:                "additional requested extra is missing",
			dependencyString:    "fastapi[standard]>=0.115.12",
			requestedDependency: "fastapi[standard,other]",
		},
		{
			name:                "extra name must match exactly",
			dependencyString:    "fastapi[standardized]>=0.115.12",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "package name must match exactly",
			dependencyString:    "other-fastapi[standard]>=1",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "plain dependency has no extras",
			dependencyString:    "fastapi>=0.115.12",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "package query does not require extras",
			dependencyString:    "fastapi[standard]>=0.115.12",
			requestedDependency: "fastapi",
			matchesRequest:      true,
		},
		{
			name:                "direct URL retains the declared extras",
			dependencyString:    "fastapi[other,standard] @ https://example.com/fastapi.whl",
			requestedDependency: "fastapi[standard]",
			matchesRequest:      true,
		},
		{
			name:                "URL text is not an extras declaration",
			dependencyString:    "fastapi @ https://example.com/[standard]/fastapi.whl",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "environment marker is not an extras declaration",
			dependencyString:    "fastapi>=0.115.12; extra == 'standard'",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "comment is not a dependency",
			dependencyString:    "# fastapi[standard]",
			requestedDependency: "fastapi[standard]",
		},
		{
			name:                "incomplete extras declaration",
			dependencyString:    "fastapi[standard",
			requestedDependency: "fastapi[standard]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.matchesRequest, dependencyStringDeclaresPackageAndExtras(tt.dependencyString, tt.requestedDependency))
		})
	}
}
