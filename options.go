package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

type Options interface {
	Options()
}

type AppInfo struct {
	Options
	AppName         string
	AppSlug         dt.Identifier
	AppConfigSubdir dt.PathSegments
	RootConfigFile  dt.RelFilepath
	DirTypes        []DirType
}

type AppInfoArgs struct {
	AppName         string
	AppSlug         dt.Identifier
	AppConfigSubdir dt.PathSegments
	RootConfigFile  dt.RelFilepath
}

func NewAppInfo(args AppInfoArgs) AppInfo {
	return AppInfo{
		AppName:         args.AppName,
		AppSlug:         args.AppSlug,
		AppConfigSubdir: args.AppConfigSubdir,
		RootConfigFile:  args.RootConfigFile,
	}
}
