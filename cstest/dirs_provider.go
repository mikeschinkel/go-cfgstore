package cstest

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mikeschinkel/go-cfgstore"
	"github.com/mikeschinkel/go-dt"
)

const (
	WindowsAppConfigRelPathSegments = `AppData\Roaming`
	macOSAppConfigRelPathSegments   = `Library/Application Support`
)

type TestDirsProviderArgs struct {
	Username     dt.PathSegment
	TestRoot     dt.DirPath
	TestRootFunc func() dt.DirPath
	ProjectDir   dt.DirPath
	ConfigSlug   dt.PathSegment
	omitTestRoot bool
}

func (args *TestDirsProviderArgs) RelConfigDir() dt.PathSegments {
	return dt.PathSegments(filepath.Join(
		"Users",
		string(args.Username),
		string(cfgstore.DotConfigPathSegment),
		string(args.ConfigSlug),
	))
}
func (args *TestDirsProviderArgs) OmitTestRoot() bool {
	return args.omitTestRoot
}

func (args *TestDirsProviderArgs) WithoutTestRoot(fn cfgstore.DirFunc) (dp dt.DirPath, err error) {
	var mutex sync.Mutex
	mutex.Lock()
	args.omitTestRoot = true
	dp, err = fn()
	args.omitTestRoot = false
	mutex.Unlock()
	return dp, err
}

func (args *TestDirsProviderArgs) GetTestRoot(dp dt.DirPath) (_ dt.DirPath) {
	if args.OmitTestRoot() {
		goto end
	}
	if args.TestRoot == "" {
		args.TestRoot = args.TestRootFunc()
	}
	dp = dt.DirPathJoin(args.TestRoot, dp)
end:
	return dp
}

func NewTestDirsProvider(args *TestDirsProviderArgs) *cfgstore.DirsProvider {
	return &cfgstore.DirsProvider{
		UserHomeDirFunc: func() (dp dt.DirPath, err error) {
			dp, err = getTestUserHomeDir(args.Username)
			if err != nil {
				goto end
			}
			dp = args.GetTestRoot(dp)
		end:
			return dp, err
		},
		GetwdFunc: func() (wd dt.DirPath, err error) {
			wd, err = dt.Getwd()
			if err != nil {
				goto end
			}
			wd = args.GetTestRoot(wd)
		end:
			return wd, err
		},
		ProjectDirFunc: func() (dp dt.DirPath, err error) {
			dp, err = getTestProjectDir(args)
			if err != nil {
				goto end
			}
			dp = args.GetTestRoot(dp)
		end:
			return dp, err
		},
		UserConfigDirFunc: func() (dp dt.DirPath, err error) {
			dp, err = getTestUserConfigDir(args.Username)
			if err != nil {
				goto end
			}
			dp = args.GetTestRoot(dp)
		end:
			return dp, err
		},
		CLIConfigDirFunc: func() (dp dt.DirPath, err error) {
			dp, err = getTestCLIConfigDir(args.Username)
			if err != nil {
				goto end
			}
			dp = args.GetTestRoot(dp)
		end:
			return dp, err
		},
	}
}

func getTestProjectDir(args *TestDirsProviderArgs) (dir dt.DirPath, err error) {
	var homeDir dt.DirPath

	homeDir, err = getTestUserHomeDir(args.Username)
	if err != nil {
		goto end
	}

	if args.ProjectDir == "" {
		err = dt.NewErr(
			dt.ErrEmpty,
			"dir_type", cfgstore.ProjectConfigDir.Slug(),
		)
		goto end
	}

	switch runtime.GOOS {
	default:
		dir = args.ProjectDir
	case "windows", "darwin", "ios":
		rel, err := args.ProjectDir.Rel(homeDir)
		if err == nil && len(rel) > 0 {
			dir = dt.DirPathJoin(homeDir, rel.UpperFirst())
			goto end
		}
		dir = args.ProjectDir
	}
end:
	return dir, err
}

func getTestUserConfigDir(username dt.PathSegment) (dir dt.DirPath, err error) {
	var homeDir dt.DirPath

	homeDir, err = getTestUserHomeDir(username)
	if err != nil {
		goto end
	}
	switch runtime.GOOS {
	case "windows":
		dir = dt.DirPathJoin(homeDir, WindowsAppConfigRelPathSegments)
	case "darwin", "ios":
		dir = dt.DirPathJoin(homeDir, macOSAppConfigRelPathSegments)
	default: // Unix
		dir = dt.DirPathJoin(homeDir, cfgstore.DotConfigPathSegment)
	}
end:
	if err != nil {
		err = dt.WithErr(err,
			cfgstore.ErrFailedGettingUserConfigDir,
		)
	}
	return dir, err
}
func getTestCLIConfigDir(username dt.PathSegment) (dir dt.DirPath, err error) {
	var homeDir dt.DirPath

	homeDir, err = getTestUserHomeDir(username)
	if err != nil {
		goto end
	}
	dir = dt.DirPathJoin(homeDir, cfgstore.DotConfigPathSegment)
end:
	if err != nil {
		err = dt.WithErr(err,
			cfgstore.ErrFailedGettingCLIConfigDir,
		)
	}
	return dir, err
}
func getTestUserHomeDir(username dt.PathSegment) (dir dt.DirPath, err error) {
	err = validateUsername(username)
	if err != nil {
		goto end
	}
	switch runtime.GOOS {
	case "windows":
		dir = `C:\Users`
	case "darwin":
		dir = "/Users"
	default:
		dir = "/home"
	}
	dir = dt.DirPathJoin(dir, username)
end:
	if err != nil {
		err = dt.WithErr(err,
			cfgstore.ErrFailedGettingUserHomeDir,
		)
	}
	return dir, err
}

func validateUsername(username dt.PathSegment) (err error) {
	if username == "" {
		err = dt.NewErr(cfgstore.ErrUsernameRequired)
		goto end
	}
	if strings.ContainsAny(string(username), `\/`) {
		err = dt.NewErr(cfgstore.ErrInvalidUsername,
			"diagnostic", `username cannot contain slashes ('\' or '/')`,
		)
		goto end
	}
end:
	return err
}
