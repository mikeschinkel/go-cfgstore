package cfgstore

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"fmt"
	"io/fs"
	"os"

	"github.com/mikeschinkel/go-dt"
)

// DefaultConfigDirType is currently hardcoded for ~/.config but having this
// const will make it easy to track down how where to change it if we want to make it
// configurable.
const DefaultConfigDirType = DotConfigDir

const ConfigBaseDirName = ".config"

// ConfigStore provides file operations for Gmail APIConfig
type ConfigStore interface {
	Load() ([]byte, error)
	Save([]byte) error
	LoadJSON(data any, opts ...jsonv2.Options) error
	SaveJSON(data any) error
	Exists() bool
	GetFilepath() (dt.Filepath, error)
	SetRelFilepath(dt.RelFilepath)
	SetConfigDir(dt.DirPath)
	ConfigDir() (dt.DirPath, error)
	WithDirType(DirType) ConfigStore
	ConfigStore()
}

var _ ConfigStore = (*configStore)(nil)

type configStore struct {
	appConfigSubdir dt.PathSegments
	configDir       dt.DirPath
	relFilepath     dt.RelFilepath
	dirType         DirType
	fs              fs.FS
}

func (cs *configStore) ConfigStore() {}

func (cs *configStore) ensureConfig(rc RootConfig, opts Options) (err error) {
	err = cs.loadConfigIfExists(rc, opts)
	if err != nil {
		// A real error occurred, bail out
		goto end
	}

	if rc == nil {
		// Config not loaded, need to create config
		err = cs.createConfig(rc, opts)
		goto end
	}

end:
	return err
}

var fp dt.Filepath

func (cs *configStore) createConfig(rc RootConfig, opts Options) (err error) {
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	rc.Normalize(fp, opts)
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

	err = cs.LoadJSON(rc, nil)
	if err != nil {
		goto end
	}
	fp, err = cs.GetFilepath()
	if err != nil {
		goto end
	}
	rc.Normalize(fp, opts)
end:
	return err
}

func (cs *configStore) Create(args ConfigStoreArgs) ConfigStore {
	ncs := NewConfigStore(args)
	ncs.SetRelFilepath(args.ConfigFile())
	return ncs
}

func NewConfigStore(args ConfigStoreArgs) ConfigStore {
	return &configStore{
		appConfigSubdir: args.ConfigDir(),
		relFilepath:     args.ConfigFile(),
		dirType:         DotConfigDir,
	}
}

func (cs *configStore) ConfigDir() (dir dt.DirPath, err error) {
	if cs.configDir != "" {
		goto end
	}
	switch cs.dirType {
	case DotConfigDir:
		dir, err = dt.UserHomeDir()
		if err != nil {
			goto end
		}
		cs.configDir = dt.DirPathJoin3(dir, ConfigBaseDirName, cs.appConfigSubdir)
	case LocalConfigDir:
		dir, err = dt.Getwd()
		if err != nil {
			goto end
		}
		cs.configDir = dt.DirPathJoin(dir, "."+cs.appConfigSubdir)
	case GoUserConfigDir:
		dir, err = dt.UserConfigDir()
		if err != nil {
			goto end
		}
		cs.configDir = dt.DirPathJoin(dir, cs.appConfigSubdir)
	case UnspecifiedConfigDir:
		err = fmt.Errorf("config dir type not set")
	default:
		err = fmt.Errorf("invalid config dir type: %d", cs.dirType)
	}

end:
	return cs.configDir, err
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
	// This is needed in case filepath contains a subdirectory, e.g. tokens/token-bill@microsoft.com.json
	err = dt.MkdirAll(dt.Dir(fp), 0755)
	if err != nil {
		goto end
	}
end:
	return fp, err
}

func (cs *configStore) SetRelFilepath(rf dt.RelFilepath) {
	cs.relFilepath = rf
}

func (cs *configStore) GetFilepath() (fp dt.Filepath, err error) {
	var dir dt.DirPath

	dir, err = cs.ConfigDir()
	if err != nil {
		goto end
	}

	if !dt.ValidRelPath(cs.relFilepath) {
		err = fmt.Errorf("path %s is not valid for use in %s", cs.relFilepath, dir)
		goto end
	}

	fp = dt.FilepathJoin(cs.configDir, cs.relFilepath)

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

	data, err = dt.FSReadFile(fSys, dt.Filepath(cs.relFilepath))
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
		err = WithErr(ErrFailedToReadConfigFile, err)
		goto end
	}

	// Use JSON v2 with any provided options (including custom unmarshalers)
	err = jsonv2.Unmarshal(jsonData, data, opts...)
	if err != nil {
		err = NewErr(ErrFailedToUnmarshalConfigFile, err)
		goto end
	}

end:
	return err
}

func (cs *configStore) Exists() (exists bool) {
	fSys, err := cs.getFS()
	if err != nil {
		goto end
	}
	_, err = dt.FSStatFile(fSys, dt.Filepath(cs.relFilepath))
	exists = err == nil

end:
	return exists
}

// SetConfigDir allows overriding config dir for unit testing.
func (cs *configStore) SetConfigDir(dir dt.DirPath) {
	cs.configDir = dir
	cs.fs = dt.DirFS(dir)
}

func (cs *configStore) WithDirType(dt DirType) ConfigStore {
	store := *cs
	store.dirType = dt
	return &store
}
