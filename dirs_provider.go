package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

type DirFunc func() (dt.DirPath, error)

type DirsProvider struct {
	UserHomeDirFunc   DirFunc
	GetwdFunc         DirFunc
	ProjectDirFunc    DirFunc
	UserConfigDirFunc DirFunc
	UserCacheDirFunc  DirFunc
}
