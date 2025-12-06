package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

// LoadConfigArgs provides arguments for loading configuration with sensible defaults.
type LoadConfigArgs struct {
	ConfigSlug   dt.PathSegment
	ConfigFile   dt.RelFilepath
	DirTypes     []DirType     // optional: defaults to [CLIConfigDirType, ProjectConfigDirType]
	DirsProvider *DirsProvider // optional: defaults to DefaultDirsProvider()
	Options      Options       // optional: can be nil
}

// LoadConfig loads configuration from one or more config stores with sensible defaults.
// This is the core flexible function that all convenience functions delegate to.
//
// Defaults applied:
// - DirTypes: [CLIConfigDirType, ProjectConfigDirType] if not specified
// - DirsProvider: DefaultDirsProvider() if not specified
// - Options: nil is acceptable (passed through to Normalize)
func LoadConfig[RC any, PRC RootConfigPtr[RC]](args LoadConfigArgs) (prc PRC, err error) {
	// Apply defaults
	if len(args.DirTypes) == 0 {
		args.DirTypes = []DirType{CLIConfigDirType, ProjectConfigDirType}
	}
	if args.DirsProvider == nil {
		args.DirsProvider = DefaultDirsProvider()
	}

	// Create config stores
	configStores := NewConfigStores(ConfigStoresArgs{
		DirTypes: args.DirTypes,
		ConfigStoreArgs: ConfigStoreArgs{
			ConfigSlug:   args.ConfigSlug,
			RelFilepath:  args.ConfigFile,
			DirsProvider: args.DirsProvider,
		},
	})

	// Load config using LoadConfigStores
	return LoadConfigStores[RC, PRC](configStores, RootConfigArgs{
		DirTypes:     args.DirTypes,
		Options:      args.Options,
		DirsProvider: args.DirsProvider,
	})
}
