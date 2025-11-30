package test

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestConfigStore_LoadFromPreCreatedFixture(t *testing.T) {
	var err error

	// Create temp test root directory
	testRoot := dtx.TempTestDir(t)
	defer cfgstore.LogOnError(testRoot.RemoveAll())

	// Setup test configuration
	username := dt.PathSegment("testuser")
	configSlug := dt.PathSegment("myapp")
	configFile := "config.json"

	args := &cstest.TestDirsProviderArgs{
		Username:   username,
		ProjectDir: "testproject",
		ConfigSlug: configSlug,
		TestRoot:   testRoot,
	}

	// Manually create the fixture directory structure that TestDirsProvider would use
	// For CLI config: <testRoot>/Users/<username>/.config/<configSlug>/
	configDirPath := filepath.Join(string(testRoot), "Users", string(username), ".config", string(configSlug))
	err = os.MkdirAll(configDirPath, 0755)
	require.NoError(t, err, "Failed to create config directory")

	// Write a fixture JSON file with known content using standard library
	expectedData := testData{Name: "Bob", Age: 30}
	jsonBytes, err := json.Marshal(expectedData)
	require.NoError(t, err, "Failed to marshal test data")

	fixtureFilePath := filepath.Join(configDirPath, configFile)
	err = os.WriteFile(fixtureFilePath, jsonBytes, 0644)
	require.NoError(t, err, "Failed to write fixture file")

	// Verify the file exists using standard library
	_, err = os.Stat(fixtureFilePath)
	require.NoError(t, err, "Fixture file should exist at %s", fixtureFilePath)

	// Create ConfigStore with TestDirsProvider pointing to testRoot
	cs := cfgstore.NewConfigStore(cfgstore.DefaultConfigDirType, cfgstore.ConfigStoreArgs{
		ConfigSlug:   configSlug,
		RelFilepath:  dt.RelFilepath(configFile),
		DirsProvider: cstest.NewTestDirsProvider(args),
	})

	// Verify ConfigStore sees the correct directory
	actualConfigDir, err := cs.ConfigDir()
	require.NoError(t, err, "ConfigStore.ConfigDir() failed")
	assert.Equal(t, dt.DirPath(configDirPath), actualConfigDir, "ConfigDir should match expected path")

	// Call LoadJSON() to load the pre-created fixture
	var loaded testData
	err = cs.LoadJSON(&loaded)
	require.NoError(t, err, "LoadJSON should successfully load the pre-created fixture")

	// Verify the data loaded correctly
	assert.Equal(t, expectedData, loaded, "Loaded data should match fixture data")
}

func cleanupFunc(t *testing.T, cs cfgstore.ConfigStore) func() {
	return func() {
		RemoveAll(t, cs)
	}
}
