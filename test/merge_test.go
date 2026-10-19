package test

import (
	"errors"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-dt/dtx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mergeConfigFile = "config.json"

// mergeConfig is a test config whose Merge gives the receiver's values
// precedence and returns a copy.
type mergeConfig struct {
	CLIOnly     string `json:"cli_only,omitempty"`
	ProjectOnly string `json:"project_only,omitempty"`
	Both        string `json:"both,omitempty"`
}

func (c *mergeConfig) RootConfig()                            {}
func (c *mergeConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *mergeConfig) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {
	result := *c
	mergeFields(&result, other.(*mergeConfig))
	return &result
}

// mergeFields fills each field of dst that is empty from src.
func mergeFields(dst, src *mergeConfig) {
	if dst.CLIOnly == "" {
		dst.CLIOnly = src.CLIOnly
	}
	if dst.ProjectOnly == "" {
		dst.ProjectOnly = src.ProjectOnly
	}
	if dst.Both == "" {
		dst.Both = src.Both
	}
}

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

// inPlaceConfig's Merge edits its receiver and returns it.
type inPlaceConfig mergeConfig

func (c *inPlaceConfig) RootConfig()                            {}
func (c *inPlaceConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *inPlaceConfig) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {
	mergeFields((*mergeConfig)(c), (*mergeConfig)(other.(*inPlaceConfig)))
	return c
}

// recordedMerges holds what each recordingConfig.Merge received as other.
var recordedMerges []recordingConfig

// recordingConfig's Merge records a copy of what it received, then merges as
// mergeConfig does.
type recordingConfig mergeConfig

func (c *recordingConfig) RootConfig()                            {}
func (c *recordingConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *recordingConfig) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {
	o := other.(*recordingConfig)
	recordedMerges = append(recordedMerges, *o)
	result := *c
	mergeFields((*mergeConfig)(&result), (*mergeConfig)(o))
	return &result
}

// wrongTypeConfig's Merge returns a concrete type other than its own.
type wrongTypeConfig mergeConfig

func (c *wrongTypeConfig) RootConfig()                            {}
func (c *wrongTypeConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *wrongTypeConfig) Merge(cfgstore.RootConfig) cfgstore.RootConfig {
	return (*mergeConfig)(c)
}

// nilConfig's Merge returns a nil RootConfig.
type nilConfig mergeConfig

func (c *nilConfig) RootConfig()                                   {}
func (c *nilConfig) Normalize(cfgstore.NormalizeArgs) error        { return nil }
func (c *nilConfig) Merge(cfgstore.RootConfig) cfgstore.RootConfig { return nil }

// nilPtrConfig's Merge returns a nil pointer of its own type.
type nilPtrConfig mergeConfig

func (c *nilPtrConfig) RootConfig()                            {}
func (c *nilPtrConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *nilPtrConfig) Merge(cfgstore.RootConfig) cfgstore.RootConfig {
	return (*nilPtrConfig)(nil)
}

func TestLoad_EachMergeReceivesTheMergeOfEarlierLayers(t *testing.T) {
	recordedMerges = nil
	f := newMergeFixture(t)
	f.write(t, cfgstore.AppConfigDirType, &recordingConfig{CLIOnly: "from-app", Both: "app"})
	f.write(t, cfgstore.CLIConfigDirType, &recordingConfig{ProjectOnly: "from-cli", Both: "cli"})
	f.write(t, cfgstore.ProjectConfigDirType, &recordingConfig{Both: "project"})

	got, err := cfgstore.LoadConfig[recordingConfig](f.loadArgs(
		cfgstore.AppConfigDirType,
		cfgstore.CLIConfigDirType,
		cfgstore.ProjectConfigDirType,
	))
	require.NoError(t, err)

	assert.Equal(t, []recordingConfig{
		// The CLI layer's Merge receives the app layer as loaded
		{CLIOnly: "from-app", Both: "app"},
		// The project layer's Merge receives the app and CLI layers merged
		{CLIOnly: "from-app", ProjectOnly: "from-cli", Both: "cli"},
	}, recordedMerges)
	assert.Equal(t, &recordingConfig{CLIOnly: "from-app", ProjectOnly: "from-cli", Both: "project"}, got)
}

func TestLoad_InPlaceMergeMatchesCopyingMerge(t *testing.T) {
	cli := mergeConfig{CLIOnly: "from-cli", Both: "cli"}
	project := mergeConfig{ProjectOnly: "from-project", Both: "project"}

	f := newMergeFixture(t)
	f.write(t, cfgstore.CLIConfigDirType, &cli)
	f.write(t, cfgstore.ProjectConfigDirType, &project)
	copied, err := cfgstore.LoadDefaultConfig[mergeConfig](f.loadArgs())
	require.NoError(t, err)

	f = newMergeFixture(t)
	f.write(t, cfgstore.CLIConfigDirType, (*inPlaceConfig)(&cli))
	f.write(t, cfgstore.ProjectConfigDirType, (*inPlaceConfig)(&project))
	inPlace, err := cfgstore.LoadDefaultConfig[inPlaceConfig](f.loadArgs())
	require.NoError(t, err)

	assert.Equal(t, copied, (*mergeConfig)(inPlace))
}

func TestLoad_MergeReturningAnotherTypeFails(t *testing.T) {
	f := newMergeFixture(t)
	f.write(t, cfgstore.CLIConfigDirType, &wrongTypeConfig{CLIOnly: "from-cli"})
	f.write(t, cfgstore.ProjectConfigDirType, &wrongTypeConfig{ProjectOnly: "from-project"})

	got, err := cfgstore.LoadDefaultConfig[wrongTypeConfig](f.loadArgs())
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, errors.Is(err, cfgstore.ErrFailedToLoadConfig), err.Error())
	assertErrValue(t, err, "dir_type", cfgstore.ProjectConfigDirType)
	assertErrValue(t, err, "expected_type", "*test.wrongTypeConfig")
	assertErrValue(t, err, "returned_type", "*test.mergeConfig")
}

