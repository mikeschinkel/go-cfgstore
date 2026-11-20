package cfgstore

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"io/fs"
	"os"
	"runtime"

	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/dtx"
)

// DefaultConfigDirType is currently hardcoded for ~/.config but having this
// const will make it easy to track down how where to change it if we want to make it
// configurable.
const DefaultConfigDirType = CLIConfigDirType

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
	return NewConfigStore(CLIConfigDirType, ConfigStoreArgs{
		ConfigSlug:  configSlug,
		RelFilepath: configFile,
	})
}

func NewProjectConfigStore(configSlug dt.PathSegment, configFile dt.RelFilepath) ConfigStore {
	return NewConfigStore(ProjectConfigDirType, ConfigStoreArgs{
		ConfigSlug:  configSlug,
		RelFilepath: configFile,
	})
}

// ProjectConfigStoreExists takes a project directory and reports whether its
// project config, <projectDir>/.<configSlug>/<configFile>, exists.
func ProjectConfigStoreExists(projectDir dt.DirPath, configSlug dt.PathSegment, configFile dt.RelFilepath) (exists bool) {

	dp := DefaultDirsProvider()
	dp.ProjectDirFunc = func() (dt.DirPath, error) {
		return projectDir, nil
	}
	store := NewConfigStore(ProjectConfigDirType, ConfigStoreArgs{
		ConfigSlug:   configSlug,
		RelFilepath:  configFile,
		DirsProvider: dp,
	})
	exists = store.Exists()

	return exists
}

// InitProjectConfig creates a project config file and returns the config a
// load of that file alone would return. The file holds no defaults, only what
// Normalize sets on an empty value: a project file holding every default would
// override each user's CLI-layer settings. The returned config has
// ApplyDefaults applied, then the file merged and normalized over it.
// It returns ErrConfigAlreadyExists if the file already exists.
func InitProjectConfig[RC any, PRC RootConfigPtr[RC]](
	configSlug dt.PathSegment,
	configFile dt.RelFilepath,
	opts Options,
) (prc PRC, err error) {
	var fp dt.Filepath
	var data []byte

	dirTypes := []DirType{ProjectConfigDirType}
	store := NewProjectConfigStore(configSlug, configFile)
	saved := makeRootConfig[RC, PRC]()

	if store.Exists() {
		err = ErrConfigAlreadyExists
		goto end
	}
	fp, err = store.GetFilepath()
	if err != nil {
		goto end
	}
	err = saved.Normalize(NormalizeArgs{
		DirType:    ProjectConfigDirType,
		SourceFile: fp,
		Options:    opts,
	})
	if err != nil {
		err = NewErr(ErrFailedToNormalizeConfig, err)
		goto end
	}
	// A zero field is omitted whatever its omitempty tag says: written as 0,
	// false or "", it would override the CLI layer's value for it.
	data, err = jsonv2.Marshal(saved, jsontext.WithIndent("  "), jsonv2.OmitZeroStructFields(true))
	if err != nil {
		goto end
	}
	err = store.Save(data)
	if err != nil {
		goto end
	}
	prc, err = layerConfigs[RC, PRC]([]configLayer[PRC]{{
		dirType: ProjectConfigDirType,
		fp:      fp,
		data:    data,
		value:   saved,
	}}, RootConfigArgs{
		DirTypes: dirTypes,
		Options:  opts,
	})
end:
	return prc, err
}

type DirsProviderArgs struct {
	CustomDirPath dt.DirPath
}

func DefaultDirsProviderWithArgs(args DirsProviderArgs) (dp *DirsProvider) {
	dp = DefaultDirsProvider()
	dp.CustomConfigDirFunc = func() (dt.DirPath, error) {
		return args.CustomDirPath, nil
	}
	return dp
}
func DefaultDirsProvider() *DirsProvider {
	dp := &DirsProvider{
		UserHomeDirFunc:   dt.UserHomeDir,
		UserConfigDirFunc: dt.UserConfigDir,
		GetwdFunc:         dt.Getwd,
		ProjectDirFunc: func() (dt.DirPath, error) {
			return dt.Getwd()
		},
	}
	dp.CLIConfigDirFunc = dp.CLIConfigDirType
	return dp
}
func NewConfigStore(dirType DirType, args ConfigStoreArgs) ConfigStore {
	if dirType == UnspecifiedConfigDirType {
		panic("NewConfigStore: DirType is Unspecified")
	}
	if args.DirsProvider == nil {
		args.DirsProvider = DefaultDirsProvider()
	}
	return &configStore{
		dirType:      dirType,
		configSlug:   args.ConfigSlug,
		relFilepath:  args.RelFilepath,
		dirsProvider: args.DirsProvider,
	}
}

// CLIConfigDirType returns the parent of a CLI app's config dir: the user config
// dir on Linux ($XDG_CONFIG_HOME, or ~/.config), and ~/.config on every other OS.
func (dp *DirsProvider) CLIConfigDirType() (dir dt.DirPath, err error) {
	switch runtime.GOOS {
	case "linux":
		// The user config dir is already $XDG_CONFIG_HOME or ~/.config.
		dir, err = dp.UserConfigDirFunc()
		if err != nil {
			err = NewErr(ErrFailedGettingUserConfigDir, err)
		}
	default:
		// macOS and Windows put CLI config in ~/.config, not the OS's app config dir.
		dir, err = dp.UserHomeDirFunc()
		if err != nil {
			err = NewErr(ErrFailedGettingUserHomeDir, err)
			goto end
		}
		dir = dt.DirPathJoin(dir, DotConfigPathSegment)
	}
end:
	return dir, err
}

func (cs *configStore) ConfigDir() (dir dt.DirPath, err error) {
	if cs.configDir != "" {
		goto end
	}
	cs.configDir, err = ConfigDir(cs.dirType, cs.configSlug, cs.dirsProvider)
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
			dt.ErrInvalid,
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
	if dtx.NoFileOrDirErr(err) {
		err = NewErr(ErrFileNotExist, err)
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

// createConfig creates a missing config file holding rc's defaults, run
// through Normalize.
func (cs *configStore) createConfig(rc RootConfig, dirType DirType, opts Options, dirTypes []DirType) (err error) {
	var fp dt.Filepath

	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	err = rc.ApplyDefaults(DefaultsArgs{
		Options:  opts,
		DirTypes: dirTypes,
	})
	if err != nil {
		err = NewErr(ErrFailedToApplyDefaults, err)
		goto end
	}
	err = rc.Normalize(NormalizeArgs{
		DirType:    dirType,
		SourceFile: fp,
		Options:    opts,
	})
	if err != nil {
		err = NewErr(ErrFailedToNormalizeConfig, err)
		goto end
	}
	err = cs.SaveJSON(rc)
end:
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
