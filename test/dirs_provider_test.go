package test

import (
	"runtime"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestDirsProvider_UserHomeDir(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)
	require.NotNil(t, provider)
	require.NotNil(t, provider.UserHomeDirFunc)

	homeDir, err := provider.UserHomeDirFunc()
	require.NoError(t, err)

	// Should be prefixed with testRoot
	assert.Contains(t, string(homeDir), string(testRoot), "UserHomeDir should contain testRoot")

	// Should contain username
	assert.Contains(t, string(homeDir), "testuser", "UserHomeDir should contain username")

	// Should be OS-appropriate
	switch runtime.GOOS {
	case "darwin":
		assert.Contains(t, string(homeDir), "/Users/testuser")
	case "windows":
		assert.Contains(t, string(homeDir), "Users\\testuser")
	default: // Unix/Linux
		assert.Contains(t, string(homeDir), "/home/testuser")
	}
}

func TestNewTestDirsProvider_UserConfigDir(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)
	require.NotNil(t, provider)
	require.NotNil(t, provider.UserConfigDirFunc)

	userDir, err := provider.UserConfigDirFunc()
	require.NoError(t, err)

	// Should be prefixed with testRoot
	assert.Contains(t, string(userDir), string(testRoot), "UserConfigDir should contain testRoot")

	require.NotNil(t, provider.CLIConfigDirFunc)

	cliDir, err := provider.CLIConfigDirFunc()
	require.NoError(t, err)

	// Should be prefixed with testRoot
	assert.Contains(t, string(cliDir), string(testRoot), "CLIConfigDirType should contain testRoot")
	assert.Contains(t, string(cliDir), "/testuser/.config", "CLIConfigDirType should end with '/testuser/.config'")
	// Should be OS-appropriate
	switch runtime.GOOS {
	case "darwin":
		assert.Contains(t, string(userDir), "/Users/testuser/Library/Application Support")
	case "windows":
		assert.Contains(t, string(userDir), "Users\\testuser\\AppData\\Roaming")
	default: // Unix/Linux
		assert.Contains(t, string(userDir), "/home/testuser/.config")
	}
}

func TestNewTestDirsProvider_ProjectDir(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)
	require.NotNil(t, provider)
	require.NotNil(t, provider.ProjectDirFunc)

	projectDir, err := provider.ProjectDirFunc()
	require.NoError(t, err)

	// Should be prefixed with testRoot
	assert.Contains(t, string(projectDir), string(testRoot), "ProjectDir should contain testRoot")

	// Should contain the project directory name
	assert.Contains(t, string(projectDir), "testproject", "ProjectDir should contain project directory name")
}

func TestNewTestDirsProvider_Getwd(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)
	require.NotNil(t, provider)
	require.NotNil(t, provider.GetwdFunc)

	wd, err := provider.GetwdFunc()
	require.NoError(t, err)

	// Should be prefixed with testRoot
	assert.Contains(t, string(wd), string(testRoot), "Getwd should contain testRoot")
}

func TestTestDirsProviderArgs_GetTestRoot(t *testing.T) {
	testRoot := dt.DirPathJoin(dt.TempDir(), "test-root")

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	// Test normal case - should prefix with testRoot
	inputPath := dt.DirPath("/some/path")
	result := args.GetTestRoot(inputPath)
	assert.Contains(t, string(result), string(testRoot))
	assert.Contains(t, string(result), "/some/path")
}

func TestTestDirsProviderArgs_WithoutTestRoot(t *testing.T) {
	testRoot := dt.DirPath("/tmp/test-root")

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	// Function that uses GetTestRoot
	testFunc := func() (dt.DirPath, error) {
		return args.GetTestRoot("/some/path"), nil
	}

	// Call without testRoot - should NOT prefix
	result, err := args.WithoutTestRoot(testFunc)
	require.NoError(t, err)

	// Result should be the unprefixed path
	assert.Equal(t, "/some/path", string(result))
}

func TestTestDirsProviderArgs_TestRootFunc(t *testing.T) {
	// Test using TestRootFunc instead of TestRoot
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRootFunc: func() dt.DirPath {
			return testRoot
		},
	}

	provider := cstest.NewTestDirsProvider(args)
	homeDir, err := provider.UserHomeDirFunc()
	require.NoError(t, err)

	// Should use TestRootFunc to get testRoot
	assert.Contains(t, string(homeDir), string(testRoot))
}

func TestTestDirsProviderArgs_RelConfigDir(t *testing.T) {
	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   dt.DirPathJoin(dt.TempDir(), "test"),
	}

	relPath := args.RelConfigDir()

	// Should be a relative path with username and config slug
	assert.Contains(t, string(relPath), "testuser")
	assert.Contains(t, string(relPath), "myapp")
	assert.Contains(t, string(relPath), ".config")
}

func TestNewTestDirsProvider_EmptyUsername(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "", // Empty username should cause error
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)

	// Should return error for empty username
	_, err := provider.UserHomeDirFunc()
	assert.Error(t, err, "Empty username should cause error")
}

func TestNewTestDirsProvider_InvalidUsername(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "user/with/slashes", // Invalid username
		ProjectDir: "testproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)

	// Should return error for invalid username
	_, err := provider.UserHomeDirFunc()
	assert.Error(t, err, "Username with slashes should cause error")
}

func TestNewTestDirsProvider_EmptyProjectDir(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "", // Empty project dir should cause error
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	provider := cstest.NewTestDirsProvider(args)

	// Should return error for empty project dir
	_, err := provider.ProjectDirFunc()
	assert.Error(t, err, "Empty project directory should cause error")
}
