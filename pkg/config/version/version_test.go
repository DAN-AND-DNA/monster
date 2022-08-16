package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SetFromString(t *testing.T) {
	r := require.New(t)

	version := Version{0, 0, 0}
	leftStr := version.SetFromString("1.0.03.2021")
	r.Equal("2021.", leftStr)

	leftStr = version.SetFromString("1.0.03")
	r.Equal("", leftStr)
	r.Equal((uint16)(1), version.X)
	r.Equal((uint16)(0), version.Y)
	r.Equal((uint16)(3), version.Z)

}

func Test_GetString(t *testing.T) {
	r := require.New(t)

	version := Version{0, 0, 0}
	leftStr := version.SetFromString("1.0.03")

	r.Equal("", leftStr)
	r.Equal("1.0.03", version.GetString())
}

func Test_Compare(t *testing.T) {
	r := require.New(t)
	version := Version{1, 3, 7}
	version1 := Version{0, 0, 0}

	r.Equal(0, Compare(MIN, version1))
	r.Equal(0, Compare(Version{1, 3, 7}, version))
	r.Equal(1, Compare(Version{1, 3, 8}, version))
	r.Equal(1, Compare(Version{2, 3, 7}, version))
	r.Equal(-1, Compare(Version{1, 2, 7}, version))
	r.Equal(-1, Compare(Version{0, 3, 7}, version))

}
