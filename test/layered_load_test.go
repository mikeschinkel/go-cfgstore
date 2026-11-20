package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// layeredConfig defaults Theme and Count in ApplyDefaults and records each
// Normalize call it receives.
type layeredConfig struct {
	Theme   string                              `json:"theme,omitempty"`
	Editor  string                              `json:"editor,omitempty"`
	Verbose bool                                `json:"verbose,omitempty"`
	Count   int                                 `json:"count,omitempty"`
	Tags    []string                            `json:"tags,omitempty"`
	Env     map[string]string                   `json:"env,omitempty"`
	Plugins cfgstore.AppendSlice[string]        `json:"plugins,omitempty"`
	Aliases cfgstore.ReplaceMap[string, string] `json:"aliases,omitempty"`
	Calls   []layerCall                         `json:"-"`
}

// layerCall records the layer a Normalize or Validate call named.
type layerCall struct {
	DirType    cfgstore.DirType
	SourceFile dt.Filepath
}

const (
	defaultTheme = "light"
	defaultCount = 3
)

func (c *layeredConfig) RootConfig() {}
func (c *layeredConfig) ApplyDefaults(cfgstore.DefaultsArgs) error {
	c.Theme = defaultTheme
	c.Count = defaultCount
	return nil
}
func (c *layeredConfig) Normalize(args cfgstore.NormalizeArgs) error {
	c.Calls = append(c.Calls, layerCall{DirType: args.DirType, SourceFile: args.SourceFile})
	return nil
}
func (c *layeredConfig) Validate(cfgstore.ValidateArgs) error { return nil }
func (c *layeredConfig) Merge(args cfgstore.MergeArgs[*layeredConfig]) error {
	return args.Decode(c)
}

// loadLayered writes cli and project as raw JSON (skipping an empty string) and
// loads both layers.
func loadLayered(t *testing.T, cli, project string) (*layeredConfig, *mergeFixture) {
	t.Helper()
	f := newMergeFixture(t)
	if cli != "" {
		f.writeRaw(t, cfgstore.CLIConfigDirType, cli)
	}
	if project != "" {
		f.writeRaw(t, cfgstore.ProjectConfigDirType, project)
	}
	got, err := cfgstore.LoadDefaultConfig[layeredConfig](f.loadArgs())
	require.NoError(t, err)
	return got, f
}

// P27: E-2661's reproduction. The CLI file sets theme dark, the project file
// sets only editor, and the default must not replace the CLI layer's theme.
func TestLoad_DefaultDoesNotHideEarlierLayer(t *testing.T) {
	got, _ := loadLayered(t, `{"theme":"dark"}`, `{"editor":"vim"}`)

	assert.Equal(t, "dark", got.Theme)
	assert.Equal(t, "vim", got.Editor)
}

// P28
func TestLoad_LaterLayerCanSetZeroValues(t *testing.T) {
	got, _ := loadLayered(t,
		`{"theme":"dark","verbose":true,"count":5}`,
		`{"theme":"","verbose":false,"count":0}`,
	)

	assert.Equal(t, "", got.Theme)
	assert.False(t, got.Verbose)
	assert.Equal(t, 0, got.Count)
}

// P29
func TestLoad_KeyAbsentFromEveryLayerHoldsDefault(t *testing.T) {
	got, _ := loadLayered(t, `{"editor":"vim"}`, `{"verbose":true}`)

	assert.Equal(t, defaultTheme, got.Theme)
	assert.Equal(t, defaultCount, got.Count)
	assert.Equal(t, "vim", got.Editor)
	assert.True(t, got.Verbose)
}

// P32, P34
func TestLoad_MapsMergeAndSlicesReplace(t *testing.T) {
	got, _ := loadLayered(t,
		`{"tags":["a","b"],"env":{"A":"1","B":"2"}}`,
		`{"tags":["c"],"env":{"B":"two","C":"3"}}`,
	)

	assert.Equal(t, []string{"c"}, got.Tags)
	assert.Equal(t, map[string]string{"A": "1", "B": "two", "C": "3"}, got.Env)
}

