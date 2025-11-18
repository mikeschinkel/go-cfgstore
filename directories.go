package cfgstore

import (
	"github.com/mikeschinkel/go-dt"
	"github.com/mikeschinkel/go-dt/de"
)

// EnsureConfigDirs creates the specified subdirectories under the given config directory.
// This is a generic utility that can be used by any app to create its required structure.
//
// Parameters:
//   - configDir: The base config directory (e.g., ~/.config/xmlui/)
//   - subdirs: Slice of path segments to create under configDir
//
// Example:
//
//	EnsureConfigDirs(configDir, []dt.PathSegment{"demos", "logs"})
//
// creates ~/.config/xmlui/demos/ and ~/.config/xmlui/logs/
func EnsureConfigDirs(configDir dt.DirPath, subdirs []dt.PathSegment) (err error) {
	var errs []error

	for _, dir := range subdirs {
		dirPath := dt.DirPathJoin(configDir, dir)
		err := dt.MkdirAll(dirPath, 0755)
		if err != nil {
			errs = append(errs, dt.NewErr(
				dt.ErrFailedCreatingDirectory,
				err,
				"dir", dirPath,
			))
		}
	}
	err = dt.CombineErrs(errs)

	return err
}
