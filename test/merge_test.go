package test

import (
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mergeConfigFile = "config.json"

// mergeConfig is a test config with no defaults whose Merge is the common
// one-line Decode.
type mergeConfig struct {
	CLIOnly     string `json:"cli_only,omitempty"`
	ProjectOnly string `json:"project_only,omitempty"`
	Both        string `json:"both,omitempty"`
}

func (c *mergeConfig) RootConfig()                                       {}
func (c *mergeConfig) ApplyDefaults(cfgstore.DefaultsArgs) error         { return nil }
func (c *mergeConfig) Normalize(cfgstore.NormalizeArgs) error            { return nil }
func (c *mergeConfig) Validate(cfgstore.ValidateArgs) error              { return nil }
func (c *mergeConfig) Merge(args cfgstore.MergeArgs[*mergeConfig]) error { return args.Decode(c) }

// mergeFixture is a test root with a DirsProvider rooted in it.
type mergeFixture struct {
	dirsProvider *cfgstore.DirsProvider
}

func newMergeFixture(t *testing.T) *mergeFixture {
	t.Helper()
	testRoot := dtx.TempTestDir(t)
	t.Cleanup(func() { cfgstore.LogOnError(testRoot.RemoveAll()) })
	return &mergeFixture{
		dirsProvider: cstest.NewTestDirsProvider(&cstest.TestDirsProviderArgs{
			Username:   "coyote",
			ProjectDir: "billboard",
			ConfigSlug: TestConfigSlug,
			TestRoot:   testRoot,
		}),
	}
}

// stores returns config stores for dirTypes, rooted in the fixture.
func (f *mergeFixture) stores(dirTypes ...cfgstore.DirType) *cfgstore.ConfigStores {
	return cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
		DirTypes: dirTypes,
		ConfigStoreArgs: cfgstore.ConfigStoreArgs{
			ConfigSlug:   TestConfigSlug,
			RelFilepath:  mergeConfigFile,
			DirsProvider: f.dirsProvider,
		},
	})
}

// writeRaw saves text as the config file of dirType's layer, so a test controls
// exactly which keys the file holds.
func (f *mergeFixture) writeRaw(t *testing.T, dirType cfgstore.DirType, text string) {
	t.Helper()
	cs := f.stores(dirType).StoreMap[dirType]
	require.NoError(t, cs.Save([]byte(text)), "writing %s config", dirType)
}

// filepath returns the config file of dirType's layer.
func (f *mergeFixture) filepath(t *testing.T, dirType cfgstore.DirType) dt.Filepath {
	t.Helper()
	fp, err := f.stores(dirType).StoreMap[dirType].GetFilepath()
	require.NoError(t, err, "getting %s config filepath", dirType)
	return fp
}

// write saves data as the config file of dirType's layer.
func (f *mergeFixture) write(t *testing.T, dirType cfgstore.DirType, data any) {
	t.Helper()
	cs := f.stores(dirType).StoreMap[dirType]
	require.NoError(t, cs.SaveJSON(data), "writing %s config", dirType)
}

func (f *mergeFixture) loadArgs(dirTypes ...cfgstore.DirType) cfgstore.LoadConfigArgs {
	return cfgstore.LoadConfigArgs{
		ConfigSlug:   TestConfigSlug,
		ConfigFile:   mergeConfigFile,
		DirTypes:     dirTypes,
		DirsProvider: f.dirsProvider,
	}
}

func TestLoad_CLIAndProjectLayersMerge(t *testing.T) {
	cli := &mergeConfig{CLIOnly: "from-cli", Both: "cli"}
	project := &mergeConfig{ProjectOnly: "from-project", Both: "project"}
	want := &mergeConfig{CLIOnly: "from-cli", ProjectOnly: "from-project", Both: "project"}

	loaders := map[string]func(f *mergeFixture) (*mergeConfig, error){
		"LoadConfig": func(f *mergeFixture) (*mergeConfig, error) {
			return cfgstore.LoadConfig[mergeConfig](f.loadArgs(cfgstore.CLIConfigDirType, cfgstore.ProjectConfigDirType))
		},
		"LoadDefaultConfig": func(f *mergeFixture) (*mergeConfig, error) {
			return cfgstore.LoadDefaultConfig[mergeConfig](f.loadArgs())
		},
		"LoadConfigStores": func(f *mergeFixture) (*mergeConfig, error) {
			dirTypes := []cfgstore.DirType{cfgstore.CLIConfigDirType, cfgstore.ProjectConfigDirType}
			return cfgstore.LoadConfigStores[mergeConfig](f.stores(dirTypes...), cfgstore.RootConfigArgs{
				DirTypes:     dirTypes,
				DirsProvider: f.dirsProvider,
			})
		},
	}
	for name, load := range loaders {
		t.Run(name, func(t *testing.T) {
			f := newMergeFixture(t)
			f.write(t, cfgstore.CLIConfigDirType, cli)
			f.write(t, cfgstore.ProjectConfigDirType, project)

			got, err := load(f)
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func assertErrValue[T any](t *testing.T, err error, key string, want T) {
	t.Helper()
	got, ok := cfgstore.ErrValue[T](err, key)
	if assert.True(t, ok, "error has no %T value for %q: %v", want, key, err) {
		assert.Equal(t, want, got, "error value for %q", key)
	}
}
