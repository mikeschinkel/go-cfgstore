package cfgstore_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Name string
	Age  int
}

func TestConfigStore_SaveLoadExists(t *testing.T) {
	var err error

	testRoot := dt.DirPathJoin(os.TempDir(), "xmlui-cli-"+uuid.NewString())
	cs, _ := getConfigStore("config/testdata.json", testRoot, cfgstore.DefaultConfigDirType)

	t.Cleanup(cleanupFunc(t, cs))

	data := testData{Name: "Alice", Age: 42}

	err = cs.SaveJSON(&data)
	require.NoError(t, err)

	exists := cs.Exists()
	assert.True(t, exists)

	var loaded testData
	err = cs.LoadJSON(&loaded)
	require.NoError(t, err)
	assert.Equal(t, data, loaded)
}

func TestConfigStore_LoadNonexistent(t *testing.T) {
	var err error

	testRoot := dtx.TempTestDir(t)
	cs, _ := getConfigStore("does-not-exist.json", testRoot, cfgstore.DefaultConfigDirType)
	t.Cleanup(cleanupFunc(t, cs))

	err = cs.LoadJSON(&testData{})
	assert.Error(t, err)

}

func TestConfigStore_SaveInvalidJSON(t *testing.T) {
	var err error

	testRoot := dtx.TempTestDir(t)
	cs, _ := getConfigStore("bad.json", testRoot, cfgstore.DefaultConfigDirType)
	t.Cleanup(cleanupFunc(t, cs))

	ch := make(chan int) // non-serializable
	err = cs.SaveJSON(ch)
	assert.Error(t, err)
}

func TestConfigStore_ConfigDir(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	cs, args := getConfigStore("", testRoot, cfgstore.DefaultConfigDirType)
	t.Cleanup(cleanupFunc(t, cs))

	cfgDir, err := cs.ConfigDir()
	assert.NoError(t, err)
	rel, err := cfgDir.Rel(testRoot)
	assert.NoError(t, err)
	assert.Equal(t, rel, args.RelConfigDir())
}

func TestConfigStores_CLIAndProjectStores(t *testing.T) {
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	args := &cstest.TestDirsProviderArgs{
		Username:   "testuser",
		ProjectDir: "myproject",
		ConfigSlug: "myapp",
		TestRoot:   testRoot,
	}

	// Create both CLI and project stores
	stores := cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
		ConfigStoreArgs: cfgstore.ConfigStoreArgs{
			ConfigSlug:   "myapp",
			RelFilepath:  "config.json",
			DirsProvider: cstest.NewTestDirsProvider(args),
		},
	})

	cliStore := stores.CLIConfigStore()
	projectStore := stores.ProjectConfigStore()

	// Both stores should work
	assert.NotNil(t, cliStore)
	assert.NotNil(t, projectStore)

	// Get their directories - should be different
	cliDir, err := cliStore.ConfigDir()
	require.NoError(t, err)

	projectDir, err := projectStore.ConfigDir()
	require.NoError(t, err)

	assert.NotEqual(t, cliDir, projectDir)

	// Both should be under testRoot
	assert.Contains(t, string(cliDir), string(testRoot))
	assert.Contains(t, string(projectDir), string(testRoot))
}

func cleanupFunc(t *testing.T, cs cfgstore.ConfigStore) func() {
	return func() {
		cstest.RemoveAll(t, cs)
	}
}
