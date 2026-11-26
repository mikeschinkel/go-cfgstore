package cfgstore

import (
	"io"
	"os"
	"path/filepath"

	"github.com/mikeschinkel/go-dt/dtx"
)

func ReadFileIfExists(file string) (bytes []byte, err error) {
	bytes, err = os.ReadFile(file)
	if dtx.NoFileOrDirErr(err) {
		err = nil
	}
	return bytes, err
}

func GetBaseFilename(fullPath string) string {
	filename := filepath.Base(fullPath)
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

func CloseOrLog(c io.Closer) {
	logger := Logger()
	defer func() {
		if err := recover(); err != nil {
			logger.Warn("Panicked on close", "error", err)
		}
	}()
	err := c.Close()
	if err != nil {
		logger.Warn("Failed to close", "error", err)
	}
}

func LogOnError(err error) {
	EnsureLogger()
	if err != nil {
		logger.Warn("Operation failed", "error", err)
	}
}
