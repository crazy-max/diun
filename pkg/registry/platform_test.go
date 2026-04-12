package registry

import (
	"testing"

	regplatform "github.com/regclient/regclient/types/platform"
	"github.com/stretchr/testify/require"
)

func TestTargetPlatformDefault(t *testing.T) {
	got, err := TargetPlatform("", "", "")
	require.NoError(t, err)
	require.Equal(t, regplatform.Local(), got)
}

func TestTargetPlatformOverride(t *testing.T) {
	got, err := TargetPlatform("linux", "arm", "v7")
	require.NoError(t, err)
	require.Equal(t, regplatform.Platform{
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v7",
	}, got)
}

func TestTargetPlatformInvalid(t *testing.T) {
	_, err := TargetPlatform("linux!", "", "")
	require.Error(t, err)
}
