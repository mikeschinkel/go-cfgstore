package cfgstore

import (
	"runtime"

	"github.com/mikeschinkel/go-dt"
)

// CacheOptions provides optional configuration for cache directory functions
type CacheOptions struct {
	DirsProvider *DirsProvider
}

// GetSharedCacheDir returns the shared cache directory for the given slug.
// Platform-specific paths:
//   - macOS: ~/Library/Caches/{slug}/
//   - Linux: ~/.cache/{slug}/
//   - Windows: %LOCALAPPDATA%\{slug}\
//
// Example: GetSharedCacheDir("xmlui") → ~/.cache/xmlui/ on Linux
func GetSharedCacheDir(slug dt.PathSegment, opts ...CacheOptions) (dt.DirPath, error) {
	return getCacheDir(slug, "", opts...)
}

// GetAppCacheDir returns an app-specific cache directory under the shared cache.
// Platform-specific paths:
//   - macOS: ~/Library/Caches/{slug}/{appName}/
//   - Linux: ~/.cache/{slug}/{appName}/
//   - Windows: %LOCALAPPDATA%\{slug}\{appName}\
//
// Example: GetAppCacheDir("xmlui", "cli") → ~/.cache/xmlui/cli/ on Linux
func GetAppCacheDir(slug, appName dt.PathSegment, opts ...CacheOptions) (dt.DirPath, error) {
	return getCacheDir(slug, appName, opts...)
}

// getCacheDir is the internal implementation for cache directory resolution
func getCacheDir(slug, appName dt.PathSegment, opts ...CacheOptions) (dt.DirPath, error) {
	var dp *DirsProvider
	if len(opts) > 0 && opts[0].DirsProvider != nil {
		dp = opts[0].DirsProvider
	} else {
		dp = &DirsProvider{
			UserCacheDirFunc: dt.UserCacheDir,
		}
	}

	cacheDir, err := dp.UserCacheDirFunc()
	if err != nil {
		return "", NewErr(ErrFailedGettingUserCacheDir, err)
	}

	// Platform-specific path construction
	var cachePath dt.DirPath
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Caches/{slug}[/{appName}]
		if appName != "" {
			cachePath = dt.DirPathJoin3(cacheDir, slug, appName)
		} else {
			cachePath = dt.DirPathJoin(cacheDir, slug)
		}
	case "windows":
		// Windows: %LOCALAPPDATA%\{slug}[{\appName}]
		if appName != "" {
			cachePath = dt.DirPathJoin3(cacheDir, slug, appName)
		} else {
			cachePath = dt.DirPathJoin(cacheDir, slug)
		}
	default:
		// Linux and others: ~/.cache/{slug}[/{appName}]
		if appName != "" {
			cachePath = dt.DirPathJoin3(cacheDir, slug, appName)
		} else {
			cachePath = dt.DirPathJoin(cacheDir, slug)
		}
	}

	return cachePath, nil
}
