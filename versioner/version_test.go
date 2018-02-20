package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testVersionResponse struct{}

func (u testVersionResponse) MinVersion() Version { return MustParse(defaultMinVersion) }
func (u testVersionResponse) MaxVersion() Version { return MustParse(defaultMaxVersion) }
func (u testVersionResponse) Deprecated() bool    { return false }

func TestVersionedResponseMinVersion(t *testing.T) {
	x := testVersionResponse{}
	smallVer := MustParse("0.1")
	reqVer := GetMinVersionFor(x)

	require.False(t, reqVer.LessThan(smallVer.Version))
}

func TestVersionedResponseMaxVersion(t *testing.T) {
	x := testVersionResponse{}
	bigVer := MustParse("6.0.0")
	reqVer := GetMaxVersionFor(x)

	require.False(t, reqVer.GreaterThan(bigVer.Version))
}

func TestVersionedDeprecated(t *testing.T) {
	x := testVersionResponse{}

	require.False(t, IsDeprecated(x))
}

func TestGenericVersioner(t *testing.T) {
	x := NewGenericVersioner("1.0.1", "1.9", true)
	require.NotNil(t, x)
	bigVer := "2.0"
	smallVer := "0.1"
	dep := IsDeprecated(x)
	require.True(t, dep)

	require.Error(t, CheckSupportedVersion(x, bigVer))
	require.Error(t, CheckSupportedVersion(x, smallVer))
	require.NoError(t, CheckSupportedVersion(x, "1.1"))
	require.NoError(t, CheckSupportedVersion(x, "1.0.1"))
	require.NoError(t, CheckSupportedVersion(x, "1.9"))
	require.Error(t, CheckSupportedVersion(x, "foobar"))
}

func TestMustParse(t *testing.T) {
	x := func() { _ = MustParse("foobar") }
	require.Panics(t, x)
}
