package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
)

func CLIConfigDir(configSlug dt.PathSegment, dps ...*DirsProvider) (cd dt.DirPath, err error) {
	var dp *DirsProvider
	if dps != nil {
		dp = dps[0]
	}
	cd, err = ConfigDir(CLIConfigDirType, configSlug, dp)
	return cd, err
}
func AppConfigDir(configSlug dt.PathSegment, dps ...*DirsProvider) (cd dt.DirPath, err error) {
	var dp *DirsProvider
	if dps != nil {
		dp = dps[0]
	}
	cd, err = ConfigDir(AppConfigDirType, configSlug, dp)
	return cd, err
}
func ProjectConfigDir(configSlug dt.PathSegment, dps ...*DirsProvider) (cd dt.DirPath, err error) {
	var dp *DirsProvider
	if dps != nil {
		dp = dps[0]
	}
	cd, err = ConfigDir(ProjectConfigDirType, configSlug, dp)
	return cd, err
}

func ConfigDir(dirType DirType, configSlug dt.PathSegment, dp *DirsProvider) (cd dt.DirPath, err error) {
	var dir dt.DirPath
	if dp == nil {
		dp = DefaultDirsProvider()
	}

	switch dirType {
	case CLIConfigDirType:
		dir, err = dp.CLIConfigDirFunc()
		if err != nil {
			goto end
		}
		cd = dt.DirPathJoin(dir, configSlug)

	case ProjectConfigDirType:
		dir, err = dp.ProjectDirFunc()
		if err != nil {
			err = NewErr(ErrFailedGettingWorkingDir, err)
			goto end
		}
		cd = dt.DirPathJoin(dir, "."+configSlug)

	case AppConfigDirType:
		dir, err = dp.UserConfigDirFunc()
		if err != nil {
			err = NewErr(ErrFailedGettingUserConfigDir, err)
			goto end
		}
		cd = dt.DirPathJoin(dir, configSlug)

	case CustomConfigDirType:
		cd, err = dp.CustomConfigDirFunc()
		if err != nil {
			err = NewErr(ErrFailedGettingCustomConfigDir, err)
			goto end
		}

	case UnspecifiedConfigDirType:
		err = NewErr(ErrConfigDirTypeNotSet)
		goto end
	default:
		err = NewErr(
			ErrInvalidConfigDirType,
			"config_dir_type", dirType,
		)
		goto end
	}
end:
	return cd, err
}
