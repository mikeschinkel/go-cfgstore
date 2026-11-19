package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-cliutil/cliutiltest"
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/appinfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cliConfig is a root config that is also a cliutil.Config, with typed methods
// for ApplyConfig to read.
type cliConfig struct {
	RegionValue string `json:"region,omitempty"`
	FormatValue string `json:"format,omitempty"`
}

func (c *cliConfig) Config()                                {}
func (c *cliConfig) RootConfig()                            {}
func (c *cliConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }
func (c *cliConfig) Merge(other cfgstore.RootConfig) cfgstore.RootConfig {
	result := *c
	earlier := other.(*cliConfig)
	if result.RegionValue == "" {
		result.RegionValue = earlier.RegionValue
	}
	if result.FormatValue == "" {
		result.FormatValue = earlier.FormatValue
	}
	return &result
}

// Region returns the configured region.
func (c *cliConfig) Region() string { return c.RegionValue }

// Format returns the configured format.
func (c *cliConfig) Format() string { return c.FormatValue }

// cliFixture is a mergeFixture with the AppInfo a go-cliutil run passes.
type cliFixture struct {
	*mergeFixture
	slug dt.PathSegment
	file dt.RelFilepath
}

func newCLIFixture(t *testing.T, slug dt.PathSegment, file dt.RelFilepath) *cliFixture {
	t.Helper()
	return &cliFixture{mergeFixture: newMergeFixture(t), slug: slug, file: file}
}

func (f *cliFixture) appInfo() appinfo.AppInfo {
	return appinfo.New(appinfo.Args{
		Name:       "cfgtest",
		ExeName:    "cfgtest",
		ConfigSlug: f.slug,
		ConfigFile: f.file,
	})
}

// configFunc returns the adapter under test, rooted in the fixture.
func (f *cliFixture) configFunc(dirTypes ...cfgstore.DirType) cliutil.ConfigFunc {
	return cfgstore.CLIConfigFunc[cliConfig](cfgstore.CLIConfigArgs{
		DirTypes:     dirTypes,
		DirsProvider: f.dirsProvider,
	})
}

// store returns dirType's store for the fixture's slug and file.
func (f *cliFixture) store(dirType cfgstore.DirType) cfgstore.ConfigStore {
	return cfgstore.NewConfigStore(dirType, cfgstore.ConfigStoreArgs{
		ConfigSlug:   f.slug,
		RelFilepath:  f.file,
		DirsProvider: f.dirsProvider,
	})
}

func (f *cliFixture) write(t *testing.T, dirType cfgstore.DirType, cfg *cliConfig) {
	t.Helper()
	require.NoError(t, f.store(dirType).SaveJSON(cfg), "writing %s config", dirType)
}

// load runs the ConfigFunc as go-cliutil would and returns the typed config.
func (f *cliFixture) load(t *testing.T, dirTypes ...cfgstore.DirType) (*cliConfig, error) {
	t.Helper()
	cfg, err := f.configFunc(dirTypes...)(cliutil.ConfigArgs{AppInfo: f.appInfo()})
	if err != nil {
		return nil, err
	}
	return cliutil.ConfigAs[*cliConfig](cfg), nil
}

func TestCLIConfigFunc_LoadsCLILayer(t *testing.T) {
	f := newCLIFixture(t, TestConfigSlug, "config.json")
	f.write(t, cfgstore.CLIConfigDirType, &cliConfig{RegionValue: "eu", FormatValue: "yaml"})

	got, err := f.load(t, cfgstore.CLIConfigDirType)

	require.NoError(t, err)
	assert.Equal(t, &cliConfig{RegionValue: "eu", FormatValue: "yaml"}, got)
}

func TestCLIConfigFunc_ProjectLayerWinsByDefault(t *testing.T) {
	f := newCLIFixture(t, TestConfigSlug, "config.json")
	f.write(t, cfgstore.CLIConfigDirType, &cliConfig{RegionValue: "eu", FormatValue: "yaml"})
	f.write(t, cfgstore.ProjectConfigDirType, &cliConfig{RegionValue: "ap"})

	got, err := f.load(t) // No DirTypes: CLI then project.

	require.NoError(t, err)
	assert.Equal(t, &cliConfig{RegionValue: "ap", FormatValue: "yaml"}, got)
}

