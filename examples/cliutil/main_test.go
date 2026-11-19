package main

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-cliutil/cliutiltest"
	"github.com/mikeschinkel/go-dt"
)

func TestMain(m *testing.M) {
	cfgstore.SetLogger(slog.New(slog.DiscardHandler))
	os.Exit(m.Run())
}

// configFunc returns the app's ConfigFunc with every directory under a temp
// root, so the test never reads real config. cli and project, when not nil,
// become those layers' config files.
func configFunc(t *testing.T, cli, project *RootConfig) cliutil.ConfigFunc {
	t.Helper()
	root := dt.DirPath(t.TempDir())
	at := func(name string) cfgstore.DirFunc {
		return func() (dt.DirPath, error) { return root.Join(name), nil }
	}
	dp := &cfgstore.DirsProvider{
		UserHomeDirFunc:   at("home"),
		UserConfigDirFunc: at("home/.config"),
		GetwdFunc:         at("project"),
		ProjectDirFunc:    at("project"),
	}
	dp.CLIConfigDirFunc = dp.CLIConfigDirType

	info := newAppInfo()
	layers := map[cfgstore.DirType]*RootConfig{
		cfgstore.CLIConfigDirType:     cli,
		cfgstore.ProjectConfigDirType: project,
	}
	for dirType, cfg := range layers {
		if cfg == nil {
			continue
		}
		store := cfgstore.NewConfigStore(dirType, cfgstore.ConfigStoreArgs{
			ConfigSlug:   info.ConfigSlug(),
			RelFilepath:  info.ConfigFile(),
			DirsProvider: dp,
		})
		err := store.SaveJSON(cfg)
		if err != nil {
			t.Fatalf("writing %s config: %v", dirType, err)
		}
	}
	return cfgstore.CLIConfigFunc[RootConfig](cfgstore.CLIConfigArgs{DirsProvider: dp})
}

func TestShow(t *testing.T) {
	t.Parallel()
	cliConfig := &RootConfig{VerbosityLevel: 2, RegionName: "eu", FormatName: "yaml"}
	tests := []struct {
		name    string
		line    string
		cli     *RootConfig
		project *RootConfig
		want    string
	}{
		{name: "no config files", line: "show", want: "region=us format=json format_source=default verbosity=1"},
		{name: "config beats defaults", line: "show", cli: cliConfig, want: "region=eu format=yaml format_source=config verbosity=2"},
		{name: "project layer wins", line: "show", cli: cliConfig, project: &RootConfig{RegionName: "ap"}, want: "region=ap format=yaml"},
		{name: "typed flags beat config", line: "--region sa -v 3 show --format csv", cli: cliConfig, want: "region=sa format=csv format_source=command line verbosity=3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := cliutiltest.Run(t, cliutiltest.Args{
				CommandLine: strings.Fields(tt.line),
				Config:      configFunc(t, tt.cli, tt.project),
				Options:     NewOptions,
				AppInfo:     newAppInfo(),
			})
			if result.Err != nil {
				t.Fatalf("run: %v\nstderr: %s", result.Err, result.Stderr)
			}
			if !strings.Contains(result.Stdout, tt.want) {
				t.Errorf("stdout: want %q in %q", tt.want, result.Stdout)
			}
		})
	}
}
