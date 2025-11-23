package cfgstore

type DirType int

func (dt DirType) String() string {
	switch dt {
	case AppConfigDirType:
		return "App config dir"
	case CLIConfigDirType:
		return "CLI config dir"
	case ProjectConfigDirType:
		return "Project config dir"
	case UnspecifiedConfigDirType:
		return "Unspecified config dir"
	default:
	}
	return "Invalid config type"
}

func (dt DirType) Slug() string {
	switch dt {
	case AppConfigDirType:
		return "app"
	case CLIConfigDirType:
		return "cli"
	case ProjectConfigDirType:
		return "project"
	case UnspecifiedConfigDirType:
		return "unspecified"
	default:
	}
	return "invalid"
}

const (
	UnspecifiedConfigDirType DirType = iota
	AppConfigDirType                 // The value os.UserConfigDir() returns
	CLIConfigDirType                 // ~/.config/xmlui
	ProjectConfigDirType             // <projectDir>/.xmlui
)
