// Command cliutil is a go-cliutil app whose config go-cfgstore loads. One line
// wires it: Config: cfgstore.CLIConfigFunc[RootConfig](...). Options and the
// show command copy config into typed fields with ApplyConfig, and a flag the
// user types still wins.
//
// It reads ~/.config/cfgstore-cliutil/config.json, creating it when missing,
// and then ./.cfgstore-cliutil/config.json when present, which wins.
//
//	go run . show
//	go run . --region ap show --format csv
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt/appinfo"
)

func init() {
	cliutil.RegisterCommand(NewShowCmd)
	cliutil.RegisterCommand(NewHelpCmd)
}

// RootConfig is the app's config file. Its typed methods are what ApplyConfig
// reads, so the compiler checks every config-to-flag mapping.
type RootConfig struct {
	VerbosityLevel int    `json:"verbosity,omitempty"`
	RegionName     string `json:"region,omitempty"`
	FormatName     string `json:"format,omitempty"`
}

// Config marks RootConfig as a cliutil.Config.
func (*RootConfig) Config() {}

// RootConfig marks RootConfig as a cfgstore.RootConfig.
func (*RootConfig) RootConfig() {}

// Normalize has nothing to clean in this config.
func (*RootConfig) Normalize(cfgstore.NormalizeArgs) error { return nil }

// Merge fills the fields c, a later layer, leaves empty from earlier layers.
func (c *RootConfig) Merge(earlier cfgstore.RootConfig) cfgstore.RootConfig {
	prev := earlier.(*RootConfig)
	merged := *c
	if merged.VerbosityLevel == 0 {
		merged.VerbosityLevel = prev.VerbosityLevel
	}
	if merged.RegionName == "" {
		merged.RegionName = prev.RegionName
	}
	if merged.FormatName == "" {
		merged.FormatName = prev.FormatName
	}
	return &merged
}

// Verbosity returns the configured verbosity, or 0 when unset.
func (c *RootConfig) Verbosity() int { return c.VerbosityLevel }

// Region returns the configured region, or "" when unset.
func (c *RootConfig) Region() string { return c.RegionName }

// Format returns the configured output format, or "" when unset.
func (c *RootConfig) Format() string { return c.FormatName }

// Options holds the app's global flags: go-cliutil's built-ins and --region.
type Options struct {
	*cliutil.GlobalOptions
	Region string
}

// NewOptions builds a fresh Options for each run.
func NewOptions() cliutil.Options {
	return &Options{GlobalOptions: cliutil.NewGlobalOptions(cliutil.GlobalOptionsArgs{})}
}

// FlagDefs declares --region.
func (o *Options) FlagDefs() []cliutil.FlagDef {
	return []cliutil.FlagDef{
		{Name: "region", Usage: "Region to use", DefaultString: "us", String: &o.Region},
	}
}

// ApplyConfig copies config into the global flags. A value the config leaves
// unset keeps the flag's default.
func (o *Options) ApplyConfig(cfg cliutil.Config) error {
	c := cliutil.ConfigAs[*RootConfig](cfg)
	if c.Verbosity() != 0 {
		o.SetVerbosity(c.Verbosity())
	}
	if c.Region() != "" {
		o.Region = c.Region()
	}
	return nil
}

// ShowCmd prints the settings this run resolved and where --format came from.
type ShowCmd struct {
	*cliutil.CmdBase
	format string
}

// NewShowCmd constructs the show command.
func NewShowCmd() cliutil.Command {
	c := &ShowCmd{}
	c.CmdBase = cliutil.NewCmdBase(cliutil.CmdArgs{
		Name:        "show",
		Description: "Show the settings config and flags produced",
		FlagSets: []*cliutil.FlagSet{{
			Name: "show",
			FlagDefs: []cliutil.FlagDef{
				{Name: "format", Usage: "Output format", DefaultString: "json", String: &c.format},
			},
		}},
	})
	return c
}

// ApplyConfig sets the format from config; a typed --format still wins.
func (c *ShowCmd) ApplyConfig(cfg cliutil.Config) error {
	format := cliutil.ConfigAs[*RootConfig](cfg).Format()
	if format != "" {
		c.format = format
	}
	return nil
}

// Handle prints the resolved settings.
func (c *ShowCmd) Handle(args cliutil.HandleArgs) error {
	opts := args.OptionsAs[*Options]()
	format, _ := args.Flags.Get("format")
	args.Writer.Printf("region=%s format=%s format_source=%s verbosity=%d\n",
		opts.Region, c.format, format.Source, opts.Verbosity())
	return nil
}

// HelpCmd shows help for the app or for one command.
type HelpCmd struct {
	*cliutil.CmdBase
	topic []string
}

// NewHelpCmd constructs the help command.
func NewHelpCmd() cliutil.Command {
	c := &HelpCmd{}
	c.CmdBase = cliutil.NewCmdBase(cliutil.CmdArgs{
		Name:        "help",
		Usage:       "help [command]",
		Description: "Show help information",
		Order:       99,
		ArgDefs:     []*cliutil.ArgDef{{Name: "command", Usage: "Command to explain", Strings: &c.topic}},
	})
	return c
}

// Handle shows the help the topic asks for.
func (c *HelpCmd) Handle(args cliutil.HandleArgs) error {
	if len(c.topic) == 0 {
		return cliutil.ShowMainHelp(args.UsageArgs())
	}
	return cliutil.ShowCmdHelp(c.topic, args.UsageArgs())
}

// newAppInfo names the app and its config slug and file, which CLIConfigFunc reads.
func newAppInfo() appinfo.AppInfo {
	return appinfo.New(appinfo.Args{
		Name:        "cliutil",
		Description: "A go-cliutil app configured by go-cfgstore",
		ExeName:     "cliutil",
		InfoURL:     "https://github.com/mikeschinkel/go-cfgstore",
		ConfigSlug:  "cfgstore-cliutil",
		ConfigFile:  "config.json",
	})
}

func main() {
	// go-cfgstore logs through the logger the app sets.
	cfgstore.SetLogger(slog.New(slog.DiscardHandler))

	runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
		AppInfo: newAppInfo(),
		Config:  cfgstore.CLIConfigFunc[RootConfig](cfgstore.CLIConfigArgs{}),
		Options: NewOptions,
	})
	err := runner.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(cliutil.ExitCode(err))
	}
}