func TestLoad_MergeReturningNilFails(t *testing.T) {
	t.Run("nil RootConfig", func(t *testing.T) {
		f := newMergeFixture(t)
		f.write(t, cfgstore.CLIConfigDirType, &nilConfig{CLIOnly: "from-cli"})
		f.write(t, cfgstore.ProjectConfigDirType, &nilConfig{ProjectOnly: "from-project"})

		got, err := cfgstore.LoadDefaultConfig[nilConfig](f.loadArgs())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, cfgstore.ErrFailedToLoadConfig), err.Error())
		assertErrValue(t, err, "dir_type", cfgstore.ProjectConfigDirType)
		assertErrValue(t, err, "returned_type", "nil")
	})
	t.Run("nil pointer of the config type", func(t *testing.T) {
		f := newMergeFixture(t)
		f.write(t, cfgstore.CLIConfigDirType, &nilPtrConfig{CLIOnly: "from-cli"})
		f.write(t, cfgstore.ProjectConfigDirType, &nilPtrConfig{ProjectOnly: "from-project"})

		got, err := cfgstore.LoadDefaultConfig[nilPtrConfig](f.loadArgs())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, errors.Is(err, cfgstore.ErrFailedToLoadConfig), err.Error())
		assertErrValue(t, err, "dir_type", cfgstore.ProjectConfigDirType)
		assertErrValue(t, err, "returned_type", "nil")
	})
}

func TestLoad_SingleLayerIsReturnedWithoutMerge(t *testing.T) {
	cli := &recordingConfig{CLIOnly: "from-cli", Both: "cli"}

	t.Run("one DirType", func(t *testing.T) {
		recordedMerges = nil
		f := newMergeFixture(t)
		f.write(t, cfgstore.CLIConfigDirType, cli)

		got, err := cfgstore.LoadCLIConfig[recordingConfig](f.loadArgs())
		require.NoError(t, err)
		assert.Equal(t, cli, got)
		assert.Empty(t, recordedMerges, "Merge should not be called")
	})
	t.Run("no project config file", func(t *testing.T) {
		recordedMerges = nil
		f := newMergeFixture(t)
		f.write(t, cfgstore.CLIConfigDirType, cli)

		got, err := cfgstore.LoadDefaultConfig[recordingConfig](f.loadArgs())
		require.NoError(t, err)
		assert.Equal(t, cli, got)
		assert.Empty(t, recordedMerges, "Merge should not be called")
	})
}

func assertErrValue[T any](t *testing.T, err error, key string, want T) {
	t.Helper()
	got, ok := cfgstore.ErrValue[T](err, key)
	if assert.True(t, ok, "error has no %T value for %q: %v", want, key, err) {
		assert.Equal(t, want, got, "error value for %q", key)
	}
}
