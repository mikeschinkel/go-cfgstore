package cfgstore

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/de"
	"github.com/mikeschinkel/go-dt/dtx"
)

// DefaultConfigDirType is currently hardcoded for ~/.config but having this
// const will make it easy to track down how where to change it if we want to make it
// configurable.
const DefaultConfigDirType = CLIConfigDir

const DotConfigPathSegment dt.PathSegment = ".config"

// ConfigStore provides file operations for Gmail APIConfig
type ConfigStore interface {
	Load() ([]byte, error)
	Save([]byte) error
	LoadJSON(data any, opts ...jsonv2.Options) error
	SaveJSON(data any) error
	Exists() bool
	GetFilepath() (dt.Filepath, error)
	GetRelFilepath() dt.RelFilepath
	SetRelFilepath(dt.RelFilepath)
	SetConfigDir(dt.DirPath)
	ConfigDir() (dt.DirPath, error)
	EnsureDirs(subdirs []dt.PathSegment) error
	WithDirType(DirType) ConfigStore
	DirType() DirType
	ConfigStore()
	ConfigSlug() dt.PathSegment
	IsNil() bool
}

var _ ConfigStore = (*configStore)(nil)

type configStore struct {
	configSlug dt.PathSegment
	// parentDir is <projectDir> ot Getwd() for ProjectConfig,
	// or ~/.config for CLIConfigStore
	// or UserConfigDir() for StdConfig
	configDir    dt.DirPath
	relFilepath  dt.RelFilepath
	dirType      DirType
	dirsProvider *DirsProvider
	fs           fs.FS
}

type ConfigStoreArgs struct {
	// ConfigSlug is the single path segment used for ~/.config/<slug>
	ConfigSlug dt.PathSegment

	// RelFilepath is the filename to be used for a file in the config directory which
	// may optionally include one or more parent paths but should not be an absolute
	// path.
	RelFilepath dt.RelFilepath

	// DirsProvider is typically never used for production code. It is intended only
	// to be used for test code in conjunction with go-the fsfix package
	DirsProvider *DirsProvider
}

func NewCLIConfigStore(configSlug dt.PathSegment, configFile dt.RelFilepath) ConfigStore {
	return NewConfigStore(CLIConfigDir, ConfigStoreArgs{
		ConfigSlug:  configSlug,
		RelFilepath: configFile,
	})
}

func NewProjectConfigStore(configSlug dt.PathSegment, configFile dt.RelFilepath) ConfigStore {
	return NewConfigStore(ProjectConfigDir, ConfigStoreArgs{
		ConfigSlug:  configSlug,
		RelFilepath: configFile,
	})
}

func NewConfigStore(dirType DirType, args ConfigStoreArgs) ConfigStore {
	if dirType == UnspecifiedConfigDir {
		panic("NewConfigStore: DirType is Unspecified")
	}
	if args.DirsProvider == nil {
		args.DirsProvider = &DirsProvider{
			UserHomeDirFunc:   dt.UserHomeDir,
			UserConfigDirFunc: dt.UserConfigDir,
			GetwdFunc:         dt.Getwd,
			ProjectDirFunc: func() (dt.DirPath, error) {
				return dt.Getwd()
			},
		}
	}
	return &configStore{
		dirType:      dirType,
		configSlug:   args.ConfigSlug,
		relFilepath:  args.RelFilepath,
		dirsProvider: args.DirsProvider,
	}
}

func RootRelative(p string) (string, error) {

	vol := filepath.VolumeName(p) // "C:" or "\\server\\share"
	if vol == "" {                // not a rooted Windows path
		return p, nil
	}
	base := vol + string(os.PathSeparator) // "C:\" or "\\server\\share\"
	rel, err := filepath.Rel(base, p)      // strip the volume root
	if err != nil {
		return "", err
	}
	if rel == "." { // exactly the root
		return "", nil
	}
	return rel, nil
}

func (cs *configStore) ConfigDir() (dir dt.DirPath, err error) {
	if cs.configDir != "" {
		goto end
	}
	{
		dp := cs.dirsProvider
		switch cs.dirType {
		case CLIConfigDir:
			dir, err = dp.UserHomeDirFunc()
			if err != nil {
				err = NewErr(ErrFailedGettingUserHomeDir, err)
				goto end
			}
			cs.configDir = dt.DirPathJoin3(dir, DotConfigPathSegment, cs.configSlug)

		case ProjectConfigDir:
			dir, err = dp.ProjectDirFunc()
			if err != nil {
				err = NewErr(ErrFailedGettingWorkingDir, err)
				goto end
			}
			cs.configDir = dt.DirPathJoin(dir, "."+cs.configSlug)

		case AppConfigDir:
			dir, err = dp.UserConfigDirFunc()
			if err != nil {
				err = NewErr(ErrFailedGettingUserConfigDir, err)
				goto end
			}
			cs.configDir = dt.DirPathJoin(dir, cs.configSlug)

		case UnspecifiedConfigDir:
			err = NewErr(ErrConfigDirTypeNotSet)
			goto end
		default:
			err = NewErr(
				ErrInvalidConfigDirType,
				"config_dir_type", cs.dirType,
			)
			goto end
		}
	}
end:
	return cs.configDir, err
}

func (cs *configStore) ConfigStore() {}

func (cs *configStore) ConfigPath() dt.PathSegment {
	return cs.configSlug
}

func (cs *configStore) SetRelFilepath(rf dt.RelFilepath) {
	cs.relFilepath = rf
}

