// Command comprehensive demonstrates every go-cfgstore feature. Each step's
// comment names what it shows, and main_test.go checks each step's output.
//
// Every step works under a temporary directory except the logger, which
// CreateWriterLogger always places under ~/.config/<slug>/logs.
//
//	go run .
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cfgstore/cstest"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
)

const (
	configSlug dt.PathSegment = "cfgstore-comprehensive"
	configFile dt.RelFilepath = "config.json"
)

func main() {
	root, err := dt.MkdirTemp(dt.TempDir(), "cfgstore-comprehensive-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err = run(os.Stdout, root)
	cfgstore.LogOnError(root.RemoveAll())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Config is the app's root config. Normalize and Merge make it a
// cfgstore.RootConfig; the json:"-" fields record what loading did.
type Config struct {
	Theme  string   `json:"theme,omitempty"`
	Editor string   `json:"editor,omitempty"`
	Layers []string `json:"-"`
}

// RootConfig marks Config as a cfgstore.RootConfig.
func (c *Config) RootConfig() {}

// Normalize runs once per layer after the file loads, or before a missing file
// is created. It cleans values and records which layer it came from. It sets no
// defaults: a default in a later layer would hide an earlier layer's value.
func (c *Config) Normalize(args cfgstore.NormalizeArgs) error {
	c.Editor = strings.ToLower(c.Editor)
	c.Layers = []string{args.DirType.Slug()}
	return nil
}

// Merge combines earlier, the merge of every earlier layer, into c, a later
// layer. c's values win; earlier fills the fields c leaves empty.
func (c *Config) Merge(earlier cfgstore.RootConfig) cfgstore.RootConfig {
	prev := earlier.(*Config)
	merged := *c
	if merged.Theme == "" {
		merged.Theme = prev.Theme
	}
	if merged.Editor == "" {
		merged.Editor = prev.Editor
	}
	merged.Layers = append(append([]string{}, prev.Layers...), c.Layers...)
	return &merged
}

// demo holds what every step shares.
type demo struct {
	w    io.Writer
	root dt.DirPath
	dp   *cfgstore.DirsProvider
}

func run(w io.Writer, root dt.DirPath) (err error) {
	d := &demo{w: w, root: root, dp: newDirsProvider(root)}
	steps := []struct {
		title string
		fn    func() error
	}{
		{"CreateWriterLogger", d.writerLogger},
		{"Config directories from a custom DirsProvider", d.configDirs},
		{"Save, SaveJSON and EnsureDirs on config stores", d.saveStores},
		{"Layered loading with Normalize and Merge", d.layeredLoad},
		{"Loading one layer", d.singleLayers},
		{"ProjectConfigStoreExists", d.projectExists},
		{"InitProjectConfig", d.initProject},
		{"Cache directories", d.cacheDirs},
		{"cstest helpers", d.cstestHelpers},
	}
	for _, step := range steps {
		d.printf("\n== %s ==\n", step.title)
		err = step.fn()
		if err != nil {
			err = fmt.Errorf("%s: %w", step.title, err)
			goto end
		}
	}
end:
	return err
}

// newDirsProvider returns a custom DirsProvider that places every directory
// under root, so the demo never touches real config.
func newDirsProvider(root dt.DirPath) *cfgstore.DirsProvider {
	at := func(name string) cfgstore.DirFunc {
		return func() (dt.DirPath, error) { return root.Join(name), nil }
	}
	dp := &cfgstore.DirsProvider{
		UserHomeDirFunc:   at("home"),
		GetwdFunc:         at("project"),
		ProjectDirFunc:    at("project"),
		UserConfigDirFunc: at("appconfig"),
		UserCacheDirFunc:  at("cache"),
	}
	// CLIConfigDirType is ~/.config, or XDG_CONFIG_HOME on Linux, under dp's home.
	dp.CLIConfigDirFunc = dp.CLIConfigDirType
	return dp
}

// printf writes to the demo's output; a failed write has nowhere to be reported.
func (d *demo) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(d.w, format, args...)
}

// show prints label and p relative to the demo root.
func (d *demo) show(label string, p string) {
	rel, err := dt.DirPath(p).Rel(d.root)
	if err != nil {
		rel = dt.PathSegments(p)
	}
	d.printf("%-14s <root>/%s\n", label+":", rel)
}

// writerLogger creates the go-cliutil Writer and JSON file logger an app shares,
// and registers the logger with cfgstore.SetLogger.
func (d *demo) writerLogger() (err error) {
	var wl cliutil.WriterLogger
	wl, err = cfgstore.CreateWriterLogger(&cfgstore.WriterLoggerArgs{
		IO:          dt.NewStdIO(d.w, d.w),
		Verbosity:   cliutil.LowVerbosity,
		ConfigSlug:  configSlug,
		LogFile:     "comprehensive.log",
		TruncateLog: true,
	})
	if err != nil {
		goto end
	}
	wl.Printf("writer ready; logging to ~/.config/%s/logs/comprehensive.log\n", configSlug)
	wl.Info("comprehensive example started")
end:
	return err
}

// configDirs shows where each DirType keeps its config.
func (d *demo) configDirs() (err error) {
	var dir dt.DirPath
	dir, err = cfgstore.CLIConfigDir(configSlug, d.dp)
	if err != nil {
		goto end
	}
	d.show("CLI", string(dir))
	dir, err = cfgstore.ProjectConfigDir(configSlug, d.dp)
	if err != nil {
		goto end
	}
	d.show("Project", string(dir))
	dir, err = cfgstore.AppConfigDir(configSlug, d.dp)
	if err != nil {
		goto end
	}
	d.show("App", string(dir))
end:
	return err
}

func (d *demo) stores() *cfgstore.ConfigStores {
	return cfgstore.NewConfigStores(cfgstore.ConfigStoresArgs{
		DirTypes: []cfgstore.DirType{cfgstore.CLIConfigDirType, cfgstore.ProjectConfigDirType},
		ConfigStoreArgs: cfgstore.ConfigStoreArgs{
			ConfigSlug:   configSlug,
			RelFilepath:  configFile,
			DirsProvider: d.dp,
		},
	})
}

// saveStores writes the CLI layer with SaveJSON and the project layer with Save,
// and creates subdirectories beside the CLI config.
func (d *demo) saveStores() (err error) {
	var fp dt.Filepath
	stores := d.stores()
	cli := stores.CLIConfigStore()
	project := stores.ProjectConfigStore()

	err = cli.SaveJSON(&Config{Theme: "dark", Editor: "vim"})
	if err != nil {
		goto end
	}
	err = project.Save([]byte(`{"editor": "Helix"}`))
	if err != nil {
		goto end
	}
	err = cli.EnsureDirs([]dt.PathSegment{"logs", "templates"})
	if err != nil {
		goto end
	}
	for _, cs := range []cfgstore.ConfigStore{cli, project} {
		fp, err = cs.GetFilepath()
		if err != nil {
			goto end
		}
		d.show(cs.DirType().Slug()+" file", string(fp))
	}
end:
	return err
}

// layeredLoad loads CLI then project config; the project layer wins.
func (d *demo) layeredLoad() (err error) {
	var cfg *Config
	cfg, err = cfgstore.LoadConfig[Config](cfgstore.LoadConfigArgs{
		ConfigSlug:   configSlug,
		ConfigFile:   configFile,
		DirsProvider: d.dp,
	})
	if err != nil {
		goto end
	}
	d.printf("merged: theme=%s editor=%s layers=%v\n", cfg.Theme, cfg.Editor, cfg.Layers)
end:
	return err
}

// singleLayers loads one layer with each convenience function.
func (d *demo) singleLayers() (err error) {
	var cfg *Config
	args := cfgstore.LoadConfigArgs{ConfigSlug: configSlug, ConfigFile: configFile, DirsProvider: d.dp}

	cfg, err = cfgstore.LoadCLIConfig[Config](args)
	if err != nil {
		goto end
	}
	d.printf("cli only:     theme=%s editor=%s\n", cfg.Theme, cfg.Editor)

	cfg, err = cfgstore.LoadProjectConfig[Config](args)
	if err != nil {
		goto end
	}
	// Normalize lowercased the project file's "Helix".
	d.printf("project only: theme=%s editor=%s\n", cfg.Theme, cfg.Editor)
end:
	return err
}

// projectExists checks for a project config file without loading it. It takes
// the project's config directory, <project>/.<slug>, not the project directory.
func (d *demo) projectExists() (err error) {
	var dir dt.DirPath
	dir, err = cfgstore.ProjectConfigDir(configSlug, d.dp)
	if err != nil {
		goto end
	}
	d.printf("exists: %v\n", cfgstore.ProjectConfigStoreExists(dir, configSlug, configFile))
	d.printf("exists elsewhere: %v\n", cfgstore.ProjectConfigStoreExists(d.root.Join("elsewhere"), configSlug, configFile))
end:
	return err
}

// initProject creates a project config in the working directory, and refuses to
// overwrite it on a second call.
func (d *demo) initProject() (err error) {
	var cfg *Config
	var wd dt.DirPath
	dir := d.root.Join("newproject")

	wd, err = dt.Getwd()
	if err != nil {
		goto end
	}
	err = dir.MkdirAll(0o755)
	if err != nil {
		goto end
	}
	err = os.Chdir(string(dir))
	if err != nil {
		goto end
	}
	defer func() { err = errors.Join(err, os.Chdir(string(wd))) }()

	cfg, err = cfgstore.InitProjectConfig[Config](configSlug, configFile, nil)
	if err != nil {
		goto end
	}
	d.show("created", string(dir.Join("."+string(configSlug), string(configFile))))
	d.printf("layers: %v\n", cfg.Layers)

	_, err = cfgstore.InitProjectConfig[Config](configSlug, configFile, nil)
	if !errors.Is(err, cfgstore.ErrConfigAlreadyExists) {
		err = fmt.Errorf("second InitProjectConfig: want ErrConfigAlreadyExists, got %w", err)
		goto end
	}
	d.printf("second call: %v\n", err)
	err = nil
end:
	return err
}

// cacheDirs shows the shared and per-app cache directories.
func (d *demo) cacheDirs() (err error) {
	var dir dt.DirPath
	opts := cfgstore.CacheOptions{DirsProvider: d.dp}

	dir, err = cfgstore.GetSharedCacheDir(configSlug, opts)
	if err != nil {
		goto end
	}
	d.show("shared cache", string(dir))
	dir, err = cfgstore.GetAppCacheDir(configSlug, "cli", opts)
	if err != nil {
		goto end
	}
	d.show("app cache", string(dir))
end:
	return err
}

// cstestHelpers uses the cstest DirsProvider, which tests use to fake a user's
// home and project directories under a test root.
func (d *demo) cstestHelpers() (err error) {
	var cfg *Config
	var rel dt.PathSegments
	args := &cstest.TestDirsProviderArgs{
		Username:   "wile",
		ProjectDir: "/home/wile/acme",
		ConfigSlug: configSlug,
		TestRoot:   d.root.Join("cstest"),
	}
	dp := cstest.NewTestDirsProvider(args)

	cfg, err = cfgstore.LoadCLIConfig[Config](cfgstore.LoadConfigArgs{
		ConfigSlug:   configSlug,
		ConfigFile:   configFile,
		DirsProvider: dp,
	})
	if err != nil {
		goto end
	}
	d.printf("fresh CLI config created: layers=%v\n", cfg.Layers)

	rel, err = cstest.GetRelConfigDir(cfgstore.NewConfigStore(cfgstore.CLIConfigDirType, cfgstore.ConfigStoreArgs{
		ConfigSlug:   configSlug,
		RelFilepath:  configFile,
		DirsProvider: dp,
	}), args)
	if err != nil {
		goto end
	}
	d.printf("config dir without test root: %s\n", rel)
	d.printf("RelConfigDir: %s\n", args.RelConfigDir())
end:
	return err
}
