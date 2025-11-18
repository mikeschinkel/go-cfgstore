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
	CLIConfigDirFunc  DirFunc
	UserCacheDirFunc  DirFunc
}

//func (dp DirsProvider) WithProjectDir(dir dt.DirPath) DirsProvider {
//	newDP := dp
//	newDP.ProjectDirFunc = func() (dt.DirPath, error) {
//		return dir, nil
//	}
//	return newDP
//}
//func (dp DirsProvider) WithUserConfigDir(dir dt.DirPath) DirsProvider {
//	newDP := dp
//	newDP.UserConfigDirFunc = func() (dt.DirPath, error) {
//		return dir, nil
//	}
//	return newDP
//}