func (cs *configStore) GetRelFilepath() dt.RelFilepath {
	return cs.relFilepath
}

func (cs *configStore) GetFilepath() (fp dt.Filepath, err error) {
	var dir dt.DirPath

	dir, err = cs.ConfigDir()
	if err != nil {
		goto end
	}

	if !cs.relFilepath.ValidPath() {
		err = NewErr(
			de.ErrInvalid,
			dt.ErrInvalidForOpen,
			"filepath", cs.relFilepath,
		)
		goto end
	}

	fp = dt.FilepathJoin(dir, cs.relFilepath)

end:
	return fp, err
}

func (cs *configStore) Save(data []byte) (err error) {
	var file *os.File
	var fullPath dt.Filepath

	fullPath, err = cs.ensureFilepath()
	if err != nil {
		goto end
	}

	file, err = dt.CreateFile(fullPath)
	if err != nil {
		goto end
	}
	defer CloseOrLog(file)

	_, err = file.Write(data)

end:
	return err
}

func (cs *configStore) SaveJSON(data any) (err error) {
	var jsonData []byte

	// Use JSON v2 with pretty printing via jsontext.WithIndent
	jsonData, err = jsonv2.Marshal(data, jsontext.WithIndent("  "))
	if err != nil {
		goto end
	}

	err = cs.Save(jsonData)

end:
	return err
}

func (cs *configStore) Load() (data []byte, err error) {
	var fSys fs.FS

	fSys, err = cs.getFS()
	if err != nil {
		err = WithErr(ErrFailedToGetConfigFileSystem, err)
		goto end
	}

	data, err = cs.relFilepath.ReadFile(fSys)
	if NoSuchFileOrDirectory(err) {
		err = NewErr(ErrFileDoesNotExist, err)
	}
	if err != nil {
		err = NewErr(ErrFailedToReadFile, err)
		goto end
	}

end:
	return data, err
}

func (cs *configStore) LoadJSON(data any, opts ...jsonv2.Options) (err error) {
	var jsonData []byte
	jsonData, err = cs.Load()
	if err != nil {
		err = NewErr(ErrFailedToReadConfigFile, err)
		goto end
	}

	// Use JSON v2 with any provided options (including custom unmarshalers)
	err = jsonv2.Unmarshal(jsonData, data, opts...)
	if err != nil {
		err = NewErr(ErrFailedToUnmarshalConfigFile, err)
		goto end
	}

end:
	if err != nil {
		err = WithErr(err, ErrFailedToLoadJSON)
	}
	return err
}

func (cs *configStore) Exists() (exists bool) {
	fSys, err := cs.getFS()
	if err != nil {
		goto end
	}
	_, err = cs.relFilepath.Stat(fSys)
	exists = err == nil

end:
	return exists
}

// SetConfigDir allows overriding config dir for unit testing.
func (cs *configStore) SetConfigDir(dir dt.DirPath) {
	cs.configDir = dir
	cs.fs = dt.DirFS(dir)
}

// EnsureDirs creates the specified subdirectories under this ConfigStore's config directory.
// This is a convenience method that delegates to EnsureConfigDirs function.
func (cs *configStore) EnsureDirs(subdirs []dt.PathSegment) (err error) {
	var configDir dt.DirPath

	configDir, err = cs.ConfigDir()
	if err != nil {
		goto end
	}
	err = EnsureConfigDirs(configDir, subdirs)

end:
	return err
}

func (cs *configStore) WithDirType(dt DirType) ConfigStore {
	store := *cs
	store.dirType = dt
	return &store
}

func (cs *configStore) DirType() DirType {
	return cs.dirType
}

func (cs *configStore) ConfigSlug() dt.PathSegment {
	return cs.configSlug
}

func (cs *configStore) IsNil() bool {
	return cs == nil
}

func (cs *configStore) ensureConfig(rc RootConfig, opts Options) (err error) {
	err = cs.loadConfigIfExists(rc, opts)
	if err != nil {
		// A real error occurred, bail out
		goto end
	}

	if rc == nil || dtx.IsZero(rc) {
		// Config not loaded, need to create config
		err = cs.createConfig(rc, opts)
		goto end
	}

end:
	return err
}

func (cs *configStore) createConfig(rc RootConfig, opts Options) (err error) {
	var fp dt.Filepath
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.Normalize(fp, opts)
	if err != nil {
		goto end
	}
	err = cs.SaveJSON(rc)
	if err != nil {
		goto end
	}
end:
	return err
}

func (cs *configStore) loadConfigIfExists(rc RootConfig, opts Options) (err error) {
	var fp dt.Filepath
	if !cs.Exists() {
		goto end
	}

	err = cs.LoadJSON(rc)
	if err != nil {
		goto end
	}
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.Normalize(fp, opts)
	if err != nil {
		goto end
	}
end:
	if err != nil {
		err = WithErr(err,
			"config_file", fp,
		)
	}
	return err
}

func (cs *configStore) getFS() (_ fs.FS, err error) {
	var dir dt.DirPath

	if cs.fs != nil {
		goto end
	}

	dir, err = cs.ConfigDir()
	if err != nil {
		goto end
	}

	cs.fs = dt.DirFS(dir)

end:
	return cs.fs, err
}

func (cs *configStore) ensureFilepath() (fp dt.Filepath, err error) {
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	// This is needed in case filepath contains a subdirectory, e.g. tokens/token-bill@microsoft.com.json
	err = fp.Dir().MkdirAll(0755)
	if err != nil {
		goto end
	}
end:
	return fp, err
}
