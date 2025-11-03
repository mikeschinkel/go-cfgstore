package cfgstore

type DirType int

func (dt DirType) String() string {
	switch dt {
	case AppConfigDir:
		return "App config dir"
	case CLIConfigDir:
		return "CLI config dir"
	case ProjectConfigDir:
		return "Project config dir"
	case UnspecifiedConfigDir:
		return "Unspecified config dir"
	default:
	}
	return "Invalid config type"
}

func (dt DirType) Slug() string {
	switch dt {
	case AppConfigDir:
		return "app"
	case CLIConfigDir:
		return "cli"
	case ProjectConfigDir:
		return "project"
	case UnspecifiedConfigDir:
		return "unspecified"
	default:
	}
	return "invalid"
}

const (
	UnspecifiedConfigDir DirType = iota
	AppConfigDir                 // The value os.UserConfigDir() returns
	CLIConfigDir                 // ~/.config/xmlui
	ProjectConfigDir             // <projectDir>/.xmlui
)
