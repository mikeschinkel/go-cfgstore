package cfgstore

import (
	"github.com/mikeschinkel/go-cliutil"
)

// CLIRootConfigPtr constrains the root config a go-cliutil app loads: *RC must
// be both a go-cfgstore RootConfig and a go-cliutil Config.
type CLIRootConfigPtr[RC any] interface {
	RootConfigPtr[RC]
	cliutil.Config
}

// CLIConfigArgs configures CLIConfigFunc. Both fields are optional and default
// as LoadConfig defaults them.
type CLIConfigArgs struct {
	DirTypes     []DirType     // optional: defaults to [CLIConfigDirType, ProjectConfigDirType]
	DirsProvider *DirsProvider // optional: defaults to DefaultDirsProvider()
}

// CLIConfigFunc returns a cliutil.ConfigFunc that loads the app's root config
// with LoadConfig, taking the config slug and file from the run's AppInfo. The
// loaded *RC is the cliutil.Config, so cliutil.ConfigAs[*RC] returns it typed.
//
// Example:
//
//	runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
//	    AppInfo: appInfo,
//	    Config:  cfgstore.CLIConfigFunc[RootConfig](cfgstore.CLIConfigArgs{}),
//	    Options: NewOptions,
//	})
func CLIConfigFunc[RC any, PRC CLIRootConfigPtr[RC]](args CLIConfigArgs) cliutil.ConfigFunc {
	return func(ca cliutil.ConfigArgs) (cfg cliutil.Config, err error) {
		var prc PRC

		prc, err = LoadConfig[RC, PRC](LoadConfigArgs{
			ConfigSlug:   ca.AppInfo.ConfigSlug(),
			ConfigFile:   ca.AppInfo.ConfigFile(),
			DirTypes:     args.DirTypes,
			DirsProvider: args.DirsProvider,
		})
		if err != nil {
			goto end
		}
		cfg = prc
	end:
		return cfg, err
	}
}