// P35
func TestLoad_AppendSliceAndReplaceMap(t *testing.T) {
	got, f := loadLayered(t,
		`{"plugins":["a","b"],"aliases":{"x":"1","y":"2"}}`,
		`{"plugins":["c"],"aliases":{"z":"3"}}`,
	)

	assert.Equal(t, cfgstore.AppendSlice[string]{"a", "b", "c"}, got.Plugins)
	assert.Equal(t, cfgstore.ReplaceMap[string, string]{"z": "3"}, got.Aliases)

	t.Run("marshal as plain JSON", func(t *testing.T) {
		data, err := json.Marshal(got)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"plugins":["a","b","c"]`)
		assert.Contains(t, string(data), `"aliases":{"z":"3"}`)
	})
	t.Run("null empties an AppendSlice", func(t *testing.T) {
		f.writeRaw(t, cfgstore.ProjectConfigDirType, `{"plugins":null}`)
		got, err := cfgstore.LoadDefaultConfig[layeredConfig](f.loadArgs())
		require.NoError(t, err)
		assert.Empty(t, got.Plugins)
	})
}

// P36
func TestLoad_NormalizeRunsOnceAfterEachLayer(t *testing.T) {
	got, f := loadLayered(t, `{"theme":"dark"}`, `{"editor":"vim"}`)

	assert.Equal(t, []layerCall{
		{cfgstore.CLIConfigDirType, f.filepath(t, cfgstore.CLIConfigDirType)},
		{cfgstore.ProjectConfigDirType, f.filepath(t, cfgstore.ProjectConfigDirType)},
	}, got.Calls)
}

// P45
func TestLoad_ProjectLayerAbsentOrEmpty(t *testing.T) {
	t.Run("missing project file is skipped", func(t *testing.T) {
		got, f := loadLayered(t, `{"theme":"dark"}`, "")

		assert.Equal(t, "dark", got.Theme)
		assert.Equal(t, []layerCall{
			{cfgstore.CLIConfigDirType, f.filepath(t, cfgstore.CLIConfigDirType)},
		}, got.Calls)
	})
	t.Run("empty project file changes nothing", func(t *testing.T) {
		cliOnly, _ := loadLayered(t, `{"theme":"dark","tags":["a"]}`, "")
		got, _ := loadLayered(t, `{"theme":"dark","tags":["a"]}`, `{}`)

		cliOnly.Calls, got.Calls = nil, nil
		assert.Equal(t, cliOnly, got)
	})
}

// P43
func TestLoad_CreatedCLIFileHoldsDefaults(t *testing.T) {
	f := newMergeFixture(t)
	got, err := cfgstore.LoadCLIConfig[layeredConfig](f.loadArgs())
	require.NoError(t, err)
	assert.Equal(t, defaultTheme, got.Theme)

	data, err := os.ReadFile(string(f.filepath(t, cfgstore.CLIConfigDirType)))
	require.NoError(t, err)
	var saved map[string]any
	require.NoError(t, json.Unmarshal(data, &saved))
	assert.Equal(t, defaultTheme, saved["theme"])
	assert.EqualValues(t, defaultCount, saved["count"])
}

// P44
func TestInitProjectConfig_WritesNoDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	got, err := cfgstore.InitProjectConfig[layeredConfig](TestConfigSlug, mergeConfigFile, nil)
	require.NoError(t, err)
	assert.Equal(t, defaultTheme, got.Theme)
	assert.Equal(t, defaultCount, got.Count)

	fp, err := cfgstore.ProjectConfigFilepath(TestConfigSlug, mergeConfigFile)
	require.NoError(t, err)
	data, err := os.ReadFile(string(fp))
	require.NoError(t, err)
	var saved map[string]any
	require.NoError(t, json.Unmarshal(data, &saved))
	assert.Empty(t, saved, "the project file must hold no defaults")
}

// maxLevelConfig keeps the highest Level any layer sets, a rule Decode cannot
// express, and records each layer Merge read through Layer.
type maxLevelConfig struct {
	Level int    `json:"level,omitempty"`
	Name  string `json:"name,omitempty"`
	seen  []maxLevelConfig
}

func (c *maxLevelConfig) RootConfig()                               {}
func (c *maxLevelConfig) ApplyDefaults(cfgstore.DefaultsArgs) error { return nil }
func (c *maxLevelConfig) Normalize(cfgstore.NormalizeArgs) error    { return nil }
func (c *maxLevelConfig) Validate(cfgstore.ValidateArgs) error      { return nil }
func (c *maxLevelConfig) Merge(args cfgstore.MergeArgs[*maxLevelConfig]) (err error) {
	var layer *maxLevelConfig

	prev := c.Level
	layer, err = args.Layer()
	if err != nil {
		goto end
	}
	c.seen = append(c.seen, maxLevelConfig{Level: layer.Level, Name: layer.Name})
	err = args.Decode(c)
	if err != nil {
		goto end
	}
	c.Level = max(prev, c.Level)
end:
	return err
}

// P33
func TestLoad_MergeCanUseLayer(t *testing.T) {
	f := newMergeFixture(t)
	f.writeRaw(t, cfgstore.CLIConfigDirType, `{"level":5,"name":"cli"}`)
	f.writeRaw(t, cfgstore.ProjectConfigDirType, `{"level":2}`)

	got, err := cfgstore.LoadDefaultConfig[maxLevelConfig](f.loadArgs())
	require.NoError(t, err)

	assert.Equal(t, 5, got.Level)
	assert.Equal(t, "cli", got.Name)
	assert.Equal(t, []maxLevelConfig{{Level: 5, Name: "cli"}, {Level: 2}}, got.seen,
		"Layer returns each file decoded alone")
}

// color is a FieldValidator with a value receiver.
type color string

func (c color) Validate() error {
	if c == "" || slices.Contains([]color{"red", "green", "blue"}, c) {
		return nil
	}
	return fmt.Errorf("unknown color %q", string(c))
}

// port is a FieldValidator with a pointer receiver.
type port int

func (p *port) Validate() error {
	if *p > 0 && *p < 65536 {
		return nil
	}
	return fmt.Errorf("port %d out of range", int(*p))
}

type colorItem struct {
	Color color `json:"color"`
}

// Embedded is inlined into its parent's JSON object.
type Embedded struct {
	Shade color `json:"shade"`
}

// validatingConfig holds FieldValidator fields at every depth the walk covers,
// and its Validate checks that Min does not exceed Max.
type validatingConfig struct {
	Embedded
	Color   color                `json:"color,omitempty"`
	UI      colorItem            `json:"ui"`
	Items   []colorItem          `json:"items,omitempty"`
	Ptr     *colorItem           `json:"ptr,omitempty"`
	Ports   map[string]port      `json:"ports,omitempty"`
	Nested  map[string]colorItem `json:"nested,omitempty"`
	Min     int                  `json:"min,omitempty"`
	Max     int                  `json:"max,omitempty"`
	Ignored color                `json:"-"`
}

// validateCalls and mergeCalls record calls the tests below assert on.
var (
	validateCalls []layerCall
	validatedMins []int
	mergeCalls    int
)

func resetValidatingRecords() {
	validateCalls, validatedMins, mergeCalls = nil, nil, 0
}

func (c *validatingConfig) RootConfig()                               {}
func (c *validatingConfig) ApplyDefaults(cfgstore.DefaultsArgs) error { return nil }
func (c *validatingConfig) Normalize(cfgstore.NormalizeArgs) error    { return nil }
func (c *validatingConfig) Validate(args cfgstore.ValidateArgs) error {
	validateCalls = append(validateCalls, layerCall{args.DirType, args.SourceFile})
	validatedMins = append(validatedMins, c.Min)
	if c.Max != 0 && c.Min > c.Max {
		return fmt.Errorf("min %d exceeds max %d", c.Min, c.Max)
	}
	return nil
}
func (c *validatingConfig) Merge(args cfgstore.MergeArgs[*validatingConfig]) error {
	mergeCalls++
	return args.Decode(c)
}

// P38
func TestLoad_ValidatesFieldsAtEveryDepth(t *testing.T) {
	resetValidatingRecords()
	f := newMergeFixture(t)
	f.writeRaw(t, cfgstore.CLIConfigDirType, `{
		"shade": "mauve",
		"color": "pink",
		"ui": {"color": "teal"},
		"items": [{"color": "red"}, {"color": "plaid"}],
		"ptr": {"color": "beige"},
		"ports": {"http": 80, "bad": 70000},
		"nested": {"k": {"color": "gold"}}
	}`)

	_, err := cfgstore.LoadCLIConfig[validatingConfig](f.loadArgs())
	require.Error(t, err)

	assert.ElementsMatch(t,
		[]string{"shade", "color", "ui.color", "items.1.color", "ptr.color", "ports.bad", "nested.k.color"},
		jsonPaths(err))
	assert.Equal(t, 0, mergeCalls, "the layered pass must not run")
}

// P37, P39
func TestLoad_ReportsEveryInvalidFieldFromEveryFile(t *testing.T) {
	resetValidatingRecords()
	f := newMergeFixture(t)
	cliFile := f.filepath(t, cfgstore.CLIConfigDirType)
	projectFile := f.filepath(t, cfgstore.ProjectConfigDirType)
	f.writeRaw(t, cfgstore.CLIConfigDirType, `{"color":"pink","ui":{"color":"teal"},"min":1}`)
	f.writeRaw(t, cfgstore.ProjectConfigDirType, `{"items":[{"color":"plaid"}],"min":9,"max":2}`)

	_, err := cfgstore.LoadDefaultConfig[validatingConfig](f.loadArgs())
	require.Error(t, err)

	byFile := map[dt.Filepath][]string{}
	var crossField []error
	for _, e := range leafErrs(err) {
		fp, _ := cfgstore.ErrValue[dt.Filepath](e, "config_file")
		if p, ok := cfgstore.ErrValue[string](e, "json_path"); ok {
			byFile[fp] = append(byFile[fp], p)
			assert.ErrorIs(t, e, cfgstore.ErrInvalidConfigValue)
			continue
		}
		assert.ErrorIs(t, e, cfgstore.ErrInvalidConfig)
		assert.Equal(t, projectFile, fp)
		crossField = append(crossField, e)
	}
	assert.ElementsMatch(t, []string{"color", "ui.color"}, byFile[cliFile])
	assert.Equal(t, []string{"items.0.color"}, byFile[projectFile])
	assert.Len(t, crossField, 1)
	assert.Contains(t, err.Error(), string(cliFile))
	assert.Contains(t, err.Error(), string(projectFile))
	assert.Contains(t, err.Error(), "min 9 exceeds max 2")

	assert.Equal(t, []layerCall{
		{cfgstore.CLIConfigDirType, cliFile},
		{cfgstore.ProjectConfigDirType, projectFile},
	}, validateCalls)
	assert.Equal(t, []int{1, 9}, validatedMins, "Validate sees each file decoded alone")
	assert.Equal(t, 0, mergeCalls, "the layered pass must not run")
}

// P40
func TestLoad_DecodeFailureReportedWithOtherFilesErrors(t *testing.T) {
	resetValidatingRecords()
	f := newMergeFixture(t)
	f.writeRaw(t, cfgstore.CLIConfigDirType, `{not json`)
	f.writeRaw(t, cfgstore.ProjectConfigDirType, `{"color":"pink"}`)

	_, err := cfgstore.LoadDefaultConfig[validatingConfig](f.loadArgs())
	require.Error(t, err)

	assert.ErrorIs(t, err, cfgstore.ErrFailedToEnsureConfig)
	assert.ErrorIs(t, err, cfgstore.ErrFailedToUnmarshalConfigFile)
	assert.Equal(t, []string{"color"}, jsonPaths(err), "the project file is still checked")
	assert.Equal(t, 0, mergeCalls)
}

// failingConfig fails the step named in failStep for the layer in failLayer.
type failingConfig struct {
	Name string `json:"name,omitempty"`
}

var (
	failStep       string
	failLayer      cfgstore.DirType
	normalizeCalls []cfgstore.DirType
)

var errStepFailed = errors.New("step failed")

func (c *failingConfig) RootConfig() {}
func (c *failingConfig) ApplyDefaults(cfgstore.DefaultsArgs) error {
	if failStep == "defaults" {
		return errStepFailed
	}
	return nil
}
func (c *failingConfig) Normalize(args cfgstore.NormalizeArgs) error {
	normalizeCalls = append(normalizeCalls, args.DirType)
	if failStep == "normalize" && args.DirType == failLayer {
		return errStepFailed
	}
	return nil
}
func (c *failingConfig) Validate(cfgstore.ValidateArgs) error { return nil }
func (c *failingConfig) Merge(args cfgstore.MergeArgs[*failingConfig]) error {
	if failStep == "merge" && args.DirType == failLayer {
		return errStepFailed
	}
	return args.Decode(c)
}

// P41, P42
func TestLoad_LayeredPassStopsAtFirstFailure(t *testing.T) {
	for _, tc := range []struct {
		step  string
		layer cfgstore.DirType
		calls []cfgstore.DirType
	}{
		{"merge", cfgstore.CLIConfigDirType, nil},
		{"normalize", cfgstore.CLIConfigDirType, []cfgstore.DirType{cfgstore.CLIConfigDirType}},
		{"merge", cfgstore.ProjectConfigDirType, []cfgstore.DirType{cfgstore.CLIConfigDirType}},
	} {
		t.Run(tc.step+" "+tc.layer.Slug(), func(t *testing.T) {
			f := newMergeFixture(t)
			f.writeRaw(t, cfgstore.CLIConfigDirType, `{"name":"cli"}`)
			f.writeRaw(t, cfgstore.ProjectConfigDirType, `{"name":"project"}`)
			failStep, failLayer, normalizeCalls = tc.step, tc.layer, nil
			t.Cleanup(func() { failStep = "" })

			got, err := cfgstore.LoadDefaultConfig[failingConfig](f.loadArgs())
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorIs(t, err, errStepFailed)
			assert.ErrorIs(t, err, cfgstore.ErrFailedToLoadConfig)
			assertErrValue(t, err, "config_file", f.filepath(t, tc.layer))
			assertErrValue(t, err, "dir_type", tc.layer)
			assert.Equal(t, tc.calls, normalizeCalls)
		})
	}
	t.Run("defaults", func(t *testing.T) {
		f := newMergeFixture(t)
		f.writeRaw(t, cfgstore.CLIConfigDirType, `{"name":"cli"}`)
		failStep = "defaults"
		t.Cleanup(func() { failStep = "" })

		got, err := cfgstore.LoadDefaultConfig[failingConfig](f.loadArgs())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.ErrorIs(t, err, errStepFailed)
		assert.ErrorIs(t, err, cfgstore.ErrFailedToApplyDefaults)
	})
}

// leafErrs returns every doterr entry inside err that names a config_file.
func leafErrs(err error) (leaves []error) {
	if _, ok := err.(interface{ IsDotErrEntry() }); ok {
		if _, ok := cfgstore.ErrValue[dt.Filepath](err, "config_file"); ok {
			return []error{err}
		}
	}
	if u, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range u.Unwrap() {
			leaves = append(leaves, leafErrs(e)...)
		}
	}
	return leaves
}

// jsonPaths returns the json_path of every invalid-value error inside err.
func jsonPaths(err error) (paths []string) {
	for _, e := range leafErrs(err) {
		if p, ok := cfgstore.ErrValue[string](e, "json_path"); ok {
			paths = append(paths, p)
		}
	}
	return paths
}
