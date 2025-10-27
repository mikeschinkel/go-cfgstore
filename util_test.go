package cfgstore_test

import (
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/mikeschinkel/go-fsfix"

	"github.com/mikeschinkel/go-testutil"
)

type ConfigDirFixturesArgs struct {
	appinfo.AppInfo
	DirTypes []cfgstore.DirType
}

// SetupConfigDirFixtures sets up a root fixture with two dir fixtures, one to
// emulate the user's ~/.config/xmlui config directory and the other to emulate
// the project's ./.xmlui config directory. The userFile and projectFile should
// be filenames containing hte respective config files for each.
func SetupConfigDirFixtures(t *testing.T, testDataDir, userFile, projectFile string, args ConfigDirFixturesArgs) (rootFix *fsfix.RootFixture, css *cfgstore.ConfigStores) {
	const (
		userDir    = ".config"
		projectDir = "project"
	)
	configFile := args.ConfigFile()
	css = cfgstore.NewConfigStores(cfgstore.ConfigStoreArgs{
		AppInfo:  args.AppInfo,
		DirTypes: args.DirTypes,
	})
	dotCS := css.StoreMap[cfgstore.DefaultConfigDirType]
	localCS := css.StoreMap[cfgstore.LocalConfigDir]

	rootFix = fsfix.NewRootFixture("config")

	dotFix := rootFix.AddDirFixture(t, userDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	dotFix.AddFileFixture(t, configFile, &fsfix.FileFixtureArgs{
		Name:    configFile,
		Content: string(testutil.LoadFile(t, userFile, true)),
	})

	localFix := rootFix.AddDirFixture(t, projectDir, &fsfix.DirFixtureArgs{Parent: rootFix})
	localFix.AddFileFixture(t, configFile, &fsfix.FileFixtureArgs{
		Name:    configFile,
		Content: string(testutil.LoadFile(t, projectFile, true)),
	})

	rootFix.Create(t)

	localCS.SetConfigDir(localFix.Dir())
	dotCS.SetConfigDir(dotFix.Dir())

	return rootFix, css
}
