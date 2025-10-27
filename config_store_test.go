package cfgstore_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Name string
	Age  int
}

func getConfigStore(fp dt.RelFilepath, dirType cfgstore.DirType) cfgstore.ConfigStore {
	args := cfgstore.ConfigStoreArgs{
		AppInfo: appinfo.New(appinfo.Args{
			AppName:    "",
			AppDescr:   "",
			AppVer:     "",
			AppSlug:    "",
			ConfigDir:  "test-app",
			ConfigFile: fp,
			ExeName:    "",
			InfoURL:    "",
		}),
	}
	return cfgstore.NewConfigStore(args).WithDirType(dirType).(cfgstore.ConfigStore)
}

func TestConfigStore_SaveLoadExists(t *testing.T) {
	var err error
	dir := dt.DirPathJoin(os.TempDir(), "xmlui-cli-"+uuid.NewString())
	t.Cleanup(func() {
		cfgstore.LogOnError(dt.RemoveAll(dir))
	})

	cs := getConfigStore("config/testdata.json", cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(dir)

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

	cs := getConfigStore("does-not-exist.json", cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(dt.TempDir(t))

	err = cs.LoadJSON(&testData{})
	assert.Error(t, err)
}

func TestConfigStore_SaveInvalidJSON(t *testing.T) {
	var err error

	cs := getConfigStore("bad.json", cfgstore.DefaultConfigDirType)
	cs.SetConfigDir(dt.TempDir(t))

	ch := make(chan int) // non-serializable
	err = cs.SaveJSON(ch)
	assert.Error(t, err)
}

func TestConfigStore_ConfigDir(t *testing.T) {
	cs := getConfigStore("", cfgstore.DefaultConfigDirType)
	dir := dt.TempDir(t)

	cs.SetConfigDir(dir)

	cfgDir, err := cs.ConfigDir()
	assert.NoError(t, err)
	assert.Equal(t, dir, cfgDir)
}