func TestCLIConfigFunc_TakesSlugAndFileFromAppInfo(t *testing.T) {
	f := newCLIFixture(t, "roadrunner", "settings/cli.json")
	f.write(t, cfgstore.CLIConfigDirType, &cliConfig{RegionValue: "us"})

	got, err := f.load(t, cfgstore.CLIConfigDirType)

	require.NoError(t, err)
	assert.Equal(t, &cliConfig{RegionValue: "us"}, got)
	fp, err := f.store(cfgstore.CLIConfigDirType).GetFilepath()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(filepath.ToSlash(string(fp)), "/roadrunner/settings/cli.json"), "config file %s", fp)
}

func TestCLIConfigFunc_ReturnsLoadError(t *testing.T) {
	f := newCLIFixture(t, TestConfigSlug, "config.json")
	fp, err := f.store(cfgstore.CLIConfigDirType).GetFilepath()
	require.NoError(t, err)
	require.NoError(t, fp.Dir().MkdirAll(0o755))
	require.NoError(t, os.WriteFile(string(fp), []byte("{not json"), 0o644))

	got, err := f.load(t, cfgstore.CLIConfigDirType)

	assert.Nil(t, got)
	assert.ErrorIs(t, err, cfgstore.ErrFailedToEnsureConfig)
	assert.ErrorIs(t, err, cfgstore.ErrFailedToUnmarshalConfigFile)
}

// cliOptions applies config to its own global flag through ApplyConfig.
type cliOptions struct {
	*cliutil.GlobalOptions
	region string
}

func newCLIOptions() cliutil.Options {
	return &cliOptions{GlobalOptions: cliutil.NewGlobalOptions(cliutil.GlobalOptionsArgs{})}
}

func (o *cliOptions) FlagDefs() []cliutil.FlagDef {
	return []cliutil.FlagDef{
		{Name: "region", Usage: "Region to use", DefaultString: "us", String: &o.region},
	}
}

func (o *cliOptions) ApplyConfig(cfg cliutil.Config) error {
	o.region = cliutil.ConfigAs[*cliConfig](cfg).Region()
	return nil
}

// showCmd prints the values config and flags produced.
type showCmd struct {
	*cliutil.CmdBase
	format string
}

func newShowCmd() cliutil.Command {
	c := &showCmd{}
	c.CmdBase = cliutil.NewCmdBase(cliutil.CmdArgs{
		Name:        "show",
		Description: "Show the settings",
		FlagSets: []*cliutil.FlagSet{{
			Name: "show",
			FlagDefs: []cliutil.FlagDef{
				{Name: "format", Usage: "Output format", DefaultString: "json", String: &c.format},
			},
		}},
	})
	return c
}

func (c *showCmd) ApplyConfig(cfg cliutil.Config) error {
	c.format = cliutil.ConfigAs[*cliConfig](cfg).Format()
	return nil
}

func (c *showCmd) Handle(args cliutil.HandleArgs) error {
	args.Writer.Printf("region=%s format=%s\n", args.OptionsAs[*cliOptions]().region, c.format)
	return nil
}

func TestCLIConfigFunc_ThroughCLIUtil(t *testing.T) {
	registry := cliutil.NewRegistry()
	registry.Register(newShowCmd)

	tests := []struct {
		name    string
		line    string
		want    string
		wantErr error
		corrupt bool
	}{
		{name: "config applied", line: "show", want: "region=eu format=yaml"},
		{name: "typed flags beat config", line: "--region ap show --format csv", want: "region=ap format=csv"},
		{name: "load error", line: "show", corrupt: true, wantErr: cliutil.ErrConfigLoadFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newCLIFixture(t, TestConfigSlug, "config.json")
			f.write(t, cfgstore.CLIConfigDirType, &cliConfig{RegionValue: "eu", FormatValue: "yaml"})
			if tt.corrupt {
				fp, err := f.store(cfgstore.CLIConfigDirType).GetFilepath()
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(string(fp), []byte("{not json"), 0o644))
			}

			result := cliutiltest.Run(t, cliutiltest.Args{
				Registry:    registry,
				CommandLine: strings.Fields(tt.line),
				Config:      f.configFunc(),
				Options:     newCLIOptions,
				AppInfo:     f.appInfo(),
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, result.Err, tt.wantErr)
				assert.ErrorIs(t, result.Err, cfgstore.ErrFailedToEnsureConfig)
				return
			}
			require.NoError(t, result.Err, "stderr: %s", result.Stderr)
			assert.Contains(t, result.Stdout, tt.want)
		})
	}
}
