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
	ErrConfigDirTypeNotSet        = errors.New("config dir type not set")
	ErrInvalidConfigDirType       = errors.New("invalid config dir type")
	ErrFailedGettingWorkingDir    = errors.New("failed to get working dir")
	ErrFailedGettingUserConfigDir = errors.New("failed to get user config dir")
	ErrFailedGettingUserHomeDir   = errors.New("failed to get user home dir")
)

var ErrFailedToEnsureConfig = errors.New("failed to ensure config")
var ErrFailedToLoadConfig = errors.New("failed to load config")
var ErrFailedToLoadJSON = errors.New("failed to load JSON")

var (
	ErrUsernameRequired = errors.New("username required")
	ErrInvalidUsername  = errors.New("invalid username")
)

var ErrInvalidConfigFilepath = errors.New("invalid config filepath")
