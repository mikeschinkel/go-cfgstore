package cfgstore

import (
	"errors"
)

var (
	ErrFailedToGetConfigFileSystem = errors.New("failed to get config file system")
	ErrFailedToReadFile            = errors.New("failed to read file")
	ErrFailedToReadConfigFile      = errors.New("failed to read config file")
	ErrFailedToUnmarshalConfigFile = errors.New("failed to unmarshal config file")
	ErrFileDoesNotExist            = errors.New("file does not exist")
)
var (
	ErrConfigDirTypeNotSet  = errors.New("config dir type not set")
	ErrInvalidConfigDirType = errors.New("invalid config dir type")
)

// TODO: Please these with dt.ErrAccessing*Dir
var (
	ErrFailedGettingWorkingDir    = errors.New("failed to get working dir")
	ErrFailedGettingUserConfigDir = errors.New("failed to get user config dir")
	ErrFailedGettingCLIConfigDir  = errors.New("failed to get CLI config dir")
	ErrFailedGettingUserHomeDir   = errors.New("failed to get user home dir")
	ErrFailedGettingUserCacheDir  = errors.New("failed to get user cache dir")
)

var ErrFailedToEnsureConfig = errors.New("failed to ensure config")
var ErrFailedToLoadConfig = errors.New("failed to load config")
var ErrFailedToLoadJSON = errors.New("failed to load JSON")
var ErrConfigAlreadyExists = errors.New("config already exists")

var (
	ErrUsernameRequired = errors.New("username required")
	ErrInvalidUsername  = errors.New("invalid username")
)

var ErrInvalidConfigFilepath = errors.New("invalid config filepath")

var ErrNoRootConfigsLoaded = errors.New("no root configs loaded")
