package test

import (
	"runtime"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIConfigDir_PerOS checks the CLI config dir on the OS running the test:
// the user config dir plus the slug on Linux, with no second .config, and
// <home>/.config plus the slug everywhere else. The home and user config dirs
// are unrelated temp dirs, so a result built from the wrong one shows.
func TestCLIConfigDir_PerOS(t *testing.T) {
	home := dt.DirPath(t.TempDir())
	userConfig := dt.DirPath(t.TempDir())

	dp := &cfgstore.DirsProvider{
		UserHomeDirFunc:   func() (dt.DirPath, error) { return home, nil },
		UserConfigDirFunc: func() (dt.DirPath, error) { return userConfig, nil },
	}
	dp.CLIConfigDirFunc = dp.CLIConfigDirType

	got, err := cfgstore.CLIConfigDir("myapp", dp)
	require.NoError(t, err)

	var want dt.DirPath
	switch runtime.GOOS {
	case "linux":
		want = dt.DirPathJoin(userConfig, "myapp")
	default:
		want = dt.DirPathJoin3(home, cfgstore.DotConfigPathSegment, "myapp")
	}
	assert.Equal(t, want, got)
}
