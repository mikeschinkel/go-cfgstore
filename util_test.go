package cfgstore_test

import (
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/mikeschinkel/go-fsfix"

	"github.com/mikeschinkel/go-testutil"
)

const (
	TestConfigSlug = "acme"
)

type ConfigDirFixturesArgs struct {
	appinfo.AppInfo
	DirTypes    []cfgstore.DirType
	TestDataDir dt.DirPath
	UserFile    dt.Filepath
	ProjectFile dt.Filepath
}

func getConfigStore(fp dt.RelFilepath, testRoot dt.DirPath, dirType cfgstore.DirType) (cfgstore.ConfigStore, *cstest.TestDirsProviderArgs) {
	args := &cstest.TestDirsProviderArgs{
		Username:   "coyote",
		ProjectDir: "billboard",
		ConfigSlug: TestConfigSlug,
		TestRoot:   testRoot,
	}
	cs := cfgstore.NewConfigStore(dirType, cfgstore.ConfigStoreArgs{
		ConfigSlug:   TestConfigSlug,
		RelFilepath:  fp,
		DirsProvider: cstest.NewTestDirsProvider(args),
	})
	return cs, args
}

// SetupConfigDirFixtures sets up a root fixture with two dir fixtures, one to
// emulate the user's ~/.config/<path> config directory and the other to emulate
// the project's ./.<path> config directory. The userFile and projectFile should
// be filenames containing hte respective config files for each.
func SetupConfigDirFixtures(t *testing.T, testDataDir dt.DirPath, userFile, projectFile dt.Filepath, args ConfigDirFixturesArgs) (rootFix *fsfix.RootFixture, css *cfgstore.ConfigStores) {

	configFile := args.ConfigFile()
	rootFix = fsfix.NewRootFixture("cfgstore")
	testArgs := &cstest.TestDirsProviderArgs{
		Username:   "coyote",
		ProjectDir: "billboard",
		ConfigSlug: "acme",
		TestRootFunc: func() dt.DirPath {
			return rootFix.Dir()
		},
	}
	css = cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
		ConfigStoreArgs: cfgstore.ConfigStoreArgs{
			ConfigSlug:   "acme",
			RelFilepath:  configFile,
			DirsProvider: cstest.NewTestDirsProvider(testArgs),
		},
	})
	cliStore := css.CLIConfigStore()
	cliDir, err := cstest.GetRelConfigDir(cliStore, testArgs)
	if err != nil {
		t.Fatal(err)
	}

	cliFix := rootFix.AddDirFixture(t, cliDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	cliFix.AddFileFixture(t, configFile, &fsfix.FileFixtureArgs{
		Name:    configFile,
		Content: string(testutil.LoadFile(t, userFile, true)),
	})

	projectStore := css.CLIConfigStore()
	projectDir, err := cstest.GetRelConfigDir(projectStore, testArgs)
	if err != nil {
		t.Fatal(err)
	}

	projectFix := rootFix.AddDirFixture(t, projectDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	projectFix.AddFileFixture(t, configFile, &fsfix.FileFixtureArgs{
		Name:    configFile,
		Content: string(testutil.LoadFile(t, projectFile, true)),
	})

	rootFix.Create(t)

	//dotCS.SetConfigDir(cliFix.Dir())
	//projectStore.SetConfigDir(projectFix.Dir())

	return rootFix, css
}
