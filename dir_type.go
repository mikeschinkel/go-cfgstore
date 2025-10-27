package cfgstore

type DirType int

const (
	UnspecifiedConfigDir DirType = iota
	GoUserConfigDir              // The value os.UserConfigDir() returns
	DotConfigDir                 // ~/.config/xmlui
	LocalConfigDir               // ./.xmlui
)
