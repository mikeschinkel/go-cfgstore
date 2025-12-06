package cfgstore

// LoadCLIConfig loads configuration from CLI config directory only (~/.config/<slug>).
// This is a convenience function for the common case of loading only user-level CLI configuration.
//
// Example:
//
//	config, err := cfgstore.LoadCLIConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
//	    ConfigSlug: dt.PathSegment("myapp"),
//	    ConfigFile: dt.RelFilepath("config.json"),
//	    Options:    myOptions,  // or nil
//	})
func LoadCLIConfig[RC any, PRC RootConfigPtr[RC]](args LoadConfigArgs) (PRC, error) {
	args.DirTypes = []DirType{CLIConfigDirType}
	return LoadConfig[RC, PRC](args)
}

// LoadProjectConfig loads configuration from project directory only (./<slug>).
// This is a convenience function for the common case of loading only project-specific configuration.
//
// Example:
//
//	config, err := cfgstore.LoadProjectConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
//	    ConfigSlug: dt.PathSegment("myapp"),
//	    ConfigFile: dt.RelFilepath("config.json"),
//	    Options:    myOptions,  // or nil
//	})
func LoadProjectConfig[RC any, PRC RootConfigPtr[RC]](args LoadConfigArgs) (PRC, error) {
	args.DirTypes = []DirType{ProjectConfigDirType}
	return LoadConfig[RC, PRC](args)
}

// LoadDefaultConfig loads configuration with default precedence: CLI + Project.
// Project configuration takes precedence over CLI configuration.
// This is the most common multi-store pattern for applications that support both
// user-level defaults and project-specific overrides.
//
// Example:
//
//	config, err := cfgstore.LoadDefaultConfig[MyConfig, *MyConfig](cfgstore.LoadConfigArgs{
//	    ConfigSlug: dt.PathSegment("myapp"),
//	    ConfigFile: dt.RelFilepath("config.json"),
//	    Options:    myOptions,  // or nil
//	})
func LoadDefaultConfig[RC any, PRC RootConfigPtr[RC]](args LoadConfigArgs) (PRC, error) {
	args.DirTypes = []DirType{CLIConfigDirType, ProjectConfigDirType}
	return LoadConfig[RC, PRC](args)
}
