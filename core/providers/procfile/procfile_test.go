package procfile

import (
	"testing"

	testingUtils "github.com/railwayapp/railpack/core/testing"
	"github.com/stretchr/testify/require"
)

func TestProcfile(t *testing.T) {
	ctx := testingUtils.CreateGenerateContext(t, "../../../examples/ruby-vanilla")
	provider := ProcfileProvider{}

	_, err := provider.Plan(ctx)
	require.NoError(t, err)

	require.Equal(t, "ruby app.rb", ctx.Deploy.StartCmd)
}
