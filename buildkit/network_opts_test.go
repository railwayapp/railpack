package buildkit

import (
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/solver/pb"
	"github.com/stretchr/testify/require"
)

func TestParseNetMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    pb.NetMode
		wantErr bool
	}{
		{name: "empty defaults to sandbox", input: "", want: llb.NetModeSandbox},
		{name: "sandbox", input: "sandbox", want: llb.NetModeSandbox},
		{name: "host", input: "host", want: llb.NetModeHost},
		{name: "none", input: "none", want: llb.NetModeNone},
		{name: "invalid", input: "bridge", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNetMode(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseExtraHosts(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		got, err := parseExtraHosts("")
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("single host", func(t *testing.T) {
		got, err := parseExtraHosts("my-other-container=192.168.107.2")
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, "my-other-container", got[0].Host)
		require.Equal(t, "192.168.107.2", got[0].IP.String())
	})

	t.Run("lowercases host", func(t *testing.T) {
		got, err := parseExtraHosts("My-Host=10.0.0.1")
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, "my-host", got[0].Host)
		require.Equal(t, "10.0.0.1", got[0].IP.String())
	})

	t.Run("multiple hosts", func(t *testing.T) {
		got, err := parseExtraHosts("a=1.1.1.1,b=2.2.2.2")
		require.NoError(t, err)
		require.Len(t, got, 2)
		require.Equal(t, "a", got[0].Host)
		require.Equal(t, "1.1.1.1", got[0].IP.String())
		require.Equal(t, "b", got[1].Host)
		require.Equal(t, "2.2.2.2", got[1].IP.String())
	})

	t.Run("invalid pair", func(t *testing.T) {
		_, err := parseExtraHosts("not-a-pair")
		require.Error(t, err)
	})

	t.Run("invalid IP", func(t *testing.T) {
		_, err := parseExtraHosts("host=not-an-ip")
		require.Error(t, err)
	})
}
